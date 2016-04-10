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

// ProcessEvent route is called by clients when a process is started
//
// Test with: curl -d '{"SystemUUID":"631a838a-9509-46de-7612-91f0a246cce9","CustomerUUID":"3d794551-91a0-4db4-6296-ffcbfc5577f9","CurrentClientTime":1,"TimeOfEvent":1,"Type":1,"PID":1,"PPID":1,"CommandLine":"this","Md5":"4ec38625fdb2bd3cf7e237f4b1387c04","Sha1":"142124bc228f603235f537298a893a95b6af5e28","Sha256":"5e2091457a435e68cc669189432635442360d70cf3f44880ccf1bd35ef393be1"}' http://127.0.0.1:8080/api/v1/ProcessEvent
func (controller *Controller) ProcessEvent(c web.C, r *http.Request) (string, int) {
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
		Type              int
		PID               int64
		PPID              int64
		Path              string
		CommandLine       string
		Md5               string
		Sha1              string
		Sha256            string // Sha256 of the file
		Size              int
		IsSigned          bool
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
		log.Errorf("Unable to find ID for System %s (customer: %s), %v", event.SystemUUID, event.CustomerUUID, err)
		return "", http.StatusBadRequest
	}

	log.Infof("ProcessEvent from system: %d\n", systemID)

	// Add the Executable to the DB
	Md5, err := hex.DecodeString(event.Md5)
	if err != nil {
		log.Errorf("Unable to decode MD5: %v", err)
		return "", http.StatusBadRequest
	}

	Sha1, err := hex.DecodeString(event.Sha1)
	if err != nil {
		log.Errorf("Unable to decode Sha1: %v", err)
		return "", http.StatusBadRequest
	}

	Sha256, err := hex.DecodeString(event.Sha256)
	if err != nil {
		log.Errorf("Unable to decode Sha256: %v", err)
		return "", http.StatusBadRequest
	}

	timeOfEvent := (utils.DBTimeNow() - event.CurrentClientTime) + event.TimeOfEvent

	// Check if we've seen this executable before
	var executableID int64
	executableID, err = db.SelectInt(`SELECT ID
		FROM ExecutableFiles
		WHERE Sha256=:sha256`,
		map[string]interface{}{
			"sha256": Sha256,
		})
	if err != nil {
		log.Errorf("Error checking for file: %v", err)
		return "", http.StatusBadRequest
	}
	if executableID == 0 {
		log.Infof("New file, adding it to the DB")

		// Assuming executable has not been found, so adding it to the DB
		executableFileInsert := &models.ExecutableFile{
			Md5:       Md5,
			Sha1:      Sha1,
			Sha256:    Sha256,
			Size:      event.Size,
			FirstSeen: timeOfEvent,
			IsSigned:  event.IsSigned,
		}

		// Save it
		if err = db.Insert(executableFileInsert); err != nil {
			// If you get an error about duplicate values, one fix is to run: select setval('executablefiles_id_seq', (select max(id) + 1 from executablefiles));
			// I don't know why it forget about that sequence number
			log.Errorf("Error while creating executable file: %v", err)
			return "", http.StatusBadRequest
		}

		executableID = executableFileInsert.ID

		// Create a task to tell this agent to get that file
		agentCommand := command.GetFileByHash(event.Sha256)
		if err = command.AddTask(db, systemID, agentCommand); err != nil {
			log.Errorf("Unable to add task for agent, %v", err)
			return "", http.StatusBadRequest
		}
	}

	// Stash the event in the DB
	processEventInsert := &models.ProcessEvent{
		SystemID:         systemID,
		ExecutableFileID: executableID,
		PID:              event.PID,
		PPID:             event.PPID,
		FilePath:         event.Path,
		CommandLine:      event.CommandLine,
		EventTime:        timeOfEvent,
	}

	// Save it
	if err = db.Insert(processEventInsert); err != nil {
		log.Errorf("Error while creating processEvent: %v", err)
		return "", http.StatusBadRequest
	}

	//
	// Record this file was seen on this system
	//
	if err = utils.RecordFileSeenOnSystem(db, executableID, systemID, timeOfEvent, event.Path); err != nil {
		log.Errorf("Error adding file to fileToSystemMap: %v", err)
		return "", http.StatusBadRequest
	}

	return GenerateResponseToAgent(db, systemID, command.Success())
}
