#!/usr/bin/env python

###############################################################################
#
# Summit Route End Point Protection
#
# This source code is licensed under the BSD-style license found in the
# LICENSE file in the root directory of this source tree.
#
###############################################################################


import time
import signal
import sys
import json
import os
import re

import pprint
import traceback

# For version checking
from distutils.version import StrictVersion

# Logging
import logging
from logging.handlers import RotatingFileHandler

# Parse arguments
import argparse

# Queue libraries
import pika

# DB libraries
import sqlalchemy
from sqlalchemy.orm import sessionmaker
from model import *

# PE Parsing libraries
from pefile.pefile import PE, DIRECTORY_ENTRY

# Signature verification libraries
from pyasn1.codec.der import encoder as der_encoder
import verifysigs.auth_data as auth_data
import verifysigs.fingerprint as fingerprint
import verifysigs.pecoff_blob as pecoff_blob
from verifysigs.asn1 import dn

import oid

# Catalog files
from pyasn1.type import univ
from pyasn1.codec.ber import encoder, decoder
from verifysigs.asn1 import pkcs7, catalog

# Amazon S3
from boto.s3.connection import S3Connection
from boto.s3.key import Key as S3Key
import tempfile

# Globals
CONFIG_FILE_PATH = './config.json'
Session = None
config = None
logger = None


class FileTask(object):
    fileID = 0

    def __init__(self, str):
        self.__dict__ = json.loads(str)


class CatalogTask(object):
    catalogID = 0

    def __init__(self, str):
        self.__dict__ = json.loads(str)


class Config(object):

    def __init__(self, file_object):
        self.__dict__ = json.load(file_object)


################################################################################
# stringToHexString takes a string (received from the DB) that is basically a
# byte array, and converts it to a hex string
def stringToHexString(str):
    return "".join("%02x" % ord(b) for b in str)


################################################################################
# getFileHashPath: Given a sha256 hex string, provides a path to the uploaded
# file
def getFileHashPath(sha256):
    return os.path.join(sha256[:2], sha256[2:4], sha256)


################################################################################
# checkLibraries checks the version of libraries to ensure we aren't using
# anything that's been untested
def checkLibraries():
    # Check library versions.  Debian installs an old version of pika.
    if StrictVersion(pika.__version__) < StrictVersion('0.9.14'):
        logger.error("Your version of Pika (%s) is too old.  Install with pip, not apt-get" % pika.__version__)
        logger.error("Pika version should be 0.9.14 or newer")
        sys.exit(-1)

    if not (sys.version_info.major == 2 and sys.version_info.minor == 7):
        logger.error("Python version is not 2.7")
        sys.exit(-1)


################################################################################
# parseFile extracts the resource info that gives the CompanyName and other info
def parseFile(fileID, filePath):
    pe_info = {
        "productname": "",
        "companyname": "",
        "filedescription": "",
        "productversion": "",
        "internalname": "",
        "fileversion": "",
        "originalfilename": ""
        }

    try:
        pe = PE(filePath, fast_load=True)
        pe.parse_data_directories(
            directories=[DIRECTORY_ENTRY['IMAGE_DIRECTORY_ENTRY_RESOURCE']])
    except Exception as e:
        logger.error("Exception parsing file %d" % fileID)
        return pe_info

    try:
        for fileinfo in pe.FileInfo:
            if fileinfo.Key == 'StringFileInfo':
                for st in fileinfo.StringTable:
                    for entry in st.entries.items():
                        if entry[0].lower() in pe_info:
                            value = entry[1]
                            value = re.sub(r'[^\x00-\x7F]', '', value)
                            pe_info[entry[0].lower()] = value
    except Exception as e:
        logger.error("Exception extracting parsed data from file %d; %s" % (fileID, e))

    if pe.FILE_HEADER.Machine == 0x14c:
        pe_info["architecture"] = 32
    elif pe.FILE_HEADER.Machine == 0x8664:
        pe_info["architecture"] = 64

    return pe_info


