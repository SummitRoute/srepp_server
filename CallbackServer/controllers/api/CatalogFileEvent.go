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
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq" // Needed for gorp
	"github.com/zenazn/goji/web"

	"qdserver/CallbackServer/command"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// CatalogFileEvent route is called by clients when a catalog file is used
func (controller *Controller) CatalogFileEvent(c web.C, r *http.Request) (string, int) {
	// Parse body into json
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Unable to read body")
		return "", http.StatusBadRequest
	}

	type eventFromClient struct {
		SystemUUID        string
		CustomerUUID      string
		CurrentClientTime int64
		TimeOfEvent       int64
		Path              string
		Sha256            string // Sha256 of the file
		Size              int
	}

	var event eventFromClient
	err = json.Unmarshal(body, &event)
	if err != nil {
		log.Errorf("Unable to unmarshal json, %v", err)
		return "", http.StatusBadRequest
	}

	db := controller.GetDatabase(c)

	systemID, err := utils.GetSystemIDFromUUID(db, event.SystemUUID, event.CustomerUUID)
	if err != nil {
		log.Errorf("Unable to parse get System ID, %v", err)
		return "", http.StatusBadRequest
	}

	// Add the Catalog to the DB
	Sha256, err := hex.DecodeString(event.Sha256)
	if err != nil {
		log.Errorf("Unable to decode Sha256: %v", err)
		return "", http.StatusBadRequest
	}

	timeOfEvent := (utils.DBTimeNow() - event.CurrentClientTime) + event.TimeOfEvent

	// Check if we've seen this executable before
	var catalogID int64
	catalogID, err = db.SelectInt(`SELECT ID
		FROM CatalogFiles
		WHERE Sha256=:sha256`,
		map[string]interface{}{
			"sha256": Sha256,
		})
	if err != nil {
		log.Errorf("Error checking for file: %v", err)
		return "", http.StatusBadRequest
	}
	if catalogID == 0 {
		log.Infof("New file, adding it to the DB")

		// Assuming executable has not been found, so adding it to the DB
		catalogFileInsert := &models.CatalogFile{
			FilePath:  event.Path,
			Sha256:    Sha256,
			Size:      event.Size,
			FirstSeen: timeOfEvent,
		}

		// Save it
		if err = db.Insert(catalogFileInsert); err != nil {
			log.Errorf("Error while creating processEvent: %v", err)
			return "", http.StatusBadRequest
		}

		catalogID = catalogFileInsert.ID

		// Create a task to tell this agent to get that file
		agentCommand := command.GetCatalogFileByHash(event.Sha256)
		if err = command.AddTask(db, systemID, agentCommand); err != nil {
			log.Errorf("Unable to add task for agent, %v", err)
			return "", http.StatusBadRequest
		}
	}

	// Else if a catalogID exists, then we've seen this catalog before so we don't need to do anything
	// TODO Might be smart to keep track of all the agents that have this catalog

	return GenerateResponseToAgent(db, systemID, command.Success())
}
