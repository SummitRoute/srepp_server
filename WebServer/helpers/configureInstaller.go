////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package helpers

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	uuid "github.com/nu7hatch/gouuid"
)

// ConfigureInstaller configures an MSI with the required data
func ConfigureInstaller(fileIn string, fileOut string, group string) error {
	// Open file to read in
	f, err := os.OpenFile(fileIn, os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("Unable to open file")
		return err
	}
	defer f.Close()

	// Read the entire file into memory
	// TODO I should remember where to read from and also use a stream reader
	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("Unable to read file")
		return err
	}

	// Find our configuration blob
	prefix := []byte("---BEGIN_BLOB---")
	start := bytes.LastIndex(fileBytes, prefix)
	if start == -1 {
		return errors.New("Unable to find prefix string")
	}
	start = start + len(prefix)
	log.Infof("Start %d", start)

	postfix := []byte("---END_BLOB---")
	end := bytes.LastIndex(fileBytes, postfix)
	if end == -1 {
		return errors.New("Unable to find postfix string")
	}
	log.Infof("End %d", end)

	// Sanity check the blob size
	expectedBlobSize := 993
	if end-start != expectedBlobSize {
		return errors.New("Blob was not of the expected size")
	}

	//
	// Set the configuration
	//
	groupID, err := uuid.ParseHex(group)
	if err != nil {
		log.Errorf("Unable to parse group")
		return err
	}

	copy(fileBytes[start:end], (*groupID)[0:16])

	// Open file for writing
	err = ioutil.WriteFile(fileOut, fileBytes, 0644)
	if err != nil {
		log.Errorf("Unable to write file")
		return err
	}

	return nil
}