################################################################################
# recordAuthenticodeHash records info in the DB for an authenticode hash and
#   looks for catalogs it matches
def recordAuthenticodeHash(session, exefile, hashType, hashValue):
    if 'md5' == hashType:
        exefile.authenticodemd5 = hashValue
    elif 'sha1' in hashType:
        exefile.authenticodesha1 = hashValue
    elif 'sha256' in hashType:
        exefile.authenticodesha256 = hashValue
    else:
        logger.warn("Unknown hash type '%s' for file %d" % (hashType, exefile.id))
        return

    session.commit()

    ctl_entries = session.query(CertificateTrustList) \
        .filter(CertificateTrustList.hash == hashValue) \
        .filter(CertificateTrustList.hashtype == hashType) \
        .all()
    for ctl in ctl_entries:
        # Match found, so update the DB

        # Set the fileID for this CTL entry
        ctl.fileid = exefile.id
        session.commit()

        # Set the FileToSignerMap, first we need the signerID
        catalog = session.query(CatalogFile) \
            .filter(CatalogFile.id == ctl.catalogid).one()
        signerID = catalog.signerid

        try:
            fts = FileToSignerMap(fileid=exefile.id, signerid=signerID)
            session.add(fts)
            session.commit()
        except sqlalchemy.exc.IntegrityError:
            # Ignore. This should only happen when we are re-anayzing the same files.
            session.rollback()

        logger.info("Match found in catalog file %d for exe %d" % (catalog.id, exefile.id))
    return


################################################################################
# recordAuthenticodeHashes calls recordAuthenticodeHash for each hash type
def recordAuthenticodeHashes(fileID, authenticode_hashes):
    # Find the executable file in the DB
    session = Session()
    try:
        exefile = session.query(ExecutableFiles) \
            .filter(ExecutableFiles.id == fileID).one()
    except sqlalchemy.orm.exc.NoResultFound:
        raise Exception("File ID not found in database")

    for hashType, hashValue in authenticode_hashes.iteritems():
        recordAuthenticodeHash(session, exefile, hashType, hashValue)


################################################################################
# getSignatures looks for signature info
def getSignatures(fileID, filePath):
    # Read in the file and get initial results
    with file(filePath, 'rb') as objf:
        fingerprinter = fingerprint.Fingerprinter(objf)
        is_pecoff = fingerprinter.EvalPecoff()
        fingerprinter.EvalGeneric()
        results = fingerprinter.HashIt()

    # File hashes
    # TODO SHOULD: double check the hashes that the agent provided
    # hashes = [x for x in results if x['name'] == 'generic']
    # if len(hashes) > 1:
    #     logger.warn("More than one generic fingerprint? Only printing first one.")
    # for hname in sorted(hashes[0].keys()):
    #     if hname != 'name':
    #         print('%s: %s' % (hname, hashes[0][hname].encode('hex')))

    if not is_pecoff:
        logger.error("File %d is not a PE file" % fileID)
        return None

    # Authenticode hashes
    hashes = [x for x in results if x['name'] == 'pecoff']
    authenticode_hashes = {}
    if len(hashes) > 1:
        logger.warn('More than one PE/COFF finger. Only using first one.')
    for hname in sorted(hashes[0].keys()):
        if hname != 'name' and hname != 'SignedData':
            hashValue = hashes[0][hname]
            hashType = hname
            if hashType in ('md5', 'sha1', 'sha256'):
                authenticode_hashes[hashType] = hashValue

    recordAuthenticodeHashes(fileID, authenticode_hashes)

    signed_pecoffs = [x for x in results if x['name'] == 'pecoff' and
                      'SignedData' in x]

    if not signed_pecoffs:
        logger.info("File %d has no signature" % fileID)
        return None

    signed_pecoff = signed_pecoffs[0]

    signed_datas = signed_pecoff['SignedData']
    # There may be multiple of these, if the windows binary was signed multiple
    # times, e.g. by different entities. Each of them adds a complete SignedData
    # blob to the binary.
    # TODO MUST (STP): Process all instances
    signed_data = signed_datas[0]

    blob = pecoff_blob.PecoffBlob(signed_data)

    # TODO SHOULD: This calls into openssl via the M2Crypto library. I should
    # avoid using that sketchy library.  Or ensure this worker is well isolated
    # and secured.
    auth = auth_data.AuthData(blob.getCertificateBlob())
    content_hasher_name = auth.digest_algorithm().name
    computed_content_hash = signed_pecoff[content_hasher_name]

    try:
        auth.ValidateAsn1()
        auth.ValidateHashes(computed_content_hash)
        auth.ValidateSignatures()
        auth.ValidateCertChains(time.gmtime())
    except auth_data.Asn1Error:
        if auth.openssl_error:
            logger.error('OpenSSL Errors for %d:\n%s' % (fileID, auth.openssl_error))
            return None

    # TODO MUST Collect all this data into a proper structure
    # logger.info('Program: %s, URL: %s' % (auth.program_name, auth.program_url))
    counter_cert_info = None
    if auth.has_countersignature:
        # logger.info(
        #     "Countersignature is present. Timestamp: %s UTC" %
        #     time.asctime(time.gmtime(auth.counter_timestamp))
        # )
        counter_cert_info = getCertInfo(auth.counter_sig_info, auth.certificates)
        counter_cert_info['timestamp'] = auth.counter_timestamp
    else:
        logger.info("Countersignature is not present.")

    # NOTE: Need to use: dn.DistinguishedName.TraverseRdn(auth.signed_data['certificates'][2]['certificate']['tbsCertificate']['subject'][0])

    # The same info exists in auth.signing_cert_id[0]
    signer_info = getCertInfo(auth.signer_info, auth.certificates)
    # TODO MUST: Collect parent certs info and store in DB
    return {"signerInfo": signer_info, "counterSignerInfo": counter_cert_info}


