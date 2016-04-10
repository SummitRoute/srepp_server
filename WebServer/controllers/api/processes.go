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
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web"

	"qdserver/WebServer/helpers"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// ProcessesJSON route
func (controller *Controller) ProcessesJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Read parameters
	start := helpers.GetParam(r.URL.Query(), "start", "^[0-9]*$", "0")
	length := helpers.GetParam(r.URL.Query(), "length", "^[0-9]*$", "25")
	ilength, err := strconv.Atoi(length)
	if err != nil || ilength > 100 {
		length = "100"
	}
	filter := helpers.GetParam(r.URL.Query(), "filter", "^[a-z]*$", "")

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	type ProcessData struct {
		Sha256      []byte
		FilePath    string
		CommandLine string
		EventTime   int64
	}

	var processes []ProcessData

	// TODO Need to check count in a smarter way so I'm not duplicating all this code

	// Get count
	filter = "%" + filter + "%"
	count, err := db.SelectInt(`SELECT count(*)
			FROM systemSets ss, systems s, ProcessEvents p, ExecutableFiles f
			WHERE ss.CustomerID=:CustomerID and ss.ID =s.SystemSetID and s.ID=p.SystemID and p.ExecutableFileID=f.ID
				and p.FilePath ILIKE :filter`,
		map[string]interface{}{
			"CustomerID": user.CustomerID,
			"filter":     filter,
		})
	if err != nil {
		// TODO MUST This probably can happen if no processes are in the DB
		log.Errorf("Unable to find processes in DB, %v", err)
		return "", http.StatusBadRequest
	}

	_, err = db.Select(&processes, `SELECT
			f.Sha256, p.FilePath, p.CommandLine, p.EventTime
			FROM systemSets ss, systems s, ProcessEvents p, ExecutableFiles f
			WHERE ss.CustomerID=:CustomerID and ss.ID =s.SystemSetID and s.ID=p.SystemID and p.ExecutableFileID=f.ID
				and p.FilePath ILIKE :filter
			ORDER BY p.EventTime DESC
			LIMIT :limit OFFSET :offset`,
		map[string]interface{}{
			"CustomerID": user.CustomerID,
			"filter":     filter,
			"limit":      length,
			"offset":     start,
		})
	if err != nil {
		// TODO MUST This probably can happen if no processes are in the DB
		log.Errorf("Unable to find processes in DB, %v", err)
		return "", http.StatusBadRequest
	}

	type ProcessDataJSON struct {
		Sha256      string
		FilePath    string
		CommandLine string
		EventTime   string
	}

	type DataTablesJSON struct {
		ITotalRecords        int               `json:"iTotalRecords"`
		ITotalDisplayRecords int               `json:"iTotalDisplayRecords"`
		SEcho                string            `json:"sEcho"`
		AaData               []ProcessDataJSON `json:"aaData"`
	}

	var dataTablesJSON DataTablesJSON
	dataTablesJSON.ITotalRecords = int(count)
	dataTablesJSON.ITotalDisplayRecords = len(processes)
	dataTablesJSON.AaData = make([]ProcessDataJSON, len(processes), len(processes))

	for index, process := range processes {
		var processDataJSON ProcessDataJSON
		processDataJSON.Sha256 = hex.EncodeToString(process.Sha256)
		processDataJSON.FilePath = process.FilePath
		processDataJSON.CommandLine = process.CommandLine

		processDataJSON.EventTime = utils.Int64ToUnixTimeString(process.EventTime, true)

		dataTablesJSON.AaData[index] = processDataJSON
	}

	contents, err := json.Marshal(dataTablesJSON)
	if err != nil {
		log.Errorf("Unable to marshal json")
		return "", http.StatusBadRequest
	}

	return string(contents), http.StatusOK
}
