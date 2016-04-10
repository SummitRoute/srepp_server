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
	"encoding/hex"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	uuid "github.com/nu7hatch/gouuid"
)

// UUIDStringToBytes Given a UUID string, converts it to a byte slice for use with the DB
func UUIDStringToBytes(UUIDString string) ([]byte, error) {
	UUIDByteArray, err := uuid.ParseHex(UUIDString)
	if err != nil {
		return nil, err
	}

	var UUIDSlice []byte
	UUIDSlice = UUIDByteArray[0:16]

	return UUIDSlice, nil
}

// ByteArrayToUUIDString Given a byte array converts it to a UUID string
func ByteArrayToUUIDString(byteArray []byte) (UUIDString string, err error) {
	parsedUUID, err := uuid.Parse(byteArray)
	if err != nil {
		return
	}

	return parsedUUID.String(), nil

}

// ByteArrayToHexString converter
func ByteArrayToHexString(byteArray []byte) string {
	return hex.EncodeToString(byteArray)
}

// Int64ToUnixTimeString returns a human readable time and date string
func Int64ToUnixTimeString(secondsFromEpoch int64, dateOnly bool) string {
	// TODO need to display the user's timezone
	timezone, err := time.LoadLocation("MST")
	if err != nil {
		log.Errorf("Unable to set timezone, %v", err)
		timezone = time.Local
	}

	t := time.Unix(secondsFromEpoch, 0).In(timezone)
	if dateOnly {
		return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
	}
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}