def ParseCerts(certs):
  # TODO(user):
  # Parse them into a dict with serial, subject dn, issuer dn, lifetime,
  # algorithm, x509 version, extensions, ...
  res = dict()
  for cert in certs:
    res[ExtractIssuer(cert)] = cert
  return res

def ExtractIssuer(cert):
  issuer = cert[0][0]['issuer']
  subject = cert[0][0]['subject']
  serial_number = int(cert[0][0]['serialNumber'])

  issuer_dn = str(dn.DistinguishedName.TraverseRdn(issuer[0]))
  subject_dn = str(dn.DistinguishedName.TraverseRdn(subject[0]))

  return (subject_dn, issuer_dn, serial_number)

def getDNString(issuerDN_dict):
    # This should follow https://www.ietf.org/rfc/rfc2253.txt

    attributes = ["CN", "L", "ST", "O", "OU", "C", "STREET", "DC", "UID"]

    result = ""
    for a in attributes:
        if a in issuerDN_dict:
            if len(result) != 0:
                result += ","
            # TODO I should escape commas and other letters
            result += "%s=%s" % (a, issuerDN_dict[a])

    return result


def getBestDNName(issuerDN_dict):
    attributes = ["CN", "OU", "O"]
    for a in attributes:
        if a in issuerDN_dict:
            return issuerDN_dict[a]
    # No good name found so just return whatever is left
    return getDNString(issuerDN_dict)


def oidToStr(oid_value):
    if oid_value in oid.OID_LOOKUP:
        return oid.OID_LOOKUP[oid_value]
    logger.warning("Unknown OID: %s" % str(oid_value))
    return str(oid_value)


