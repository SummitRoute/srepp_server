////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq" // Needed for gorp
	uuid "github.com/nu7hatch/gouuid"
	"github.com/zenazn/goji/web"

	"qdserver/CallbackServer/command"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// RegisterAgent route
func (controller *Controller) RegisterAgent(c web.C, r *http.Request) (string, int) {
	// Parse body into json
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Unable to read body")
		return "", http.StatusBadRequest
	}

	type Registration struct {
		CustomerUUID string
		AgentVersion string

		OSHumanName  string
		OSVersion    string
		Manufacturer string
		Model        string
		Arch         string
		MachineName  string
		MachineGUID  string
	}
	var registration Registration
	err = json.Unmarshal(body, &registration)
	if err != nil {
		log.Errorf("Unable to unmarshal json")
		return "", http.StatusBadRequest
	}

	// Generate SystemUUID for the agent
	SystemUUID, _ := uuid.NewV4()
	// TODO Check for collisions on the UUID

	//
	// Record this in the database
	//
	db := controller.GetDatabase(c)

	// Convert Customer UUID into a byte slice
	customerUUIDslice, err := utils.UUIDStringToBytes(registration.CustomerUUID)
	if err != nil {
		log.Errorf("Unable to parse Customer UUID")
		return "", http.StatusBadRequest
	}

	// Get Customer from UUID
	var customer models.Customer
	// TODO This needs a key or something so it isn't slow
	err = db.SelectOne(&customer, "select * from customers where UUID=:uuid",
		map[string]interface{}{
			"uuid": customerUUIDslice,
		})
	if err != nil {
		log.Errorf("Unable to find customer in DB")
		return "", http.StatusBadRequest
	}

	// Look for Default SystemSet, and if you can't find it, add it
	var systemSet models.SystemSet
	var systemSetID int64

	err = db.SelectOne(&systemSet, "select * from systemsets where CustomerID=:customerID and Name like :name",
		map[string]interface{}{
			"customerID": customer.ID,
			"name":       "Default",
		})
	if err != nil {
		// TODO Need to check the err means "not found" and not something bad

		// Add SystemSet
		systemSetInsert := &models.SystemSet{
			CustomerID:   customer.ID,
			Name:         "Default",
			CreationDate: utils.DBTimeNow(),
		}

		// Save it
		if err = db.Insert(systemSetInsert); err != nil {
			log.Errorf("Error while creating systemset: %v", err)
			return "", http.StatusBadRequest
		}
		systemSetID = systemSetInsert.ID
	} else {
		systemSetID = systemSet.ID
	}

	var SystemUUIDslice []byte
	SystemUUIDslice = SystemUUID[0:16]

	// Create system DB object
	systemInsert := &models.System{
		SystemSetID:  systemSetID,
		SystemUUID:   SystemUUIDslice,
		AgentVersion: registration.AgentVersion,

		OSHumanName:  registration.OSHumanName,
		OSVersion:    registration.OSVersion,
		Manufacturer: registration.Manufacturer,
		Model:        registration.Model,
		Arch:         registration.Arch,
		MachineName:  registration.MachineName,
		MachineGUID:  registration.MachineGUID,

		FirstSeen: utils.DBTimeNow(),
		LastSeen:  utils.DBTimeNow(),
	}

	// Save it
	if err = db.Insert(systemInsert); err != nil {
		log.Errorf("Error while creating system: %v", err)
		return "", http.StatusBadRequest
	}

	systemID := systemInsert.ID

	return GenerateResponseToAgent(db, systemID, command.SetSystemUUID(SystemUUID.String()))
}
