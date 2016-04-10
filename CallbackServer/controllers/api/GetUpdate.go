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
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"qdserver/lib/utils"

	log "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq" // Needed for gorp
	"github.com/zenazn/goji/web"
)

// GetUpdate route
func (controller *Controller) GetUpdate(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Parse body into json
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Unable to read body")
		return "", http.StatusBadRequest
	}

	type ClientData struct {
		SystemUUID        string
		CustomerUUID      string
		CurrentClientTime int64
		Version           string
	}
	var clientData ClientData
	err = json.Unmarshal(body, &clientData)
	if err != nil {
		log.Errorf("Unable to unmarshal json")
		return "", http.StatusBadRequest
	}

	// Search database for what update we can use
	var UpdateToVersion string
	err = db.SelectOne(&UpdateToVersion, `SELECT VersionTo
		FROM Updates
		WHERE VersionFrom=:version
		ORDER BY VersionTo
		Limit 1`,
		map[string]interface{}{
			"version": clientData.Version,
		})
	if err != nil {
		log.Errorf("Error checking version to update to: %v", err)
		return "", http.StatusBadRequest
	}

	updaterName := fmt.Sprintf("update-%s-%s.exe", clientData.Version, UpdateToVersion)
	updatePath := path.Join("updates", updaterName)

	if !utils.CheckExists(updatePath) {
		// TODO Need to get file from S3 to local
		log.Errorf("Update file does not exist: %v", err)
		return "", http.StatusBadRequest
	}

	// File should now exist so read it
	contents, err := ioutil.ReadFile(updatePath)
	if err != nil {
		log.Errorf("Unable to read file, %v", err)
		return "", http.StatusBadRequest
	}

	c.Env["Content-Type"] = "application/octet-stream"
	c.Env["Content-Length"] = fmt.Sprintf("%d", len(contents))

	return string(contents), http.StatusOK
}