def getCertInfo(cert_info, certs):
    results = {
        "Version": "",
        "Subject": "",
        "SubjectShortName": "",
        "SerialNumber": "",
        "DigestAlgorithm": "",
        "DigestEncryptionAlgorithm": "",
        "DigestEncryptionAlgorithmKeySize": 0  # TODO Must find
    }

    issuer_and_serial = cert_info['issuerAndSerialNumber']
    issuer = issuer_and_serial['issuer']
    serial_number = int(issuer_and_serial['serialNumber'])
    issuerDN_dict = dn.DistinguishedName.TraverseRdn(issuer[0])
    subjectDN_dict = None

    # This is such a sloppy way of getting this info
    for (subject_dn, issuer, serial), cert in certs.items():
        if issuer == str(issuerDN_dict):
            subjectDN_dict = dn.DistinguishedName.TraverseRdn(cert[0][0]['subject'][0])
            break
    if subjectDN_dict is None:
        return None

    results["SerialNumber"] = '%x' % serial_number
    results["Subject"] = getDNString(subjectDN_dict)
    results["SubjectShortName"] = getBestDNName(subjectDN_dict)

    results["Version"] = cert_info["version"].prettyPrint()
    # TODO MUST Get key size (RSA 2048? 1024?)
    # TODO SHOULD Get info about the key constraints (uses), revocation list, validity period, etc.
    results["DigestAlgorithm"] = oidToStr(cert_info["digestAlgorithm"]['algorithm'])
    results["DigestEncryptionAlgorithm"] = oidToStr(cert_info['digestEncryptionAlgorithm']['algorithm'])

    return results


################################################################################
# copyS3FileToLocal
def copyS3FileToLocal(s3auth, s3path):
    conn = S3Connection(s3auth["access_key"], s3auth["secret_key"])

    bucket = conn.get_bucket(s3auth["bucket_name"])
    k = bucket.get_key(s3path, validate=True)

    f = tempfile.NamedTemporaryFile(prefix="fa-", dir=config.tmp_file_path, delete=False)
    filepath = os.path.abspath(f.name)
    k.get_contents_to_filename(filename=filepath)
    return filepath


################################################################################
# analyzeExecutable
def analyzeExecutable(fileID):
    logger.info("Processing executable file %d" % fileID)

    session = Session()
    try:
        file = session.query(ExecutableFiles) \
            .filter(ExecutableFiles.id == fileID).one()
    except sqlalchemy.orm.exc.NoResultFound:
        raise Exception("File ID not found in database")

    sha256 = stringToHexString(file.sha256)
    filehashpath = getFileHashPath(sha256)

    filepath = copyS3FileToLocal(config.aws["s3"]["exe"], filehashpath)

    statinfo = os.stat(filepath)
    if statinfo.st_size != file.size:
        raise Exception("File sizes did not match (%d != %d)" % (statinfo.st_size, file.size))

    pe_info = parseFile(fileID, filepath)
    pe_info['analysisdate'] = time.time()

    # TODO I should run some checks on this stuff
    session.query(ExecutableFiles) \
        .filter(ExecutableFiles.id == fileID).update(pe_info)
    session.commit()

    sig_data = getSignatures(fileID, filepath)

    if sig_data is not None:
        # Store signature info in DB
        if sig_data['signerInfo'] is not None:
            signerID = addSignerToDB(session, sig_data['signerInfo'])

            ftsMap = session.query(FileToSignerMap) \
                .filter(FileToSignerMap.fileid == fileID) \
                .filter(FileToSignerMap.signerid == signerID).first()
            if ftsMap is None:
                ftsMap = FileToSignerMap()
                ftsMap.fileid = fileID
                ftsMap.signerid = signerID
                session.add(ftsMap)

        if sig_data['counterSignerInfo'] is not None:
            counterSignerID = addSignerToDB(session, sig_data['counterSignerInfo'])

            ftsMap = session.query(FileToCounterSignerMap) \
                .filter(FileToCounterSignerMap.fileid == fileID) \
                .filter(FileToCounterSignerMap.signerid == counterSignerID).first()
            if ftsMap is None:
                ftsMap = FileToCounterSignerMap()
                ftsMap.fileid = fileID
                ftsMap.timestamp = sig_data['counterSignerInfo']['timestamp']
                ftsMap.signerid = counterSignerID
                session.add(ftsMap)

        session.commit()

    # TODO Need to ensure that during exceptions we clean up these files
    os.remove(filepath)
    return


################################################################################
# executableFileCallback is called whenever a new task is received to analyze
# a PE file
def executableFileCallback(ch, method, properties, body):
    fileID = -1
    try:
        filetask = FileTask(body)
        fileID = filetask.FileID
        analyzeExecutable(fileID)

    except Exception as e:
        logger.error("Exception processing file %d; %s - %s" % (fileID, e, traceback.format_exc()))

    # Exception was thrown so we can look into this file later, so we ack this
    # Otherwise workers would keep getting passed it.
    ch.basic_ack(delivery_tag=method.delivery_tag)


