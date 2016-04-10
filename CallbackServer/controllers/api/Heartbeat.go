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
	"github.com/zenazn/goji/web"

	"qdserver/CallbackServer/command"
	"qdserver/lib/utils"
)

// Heartbeat route
// curl -d '{"SystemUUID":"29c0b4f4-d6ab-46d8-604a-268863059f76","CustomerUUID":"3d794551-91a0-4db4-6296-ffcbfc5577f9","CurrentClientTime":1415296881 }' http://127.0.0.1:8080/api/v1/Heartbeat
func (controller *Controller) Heartbeat(c web.C, r *http.Request) (string, int) {
	// Parse body into json
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Unable to read body")
		return "", http.StatusBadRequest
	}

	type Heartbeat struct {
		SystemUUID        string
		CustomerUUID      string
		CurrentClientTime int64
	}
	var heartbeat Heartbeat
	err = json.Unmarshal(body, &heartbeat)
	if err != nil {
		log.Errorf("Unable to unmarshal json")
		return "", http.StatusBadRequest
	}

	db := controller.GetDatabase(c)

	systemID, err := utils.GetSystemIDFromUUID(db, heartbeat.SystemUUID, heartbeat.CustomerUUID)
	if err != nil {
		log.Errorf("Unable to parse get System ID, %v", err)
		return "", http.StatusBadRequest
	}
	log.Infof("SystemID: %d", systemID)

	return GenerateResponseToAgent(db, systemID, command.Nop())
}
