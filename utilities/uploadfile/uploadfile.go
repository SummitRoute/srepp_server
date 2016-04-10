////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"os"

	"qdserver/lib/utils"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: uploadfile <filepath>\n")
		fmt.Printf("Simple test utility to ensure I can upload files to S3.\n")
		fmt.Printf("Given a file, generates a SHA256 hash and upload the binary to the hash path on S3\n")
		return
	}

	filename := os.Args[1]

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		return
	}

	defer f.Close()

	ExeUpload := utils.AwsS3{
		BucketName: "summitroute.file.uploads",
		AccessKey:  "YOUR_ACCESS_KEY",
		SecretKey:  "YOUR_SECRET_KEY",
		Region:     "us-east-1",
	}

	sha256, err := utils.Sha256File(f)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	sha256String := utils.ByteArrayToHexString(sha256)

	err = utils.UploadToS3HashPath(ExeUpload, sha256String, f)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("Success\n")
}
