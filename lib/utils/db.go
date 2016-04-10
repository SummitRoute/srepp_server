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
	"database/sql"
	"qdserver/lib/models"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"
)

// RecordFileSeenOnSystem helper function to update the FileToSystemMap table
//
// Inserts or updates the table to show this file has been seen on this system and provide some meta data
//
func RecordFileSeenOnSystem(db *gorp.DbMap, fileID int64, systemID int64, timeOfEvent int64, filePath string) (err error) {
	var fileToSystemMap models.FileToSystemMap

	// Check if we have a copy
	err = db.SelectOne(&fileToSystemMap, `SELECT *
      FROM FileToSystemMap
      WHERE FileID=:fileID and SystemID=:systemID`,
		map[string]interface{}{
			"fileID":   fileID,
			"systemID": systemID,
		})

	// Set common values whether we add or update it
	fileToSystemMap.LastSeen = timeOfEvent
	fileToSystemMap.FilePath = filePath

	if err == sql.ErrNoRows {
		// This file has not been seen on this system, so add it

		fileToSystemMap.FirstSeen = timeOfEvent
		fileToSystemMap.SystemID = systemID
		fileToSystemMap.FileID = fileID

		// Save it
		err = db.Insert(&fileToSystemMap)
	} else if err == nil {
		_, err = db.Update(&fileToSystemMap)
	}

	// Return any errors collected
	return
}

// GetNullString returns the string value if valid, else the default value
func GetNullString(str sql.NullString, defaultStr string) string {
	if str.Valid != true {
		return defaultStr
	}
	return str.String
}

// GetSystemIDFromUUID Returns the System ID given the System UUID and CustomerUUID
func GetSystemIDFromUUID(db *gorp.DbMap, SystemUUIDString string, CustomerUUIDString string) (int64, error) {
	SystemUUID, err := UUIDStringToBytes(SystemUUIDString)
	if err != nil {
		log.Warningf("Unable to convert SystemUUID (%s) to bytes, %v", SystemUUIDString, err)
		return 0, nil
	}

	CustomerUUID, err := UUIDStringToBytes(CustomerUUIDString)
	if err != nil {
		log.Warningf("Unable to convert CustomerUUID (%s) to bytes, %v", CustomerUUIDString, err)
		return 0, nil
	}

	SystemID, err := db.SelectInt(`SELECT s.ID
            FROM systemSets ss, systems s, customers c
            WHERE c.UUID=:customerUUID and ss.CustomerID = c.ID and ss.ID =s.SystemSetID and s.SystemUUID = :systemUUID`,
		map[string]interface{}{
			"customerUUID": CustomerUUID,
			"systemUUID":   SystemUUID,
		})
	if err != nil {
		log.Warningf("Could not find system ID, rogue client? %v", CustomerUUIDString, err)
		return 0, err
	}

	return SystemID, nil
}

// DBTimeNow returns the current unix time as suitable for the database
func DBTimeNow() int64 {
	return time.Now().UTC().Unix()
}

// ConvertYYYYMMDDtoUnix takes a date in the form 2013-01-20 and converts it to unix time for the database
func ConvertYYYYMMDDtoUnix(input string) int64 {
	const timeFormat = "2006-01-02"

	timezone, err := time.LoadLocation("MST")
	if err != nil {
		log.Errorf("Unable to set timezone, %v", err)
		timezone = time.Local
	}

	t, err := time.ParseInLocation(timeFormat, input, timezone)
	if err != nil {
		log.Errorf("Unable to parse time, %v", err)
		t = time.Now()
	}

	return t.Unix()
}

// GetStringMatchOperator translates an operator from JSON into SQL
func GetStringMatchOperator(jsonOperator string) string {
	if jsonOperator == "==" {
		return "LIKE"
	} else if jsonOperator == "!=" {
		return "NOT LIKE"
	}

	log.Warningf("Unknown string operator: %s", jsonOperator)
	return "LIKE"
}

// GetStringSubstrMatchOperator translates an operator from JSON into SQL
func GetStringSubstrMatchOperator(jsonOperator string) string {
	if jsonOperator == "==" {
		return "LIKE"
	} else if jsonOperator == "contains" {
		return "ILIKE"
	} else if jsonOperator == "!contains" {
		return "NOT ILIKE"
	} else if jsonOperator == "!=" {
		return "NOT LIKE"
	}

	log.Warningf("Unknown string operator: %s", jsonOperator)
	return "LIKE"
}

// GetIntMatchOperator translates an operator from JSON into SQL
func GetIntMatchOperator(jsonOperator string) string {
	if jsonOperator == "==" {
		return "="
	} else if jsonOperator == "!=" {
		return "!="
	} else if jsonOperator == "<" {
		return "<"
	} else if jsonOperator == "<=" {
		return "<="
	} else if jsonOperator == ">" {
		return ">"
	} else if jsonOperator == ">=" {
		return ">="
	}

	log.Warningf("Unknown int operator: %s", jsonOperator)
	return "="
}

// GetDateMatchOperator translates an operator from JSON into SQL
func GetDateMatchOperator(jsonOperator string) string {
	return GetIntMatchOperator(jsonOperator)
}