################################################################################
# catalogFileCallback
def catalogFileCallback(ch, method, properties, body):
    catalogID = -1
    try:
        catalogtask = CatalogTask(body)
        catalogID = catalogtask.CatalogID
        analyzeCatalog(catalogID)

    except Exception as e:
        logger.error("Exception processing catalog %d; %s - %s" % (fileID, e, traceback.format_exc()))

    # Exception was thrown so we can look into this file later, so we ack this
    # Otherwise workers would keep getting passed it.
    ch.basic_ack(delivery_tag=method.delivery_tag)

################################################################################
# addSignerToDB adds the dict data to the DB
def addSignerToDB(session, fileSignerData):
    # Sanity check
    if fileSignerData is None:
        return None

    # Check if signer is in the DB already
    signer = None

    signer = session.query(Signer) \
        .filter(Signer.serialnumber == fileSignerData['SerialNumber']) \
        .filter(Signer.subject == fileSignerData['Subject']).first()

    if signer is None:
        # Add it to the DB
        signer = Signer()
        signer.version = fileSignerData['Version']
        signer.subject = fileSignerData['Subject']
        signer.subjectshortname = fileSignerData['SubjectShortName']
        signer.serialnumber = fileSignerData['SerialNumber'] # TODO MUST This is being saved as \x3330 for \x30
        signer.digestalgorithm = fileSignerData['DigestAlgorithm']
        signer.digestencryptionalgorithm = fileSignerData['DigestEncryptionAlgorithm']
        signer.digestencryptionalgorithmkeysize = fileSignerData['DigestEncryptionAlgorithmKeySize']

        session.add(signer)
        session.commit()
    return signer.id

