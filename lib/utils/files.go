////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package utils

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
)

// ExpandSha256HexstringToPath converts a sha hash like
// "0a8ce026714e03e72c619307bd598add5f9b639cfd91437cb8d9c847bf9f6894" to
// "0a/8c/0a8ce026714e03e72c619307bd598add5f9b639cfd91437cb8d9c847bf9f6894"
// Assumes the string has already been sanity checked to ensure it's a lowercase sha256
func ExpandSha256HexstringToPath(fileName string) string {
	nameChars := strings.Split(fileName, "")
	root := nameChars[0] + nameChars[1]
	child := nameChars[2] + nameChars[3]
	return path.Join(root, child, fileName)
}

// Sha256File returns the sha256 hash of a file
func Sha256File(f *os.File) (hash []byte, err error) {
	offset, err := f.Seek(0, 0) // Seek to the start of the file
	if err != nil || offset != 0 {
		return nil, fmt.Errorf("Unable to seek to start of file: %v", err)
	}

	hasher := sha256.New()
	// TODO Should read chunks at a time
	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file: %v", err)
	}

	hasher.Write(fileBytes)
	hash = hasher.Sum(nil)

	return hash, nil
}

// AwsS3 is used in config.json files
type AwsS3 struct {
	BucketName string `json:"bucket_name"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	Region     string `json:"region"`
}

// UploadToS3HashPath given a filename that is a hex encoded hash, create a special directory structure for this file
// and upload to S3
func UploadToS3HashPath(awsS3 AwsS3, filename string, f *os.File) (err error) {
	return UploadToS3(awsS3, ExpandSha256HexstringToPath(filename), filename, f)

}

// UploadToS3 uploads the file to the AwsS3 bucket
// AwsS3 is the credentials
// filepath is the path inside the S3 bucket to upload to
// filename is name of the file
// f is the file handle for the file to upload
func UploadToS3(awsS3 AwsS3, filepath, filename string, f *os.File) (err error) {
	auth := aws.Auth{AccessKey: awsS3.AccessKey, SecretKey: awsS3.SecretKey}
	s3Service := s3.New(auth, aws.Regions["us-east-1"])
	bucket := s3Service.Bucket(awsS3.BucketName)

	// TODO Should read f in chunks

	// Seek to the start of the file
	offset, err := f.Seek(0, 0)
	if err != nil || offset != 0 {
		return fmt.Errorf("Unable to seek to start of: %v", err)
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("Unable to read file: %v", err)
	}

	disposition := fmt.Sprintf("attachment; filename=\"%s\"", filename)
	err = bucket.Put(filepath, contents, "binary/octet-stream", s3.Private, s3.Options{ContentDisposition: disposition})
	return err
}

// CheckExists returns true if the file exists, else false
func CheckExists(path string) (doesExist bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
