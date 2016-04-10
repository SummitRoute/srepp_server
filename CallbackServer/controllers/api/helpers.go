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
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"

	"qdserver/CallbackServer/command"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// GenerateResponseToAgent creates a response the agent expects given a command
// If the command is a NOP, check if we have any tasks
func GenerateResponseToAgent(db *gorp.DbMap, systemID int64, response command.ResponseToAgent) (string, int) {
	if response == command.Nop() {
		var tasks []models.Task
		_, err := db.Select(&tasks, `SELECT *
						FROM Tasks t
						WHERE t.SystemID=:systemID and t.DeployedToAgentDate = 0
						ORDER BY CreationDate`,
			map[string]interface{}{
				"systemID": systemID,
			})
		if err != nil {
			log.Errorf("Problem searching Tasks table... this is bad, %v", err)
			// TODO Need to do something in case of this error
		} else if len(tasks) > 0 {
			task := tasks[0]

			task.DeployedToAgentDate = utils.DBTimeNow()
			// TODO MUST I need to provide a task ID to the agent so when it responds I can mark this as completed.  Else if I send
			//   the agent a task and it fails, we never resend.

			// Update it in the DB
			_, err := db.Update(&task)
			if err != nil {
				log.Errorf("Can't update task: %v", err)
				// TODO Need to do something smart with this error
			}

			return task.Command, http.StatusOK
		}
	}

	// json response
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Unable to marshal json response, %v", err)
		return "", http.StatusBadRequest
	}

	return string(responseBytes), http.StatusOK
}