################################################################################
# analyzeCatalog receives an int for the catalogID
def analyzeCatalog(catalogID):
    logger.info("Analyzing catalog %d", catalogID)
    # Catalogs are PKC7 data in DER format, so we can use openssl as:
    #  openssl pkcs7 -in d454c2c219cc40111d121a1c3ef76d3728000c67fa36f733363e872c48d8e2ba -inform DER -print_certs
    # and
    #  openssl asn1parse -in d454c2c219cc40111d121a1c3ef76d3728000c67fa36f733363e872c48d8e2ba -inform DER

    session = Session()
    try:
        catalogfile = session.query(CatalogFile) \
            .filter(CatalogFile.id == catalogID).one()
    except sqlalchemy.orm.exc.NoResultFound:
        raise Exception("Catalog ID not found in database")

    sha256 = stringToHexString(catalogfile.sha256)
    filehashpath = getFileHashPath(sha256)

    filepath = copyS3FileToLocal(config.aws["s3"]["catalog"], filehashpath)

    # Sanity check file is there
    statinfo = os.stat(filepath)
    if statinfo.st_size != catalogfile.size:
        raise Exception("File sizes did not match (%d != %d)" % (statinfo.st_size, catalogfile.size))

    with open(filepath, "rb") as f:
        contents = f.read()
        # Sanity check the whole file was read
        if len(contents) != catalogfile.size:
            raise Exception("File sizes of read data did not match (%d != %d)" % (len(contents), catalogfile.size))

        asn1Data, rest = decoder.decode(contents, asn1Spec=catalog.CatalogFile())
        if rest:
            logger.warn("Trailing data (%d bytes) in catalog %d" % (len(rest), catalogID))

        #
        # Get signer for this catalog
        #

        # TODO Should authenticate this like I do with PE files, and store the other certs and countersig
        certs = ParseCerts(asn1Data['PKCS7']['certificates'])
        signerID = addSignerToDB(session, getCertInfo(asn1Data['PKCS7']['signerInfos'][0], certs))
        logger.info("Signer ID: %d" % signerID)

        #
        # Get list of files this catalog says it trusts
        #

        if asn1Data['PKCS7']['contentInfo']['contentType'] != univ.ObjectIdentifier(oid.szOID_CTL):
            raise Exception("Catalog does not have a Certificate Trust List, content type was %s" % asn1Data['PKCS7']['contentInfo']['contentType'])
        CTL_bytes = asn1Data['PKCS7']['contentInfo']['content']

        ctl, rest = decoder.decode(CTL_bytes, asn1Spec=catalog.CertificateTrustList())
        if rest:
            logger.warn("Trailing data (%d bytes) in catalog %d" % (len(rest), catalogID))
        logger.info("Number of CTL members: %d", len(ctl['catalogList']))

        counter = 0
        for memberSet in ctl['catalogList']:
            # The memberSet has two objects:
            #   - 1.3.6.1.4.1.311.12.2.2 = CAT_MEMBERINFO_OBJID
            #      This just provides the CLSID: {C689AAB8-8E78-11D0-8C47-00C04FC295EE}
            #   - 1.3.6.1.4.1.311.2.1.4 = SPC_INDIRECT_DATA_OBJID
            #      This holds the data we want
            for member in memberSet['CatalogMemberSet']:
                if member['OID'] == univ.ObjectIdentifier(oid.SPC_INDIRECT_DATA_OBJID):
                    asn1Object, rest = decoder.decode(member['Object'], asn1Spec=catalog.DataObjectSet())
                    if rest:
                        logger.warn("Trailing data (%d bytes) in member %d in catalog %d" % (len(rest), counter, catalogID))
                    hashObject = asn1Object['DataObjectSequence']['HashObject']
                    hashType = "sha1"
                    if oidToStr(hashObject['OIDSequence']['OID']) == "sha1":
                        hashType = "sha1"
                    else:
                        logger.warn("Digest was not sha1 for member %d in catalog %d" % (counter, catalogID))
                        break
                    digest = hashObject['Hash'].asOctets()

                    # logger.info(stringToHexString(digest))

                    ctl_entry = None

                    try:
                        # Record this in the DB
                        ctl_entry = CertificateTrustList(
                            catalogid=catalogID,
                            hash=digest,
                            hashtype=hashType
                            )
                        session.add(ctl_entry)
                        session.commit()
                    except sqlalchemy.exc.IntegrityError:
                        # Ignore when we try to add the same data twice.
                        # This can happen because:
                        # - we are re-analyzing the same catalog
                        # - the catalog contains multiples of the same data
                        #  (yes this happens, such as the sha1 "1c83b2e2d28253403e3d45b59fe908b4fa730c5d"
                        #  appearing in C:\Windows\system32\CatRoot\{F750E6C3-38EE-11D1-85E5-00C04FC295EE}\ntexe.cat
                        #  twice)
                        session.rollback() # Must issue this to continue using the session

                        # Get the CTL ID for the CTL that already exists
                        # There should only be one, as we are filtering by the primary key
                        ctl_entry = session.query(CertificateTrustList) \
                            .filter(CertificateTrustList.hash == digest) \
                            .filter(CertificateTrustList.catalogid == catalogID).first()

                    # Look for matching executable files
                    exefiles = None
                    if hashType == 'md5':
                        exefiles = session.query(ExecutableFiles) \
                            .filter(ExecutableFiles.authenticodemd5 == digest).all()
                    elif hashType == 'sha1':
                        exefiles = session.query(ExecutableFiles) \
                            .filter(ExecutableFiles.authenticodesha1 == digest).all()
                    if hashType == 'sha256':
                        exefiles = session.query(ExecutableFiles) \
                            .filter(ExecutableFiles.authenticodesha256 == digest).all()

                    if exefiles is not None:
                        if len(exefiles) > 1:
                             logger.warn("More tha one match found for this hash, that should not happen")

                        for exefile in exefiles:
                            logger.info("Match found of catalog %d with file %d" % (catalogID, exefile.id))

                            # Save our match in the catalog
                            ctl_entry.fileid = exefile.id
                            session.commit()

                            # Set the FileToSignerMap
                            try:
                                fts = FileToSignerMap(fileid=exefile.id, signerid=signerID)
                                session.add(fts)
                                session.commit()
                            except sqlalchemy.exc.IntegrityError:
                                # Ignore. This should only happen when we are re-anayzing the same files.
                                session.rollback()

                    # TODO Need to do something for non-sha1
                    # TODO MUST For each hash added, need to check if a file matches this hash

                    counter+=1
        if counter != len(ctl['catalogList']):
            logger.warn("Only %d of %d certificate trust members were found in catalog %d" % (counter, len(ctl['catalogList']), catalogID))

        # Record that we analyzed this catalog
        catalogfile.analysisdate = time.time()
        catalogfile.signerid = signerID
        session.commit()

    # TODO Need to ensure that during exceptions we clean up these files
    os.remove(filepath)

    return


