#!/usr/bin/env python

# Copyright 2011 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Author: caronni@google.com (Germano Caronni)
# Partially derived from pyasn1 examples.

"""Subset of PKCS#7 message syntax."""

# Use as a guide: http://dev.pyra-handheld.com/index.php/p/pyra-kernel/source/tree/f5e961c49775eb4fdfcc93e3f39b845d571ee6cb/crypto/asymmetric_keys/pkcs7.asn1

from pyasn1.type import namedtype
from pyasn1.type import tag
from pyasn1.type import univ
import x509
from x509_time import Time


class Attribute(univ.Sequence):
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('type', x509.AttributeType()),
      namedtype.NamedType('values', univ.SetOf(
          componentType=x509.AttributeValue())))


class ContentType(univ.ObjectIdentifier):
  pass


class Version(univ.Integer):
  pass


class DigestAlgorithmIdentifier(x509.AlgorithmIdentifier):
  pass


class DigestAlgorithmIdentifiers(univ.SetOf):
  componentType = DigestAlgorithmIdentifier()


class Digest(univ.OctetString):
  pass


class DigestInfo(univ.Sequence):
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('digestAlgorithm', DigestAlgorithmIdentifier()),
      namedtype.NamedType('digest', Digest()))


class ContentInfo(univ.Sequence):
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('contentType', ContentType()),
      namedtype.OptionalNamedType('content', univ.Any().subtype(
          explicitTag=tag.Tag(tag.tagClassContext,
                              tag.tagFormatConstructed, 0))))


class IssuerAndSerialNumber(univ.Sequence):
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('issuer', x509.Name()),
      namedtype.NamedType('serialNumber', x509.CertificateSerialNumber()))


class Attributes(univ.SetOf):
  componentType = Attribute()


class ExtendedCertificateInfo(univ.Sequence):
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('version', Version()),
      namedtype.NamedType('certificate', x509.Certificate()),
      namedtype.NamedType('attributes', Attributes()))


class SignatureAlgorithmIdentifier(x509.AlgorithmIdentifier):
  pass


class Signature(univ.BitString):
  pass


class ExtendedCertificate(univ.Sequence):
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('extendedCertificateInfo', ExtendedCertificateInfo()),
      namedtype.NamedType('signatureAlgorithm', SignatureAlgorithmIdentifier()),
      namedtype.NamedType('signature', Signature()))


class ExtendedCertificateOrCertificate(univ.Choice):
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('certificate', x509.Certificate()),
      namedtype.NamedType('extendedCertificate', ExtendedCertificate().subtype(
          implicitTag=tag.Tag(tag.tagClassContext,
                              tag.tagFormatConstructed, 0))))


class ExtendedCertificatesAndCertificates(univ.SetOf):
  componentType = ExtendedCertificateOrCertificate()


class SerialNumber(univ.Integer):
  pass


class CertificateRevocationLists(univ.Any):
  pass


class DigestEncryptionAlgorithmIdentifier(x509.AlgorithmIdentifier):
  pass


class EncryptedDigest(univ.OctetString):
  pass


class SignerInfo(univ.Sequence):
  """As defined by PKCS#7."""
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('version', Version()),
      namedtype.NamedType('issuerAndSerialNumber', IssuerAndSerialNumber()),
      namedtype.NamedType('digestAlgorithm', DigestAlgorithmIdentifier()),
      namedtype.OptionalNamedType(
          'authenticatedAttributes', Attributes().subtype(implicitTag=tag.Tag(
              tag.tagClassContext, tag.tagFormatConstructed, 0))),
      namedtype.NamedType('digestEncryptionAlgorithm',
                          DigestEncryptionAlgorithmIdentifier()),
      namedtype.NamedType('encryptedDigest', EncryptedDigest()),
      namedtype.OptionalNamedType('unauthenticatedAttributes',
                                  Attributes().subtype(implicitTag=tag.Tag(
                                      tag.tagClassContext,
                                      tag.tagFormatConstructed, 1))))


class SignerInfos(univ.SetOf):
  componentType = SignerInfo()


class SignedData(univ.Sequence):
  """As defined by PKCS#7."""
  componentType = namedtype.NamedTypes(
      namedtype.NamedType('version', Version()),
      namedtype.NamedType('digestAlgorithms', DigestAlgorithmIdentifiers()),
      namedtype.NamedType('contentInfo', ContentInfo()),
      namedtype.OptionalNamedType(
          'certificates', ExtendedCertificatesAndCertificates().subtype(
              implicitTag=tag.Tag(tag.tagClassContext,
                                  tag.tagFormatConstructed, 0))),
      namedtype.OptionalNamedType('crls', CertificateRevocationLists().subtype(
          implicitTag=tag.Tag(tag.tagClassContext,
                              tag.tagFormatConstructed, 1))),
      namedtype.NamedType('signerInfos', SignerInfos()))

class CountersignInfo(SignerInfo):
  pass


class SigningTime(Time):
  pass


class CatalogFile(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('pkcs7_oid', univ.ObjectIdentifier()),
        namedtype.NamedType('PKCS7',  SignedData().subtype(explicitTag=tag.Tag(
            128, 32, 0))))





#
# The items below are not defined anywhere so I've reveresed them
#

class CatalogIDSequence(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('OID', univ.ObjectIdentifier())) # Should be szOID_CATALOG_LIST = "1.3.6.1.4.1.311.12.1.1"


class OIDSequence(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('OID', univ.ObjectIdentifier()), # Should be szOID_CATALOG_LIST_MEMBER
        namedtype.NamedType('Null',univ.Null()))




class HashObject(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('OIDSequence', OIDSequence()), # Should be "1.3.14.3.2.26" = Sha1
        namedtype.NamedType('Hash',univ.OctetString()))


class DataObjectSequence(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('PEImageData', univ.Any()), # Should be SPC_PE_IMAGE_DATA_OBJID inside the following: Sequence(OID(SPC_PE_IMAGE_DATA_OBJID), Sequence(BitString, [0]([1]([0](28 byte hex encoded "<<<invalid>>>")))
        namedtype.NamedType('HashObject', HashObject()))

class DataObjectSet(univ.Set):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('DataObjectSequence', DataObjectSequence()))


class IndirectDataObject(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('OID', univ.ObjectIdentifier()), # Should be SPC_INDIRECT_DATA_OBJID
        namedtype.NamedType('DataObjectSet', univ.Any())) #DataObjectSet()))


class MemberInfo(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('OID', univ.ObjectIdentifier()), # Should be CAT_MEMBERINFO_OBJID
        namedtype.NamedType('Object', univ.Any()))


class CatalogMemberSet(univ.SetOf):
    componentType = MemberInfo()

class CatalogMember(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('DoubleEncoded',  univ.OctetString()), # 82 bytes to say the sha1 hash in hex encoded ascii of the hex of the hash
        namedtype.NamedType('CatalogMemberSet', CatalogMemberSet())
        )

class CatalogList(univ.SequenceOf):
    componentType = CatalogMember()

class CatNamedValue(univ.Sequence):
    pass

class CertificateTrustList(univ.Sequence):
    componentType = namedtype.NamedTypes(
        namedtype.NamedType('catalogListOID', CatalogIDSequence()),
        namedtype.NamedType('digest',  univ.OctetString()), # Should be 16 bytes
        namedtype.NamedType('time',  Time()),
        namedtype.NamedType('memberOID', OIDSequence()),
        namedtype.NamedType('catalogList', CatalogList()),
        namedtype.NamedType('catNameValue',  CatNamedValue().subtype(explicitTag=tag.Tag(128, 32, 0)))
        )
