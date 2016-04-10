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
	"net/http"

	log "github.com/Sirupsen/logrus"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/zenazn/goji/web"

	"qdserver/WebServer/helpers"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// getSortSystemColumn receives a GET parameter and converts it to our column names
func getSortSystemColumn(str string) string {
	switch str {
	case "Machine Name":
		return "MachineName"
	case "Comment":
		return "Comment"
	case "OS":
		return "OS"
	case "Manufacturer":
		return "Manufacturer"
	case "Model":
		return "Model"
	case "First Seen":
		return "FirstSeen"
	case "Last Seen":
		return "LastSeen"
	default:
		return "LastSeen"
	}
}

// SystemData is received directly from the database
type SystemData struct {
	SystemUUID   []byte
	MachineGUID  string
	AgentVersion string
	Comment      string
	OSHumanName  string
	OSVersion    string
	Manufacturer string
	Model        string
	Arch         string
	MachineName  string
	FirstSeen    int64
	LastSeen     int64
}

// SystemDataJSON is sent in json responses
type SystemDataJSON struct {
	System       string
	MachineGUID  string
	AgentVersion string
	Comment      string
	OSHumanName  string
	OSVersion    string
	Manufacturer string
	Model        string
	Arch         string
	MachineName  string
	FirstSeen    string
	LastSeen     string
}

// SystemInfoJSON route
func (controller *Controller) SystemInfoJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	//
	// Read parameters
	//
	systemUUIDStr := helpers.GetParam(r.URL.Query(), "uuid", "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$", "")
	if systemUUIDStr == "" {
		log.Errorf("Badly formatted uuid string")
		return "", http.StatusBadRequest
	}
	systemUUID, err := utils.UUIDStringToBytes(systemUUIDStr)
	if systemUUIDStr == "" {
		log.Errorf("Badly formatted uuid string (could not parse)")
		return "", http.StatusBadRequest
	}

	log.Infof("Looking for %s", systemUUIDStr)

	filterVars := map[string]interface{}{
		"customerID": user.CustomerID,
		"systemUUID": systemUUID,
	}

	var system SystemData
	err = db.SelectOne(&system, `SELECT
		s.SystemUUID, s.MachineGUID, s.AgentVersion, s.Comment, s.OSHumanName, s.OSVersion, s.Manufacturer, s.Model, s.Arch, s.MachineName, s.FirstSeen, s.LastSeen
		FROM systemSets ss, systems s
		WHERE CustomerID=:customerID and ss.ID =s.SystemSetID and s.SystemUUID=:systemUUID`,
		filterVars)
	if err != nil {
		// TODO MUST This probably can happen if no agents have called in yet
		log.Errorf("Unable to find systems in DB, %v", err)
		return "", http.StatusBadRequest
	}

	var systemDataJSON SystemDataJSON
	systemDataJSON.System, err = utils.ByteArrayToUUIDString(system.SystemUUID)
	if err != nil {
		log.Errorf("Unable to parse SystemUUID, %v", err)
		return "", http.StatusBadRequest
	}

	// TODO Other than the SystemUUID, there is a ridiculous amount of duplicate code here copying the same things between systemDataJSON and system
	systemDataJSON.MachineGUID = system.MachineGUID
	systemDataJSON.AgentVersion = system.AgentVersion
	systemDataJSON.Comment = system.Comment
	systemDataJSON.OSHumanName = system.OSHumanName
	systemDataJSON.OSVersion = system.OSVersion
	systemDataJSON.Manufacturer = system.Manufacturer
	systemDataJSON.Model = system.Model
	systemDataJSON.Arch = system.Arch
	systemDataJSON.MachineName = system.MachineName

	systemDataJSON.LastSeen = utils.Int64ToUnixTimeString(system.LastSeen, false)
	systemDataJSON.FirstSeen = utils.Int64ToUnixTimeString(system.FirstSeen, false)

	contents, err := json.Marshal(systemDataJSON)
	if err != nil {
		log.Errorf("Unable to marshal json")
		return "", http.StatusBadRequest
	}

	return string(contents), http.StatusOK

}

// SystemsJSON route
func (controller *Controller) SystemsJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	//
	// Read parameters
	//
	dataTableParams := helpers.GetDatatableParams(r.URL.Query(), "LastSeen")
	dataTableParams.SortColumn = getSortSystemColumn(dataTableParams.SortColumn)

	filterVars := map[string]interface{}{
		"customerID": user.CustomerID,
		"limit":      dataTableParams.Length,
		"offset":     dataTableParams.Start,
	}

	sqlString := `systemSets ss, systems s
	WHERE CustomerID=:customerID and ss.ID =s.SystemSetID`

	//
	// Get count
	//
	count, err := db.SelectInt(fmt.Sprintf("SELECT count(*) FROM %s", sqlString),
		filterVars)
	if err != nil {
		// TODO MUST This probably can happen if no files are in the DB
		log.Errorf("Unable to find systems in DB, %v", err)
		return "", http.StatusBadRequest
	}

	var systems []SystemData
	sqlStatement := fmt.Sprintf(`SELECT
		s.SystemUUID, s.AgentVersion, s.Comment, s.OSHumanName, s.OSVersion, s.Manufacturer, s.Model, s.Arch, s.MachineName, s.FirstSeen, s.LastSeen
		FROM %s
		ORDER BY %s %s
		LIMIT :limit OFFSET :offset`, sqlString, dataTableParams.SortColumn, dataTableParams.SortOrder)
	_, err = db.Select(&systems, sqlStatement,
		filterVars)
	if err != nil {
		// TODO MUST This probably can happen if no agents have called in yet
		log.Errorf("Unable to find systems in DB, %v", err)
		return "", http.StatusBadRequest
	}

	type DataTablesJSON struct {
		ITotalRecords        int              `json:"iTotalRecords"`
		ITotalDisplayRecords int              `json:"iTotalDisplayRecords"`
		SEcho                string           `json:"sEcho"`
		AaData               []SystemDataJSON `json:"aaData"`
	}

	var dataTablesJSON DataTablesJSON
	dataTablesJSON.ITotalRecords = int(count)
	dataTablesJSON.ITotalDisplayRecords = len(systems)
	dataTablesJSON.AaData = make([]SystemDataJSON, len(systems), len(systems))

	for index, system := range systems {
		var systemDataJSON SystemDataJSON
		systemUUIDasUUID, err := uuid.Parse(system.SystemUUID)
		if err != nil {
			log.Errorf("Unable to parse SystemUUID, %v", err)
			return "", http.StatusBadRequest
		}
		systemDataJSON.System = systemUUIDasUUID.String()

		// TODO Other than the SystemUUID, there is a ridiculous amount of duplicate code here copying the same things between systemDataJSON and system
		systemDataJSON.AgentVersion = system.AgentVersion
		systemDataJSON.Comment = system.Comment
		systemDataJSON.OSHumanName = system.OSHumanName
		systemDataJSON.OSVersion = system.OSVersion
		systemDataJSON.Manufacturer = system.Manufacturer
		systemDataJSON.Model = system.Model
		systemDataJSON.Arch = system.Arch
		systemDataJSON.MachineName = system.MachineName

		systemDataJSON.LastSeen = utils.Int64ToUnixTimeString(system.LastSeen, true)
		systemDataJSON.FirstSeen = utils.Int64ToUnixTimeString(system.FirstSeen, true)

		dataTablesJSON.AaData[index] = systemDataJSON

	}

	contents, err := json.Marshal(dataTablesJSON)
	if err != nil {
		log.Errorf("Unable to marshal json")
		return "", http.StatusBadRequest
	}

	return string(contents), http.StatusOK
}