################################################################################
# initQueue initializes the queue
def initQueue(channel, callback, queueName):
    client_params = {"x-ha-policy": "all"}
    channel.queue_declare(
        queue=queueName,
        durable=True,
        auto_delete=False,
        arguments=client_params)
    channel.basic_consume(callback, queue=queueName)


################################################################################
# exit_gracefully is called when you hit Ctrl+C
def exit_gracefully(signum, frame):
    # From
    # http://stackoverflow.com/questions/18114560/python-catch-ctrl-c-command-prompt-really-want-to-quit-y-n-resume-executi
    signal.signal(signal.SIGINT, original_sigint)

    logger.info('Exiting (gracefully)')
    sys.exit(1)


################################################################################
# main function
if __name__ == '__main__':
    # Parse arguments
    parser = argparse.ArgumentParser(description='Worker for the task queues')
    parser.add_argument('--catalog', type=int, nargs='?',  help='Catalog file ID')
    parser.add_argument('--exe', type=int, nargs='?',  help='Executable file ID')
    args = parser.parse_args()

    # Set up signal handler
    original_sigint = signal.getsignal(signal.SIGINT)
    signal.signal(signal.SIGINT, exit_gracefully)

    # Load config
    config = Config(open(CONFIG_FILE_PATH))

    # Setup logger
    logger = logging.getLogger("fileTaskWorker")
    logger.setLevel(logging.INFO)

    # File log
    handler = RotatingFileHandler(config.logpath, maxBytes=1024*1024*20,
                                  backupCount=5)
    handler.setLevel(logging.INFO)
    logformatter = logging.Formatter('%(asctime)s [%(name)s:%(process)d %(levelname)s] %(message)s')
    handler.setFormatter(logformatter)
    logger.addHandler(handler)

    # Log to stdout
    consoleHandler = logging.StreamHandler()
    consoleHandler.setFormatter(logformatter)
    logger.addHandler(consoleHandler)

    # Ensure we are using libraries we've tested aganst
    checkLibraries()

    # Create temp dir
    try:
        os.makedirs(config.tmp_file_path)
    except:
        # An exception is thrown if the directory already exists, so ignore it
        pass

    #
    # Connect to DB
    #
    db = create_engine(config.db["connection_string"])
    db.echo = False

    Session = sessionmaker(bind=db)

    #
    # Check command-line arguments
    #
    if args.catalog is not None:
        analyzeCatalog(args.catalog)
        sys.exit(0)
    elif args.exe is not None:
        analyzeExecutable(args.exe)
        sys.exit(0)

    #
    # Connect to queue
    #
    credentials = pika.PlainCredentials(
        config.queue["username"], config.queue["password"])


    while True:
        connection = pika.BlockingConnection(pika.ConnectionParameters(
            host=config.queue["host"],
            credentials=credentials,
            heartbeat_interval=20))
        channel = connection.channel()
        channel.basic_qos(prefetch_count=1)

        initQueue(channel, executableFileCallback, "analyzefile")
        initQueue(channel, catalogFileCallback, "analyzecatalog")

        logger.info('Watching queue')

        try:
            channel.start_consuming()
        except pika.exceptions.ConnectionClosed:
            logger.info('Exception the connection was closed, reconnecting')
            pass
        except:
            raise
