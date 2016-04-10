////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////


package command

import (
	"encoding/json"

	"github.com/coopernurse/gorp"

	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// ResponseToAgent struct
type ResponseToAgent struct {
	Command   string
	Arguments interface{}
}

// SetSystemUUID command
func SetSystemUUID(systemUUID string) ResponseToAgent {
	response := ResponseToAgent{
		Command: "SetSystemUUID",
		Arguments: struct {
			SystemUUID string
		}{
			systemUUID,
		}}
	return response
}

// GetFileByHash command
func GetFileByHash(sha256 string) ResponseToAgent {
	response := ResponseToAgent{
		Command: "GetFileByHash",
		Arguments: struct {
			Sha256 string
		}{
			sha256,
		}}
	return response
}

// GetCatalogFileByHash command
func GetCatalogFileByHash(sha256 string) ResponseToAgent {
	response := ResponseToAgent{
		Command: "GetCatalogFileByHash",
		Arguments: struct {
			Sha256 string
		}{
			sha256,
		}}
	return response
}

// Update command
func Update(newVersion string) ResponseToAgent {
	response := ResponseToAgent{
		Command: "Update",
		Arguments: struct {
			NewVersion string
		}{
			newVersion,
		}}
	return response
}

// Stall command is used when something has gone wrong on the server, so we delay the client so I have time to hopefully fix it
func Stall() ResponseToAgent {
	response := ResponseToAgent{
		Command:   "Stall",
		Arguments: nil}
	return response
}

// Nop command
func Nop() ResponseToAgent {
	response := ResponseToAgent{
		Command:   "NOP",
		Arguments: nil}
	return response
}

// Success command tells the agent it was successful for times when process events are sent
func Success() ResponseToAgent {
	response := ResponseToAgent{
		Command:   "Success",
		Arguments: nil}
	return response
}

// AddTask stores a task to be sent to a system next time it calls in
func AddTask(db *gorp.DbMap, systemID int64, agentCommand ResponseToAgent) (err error) {
	// Convert command to json
	jsonCommand, err := json.Marshal(agentCommand)
	if err != nil {
		return err
	}

	task := &models.Task{
		SystemID:     systemID,
		CreationDate: utils.DBTimeNow(),
		Command:      string(jsonCommand),
	}

	// Save it
	if err = db.Insert(task); err != nil {
		return err
	}

	return nil
}
