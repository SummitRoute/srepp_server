#!/usr/bin/env python

###############################################################################
#
# Summit Route End Point Protection
#
# This source code is licensed under the BSD-style license found in the
# LICENSE file in the root directory of this source tree.
#
###############################################################################

from sqlalchemy import *
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class ExecutableFiles(Base):
    __tablename__ = 'executablefiles'
    id = Column(BigInteger, primary_key=True)
    md5 = Column(LargeBinary(length=32))
    sha1 = Column(LargeBinary(length=40))
    sha256 = Column(LargeBinary(length=64))

    authenticodemd5 = Column(LargeBinary(length=32))
    authenticodesha1 = Column(LargeBinary(length=40))
    authenticodesha256 = Column(LargeBinary(length=64))

    codesectionsha256 = Column(LargeBinary(length=64))
    size = Column(Integer)
    issigned = Column(Boolean)
    firstseen = Column(BigInteger)
    executiontype = Column(Integer)

    uploaddate = Column(BigInteger)

    analysisdate = Column(BigInteger)

    companyname = Column(String)
    productversion = Column(String)
    productname = Column(String)
    filedescription = Column(String)
    internalname = Column(String)
    fileversion = Column(String)
    originalfilename = Column(String)
    architecture = Column(Integer)


class FileToSignerMap(Base):
    __tablename__ = 'filetosignermap'
    fileid = Column(BigInteger, primary_key=True)
    signerid = Column(BigInteger, primary_key=True)


class FileToCounterSignerMap(Base):
    __tablename__ = 'filetocountersignermap'
    fileid = Column(BigInteger, primary_key=True)
    timestamp = Column(BigInteger)
    signerid = Column(BigInteger, primary_key=True)


class Signer(Base):
    __tablename__ = 'signers'
    id = Column(BigInteger, primary_key=True)
    version = Column(Integer)
    subject = Column(String)
    subjectshortname = Column(String)
    serialnumber = Column(LargeBinary)
    digestalgorithm = Column(String)
    digestencryptionalgorithm = Column(String)
    digestencryptionalgorithmkeysize = Column(Integer)
    issuerid = Column(BigInteger)

class CatalogFile(Base):
    __tablename__ = 'catalogfiles'
    id = Column(BigInteger, primary_key=True)
    filepath  = Column(String)
    sha256    = Column(LargeBinary)
    size      = Column(Integer)
    firstseen = Column(BigInteger)
    uploaddate = Column(BigInteger)
    analysisdate = Column(BigInteger)
    signerid = Column(BigInteger)


class CertificateTrustList(Base):
    __tablename__ = 'certificatetrustlist'
    catalogid = Column(BigInteger, primary_key=True)
    hash    = Column(LargeBinary, primary_key=True)
    hashtype = Column(String)
    fileid = Column(BigInteger)
