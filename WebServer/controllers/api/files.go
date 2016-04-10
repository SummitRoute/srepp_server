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
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"
	"github.com/zenazn/goji/web"

	"qdserver/WebServer/helpers"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// getExeSortColumn receives a GET parameter and converts it to our column names
func getExeSortColumn(str string) string {
	switch str {
	case "Path":
		return "FilePath"
	case "First Seen":
		return "FirstSeen"
	case "Last Seen":
		return "LastSeen"
	case "Count":
		return "NumSystems"
	case "Product Name":
		return "ProductName"
	case "Company Name":
		return "CompanyName"
	case "Signer":
		return "SignerSubjectShortName"
	default:
		return "LastSeen"
	}
}

// FilteredFile contains the database information returned by FilterFiles
type FilteredFile struct {
	FileID                 int64
	Sha256                 []byte
	FilePath               string
	FirstSeen              int64
	LastSeen               int64
	ProductName            string
	CompanyName            string
	NumSystems             int
	SignerSubjectShortName sql.NullString
}

//
// GetSha256 returns the hex encoded value
//
func (ff *FilteredFile) GetSha256() string {
	return hex.EncodeToString(ff.Sha256)
}

//
// GetSignerSubjectShortName returns the SignerSubjectShortName with nulls taken into account
//
func (ff *FilteredFile) GetSignerSubjectShortName() string {
	return utils.GetNullString(ff.SignerSubjectShortName, "")
}

//
// FilterFiles is a class used to return data about a customer's files, given various filters
//
type FilterFiles struct {
	db                   *gorp.DbMap
	restrictedView       string
	restrictions         []string
	filterVars           map[string]interface{}
	outerRestrictionsStr string
	outerRestrictions    []string
	outerFilterVars      map[string]interface{}
}

//
// AddRestriction adds a WHERE clause to the view
//
func (ff *FilterFiles) AddRestriction(restriction string, varName string, varValue interface{}) {
	ff.restrictions = append(ff.restrictions, restriction)
	ff.filterVars[varName] = varValue
}

//
// AddDateRestriction returns a SQL string for a date restriction
//
func (ff *FilterFiles) AddDateRestriction(sqlVar string, operator string, secondsFromEpoch int64) {
	const SecondsInADay = 60 * 60 * 24
	restriction := ""

	if operator == "==" {
		restriction = fmt.Sprintf("(%s >= %d AND %s < %d)", sqlVar, secondsFromEpoch, sqlVar, secondsFromEpoch+SecondsInADay)
	} else if operator == "!=" {
		restriction = fmt.Sprintf("(%s < %d OR %s >= %d)", sqlVar, secondsFromEpoch, sqlVar, secondsFromEpoch+SecondsInADay)
	} else if operator == "<" {
		restriction = fmt.Sprintf("%s < %d", sqlVar, secondsFromEpoch)
	} else if operator == "<=" {
		restriction = fmt.Sprintf("%s < %d", sqlVar, secondsFromEpoch+SecondsInADay)
	} else if operator == ">" {
		restriction = fmt.Sprintf("%s >= %d", sqlVar, secondsFromEpoch+SecondsInADay)
	} else if operator == ">=" {
		restriction = fmt.Sprintf("%s >= %d", sqlVar, secondsFromEpoch)
	} else {
		log.Warningf("Unknown operator: %s", operator)
		return
	}

	ff.restrictions = append(ff.restrictions, restriction)

	return
}

//
// AddOuterRestriction adds a WHERE clause to the view that can acces the group variables
//
func (ff *FilterFiles) AddOuterRestriction(restriction string, varName string, varValue interface{}) {
	ff.outerRestrictions = append(ff.outerRestrictions, restriction)
	ff.outerFilterVars[varName] = varValue
}

//
// InitRestrictedView initialize th SQL view to use, this gets called automatically if you don't call it yourself.
//
func (ff *FilterFiles) InitRestrictedView() {
	restrictionsStr := ""
	for _, restriction := range ff.restrictions {
		restrictionsStr += " AND " + restriction
	}

	// Set up a view that we'll use that only returns the file IDs
	ff.restrictedView = fmt.Sprintf(`(SELECT FileId
		FROM systemSets ss, systems s, filetosystemmap fsm, ExecutableFiles f
		WHERE ss.CustomerID=:customerID AND ss.ID =s.SystemSetID AND s.ID=fsm.SystemID AND fsm.FileID=f.ID
		%s GROUP BY FileId) restrictors`, restrictionsStr)
}

//
// InitOuterRestrictions initialize th SQL view to use, this gets called automatically if you don't call it yourself.
//
func (ff *FilterFiles) InitOuterRestrictions() {
	outerRestrictionsStr := " "
	for _, restriction := range ff.outerRestrictions {
		outerRestrictionsStr += " AND " + restriction
	}

	ff.outerRestrictionsStr = outerRestrictionsStr

	for k, v := range ff.outerFilterVars {
		ff.filterVars[k] = v
	}
}

//
// GetCount returns the count of the view
//
func (ff *FilterFiles) GetCount() (int64, error) {
	sqlStatement := ff.GetSQL("count(*)", "")

	return ff.db.SelectInt(sqlStatement, ff.filterVars)
}

//
// GetFilteredFiles returns the sql data in a FilteredFile array
//
func (ff *FilterFiles) GetFilteredFiles() (files []FilteredFile, err error) {
	ordering := fmt.Sprintf("ORDER BY %s %s LIMIT :limit OFFSET :offset", ff.filterVars["sortColumn"], ff.filterVars["sortOrder"])

	sqlStatement := ff.GetSQL(`v.FileId,
		f.Sha256,
		v.FilePath as FilePath,
		v.FirstSeen as FirstSeen,
		v.LastSeen as LastSeen,
		f.ProductName as ProductName,
		f.CompanyName as CompanyName,
		v.NumSystems as NumSystems,
		fts.SubjectShortName as SignerSubjectShortName`,
		ordering)

	_, err = ff.db.Select(&files, sqlStatement, ff.filterVars)
	return
}

//
// GetSQL returns a SQL string for the given select string
//
func (ff *FilterFiles) GetSQL(whatToSelect string, ordering string) (sqlstring string) {
	if ff.restrictedView == "" {
		ff.InitRestrictedView()
	}

	if ff.outerRestrictionsStr == "" {
		ff.InitOuterRestrictions()
	}

	// Innermost select gets the file ids that will be of interest to us
	// then we get the aggregate info for all these files (firstseen,lastseen, and the an example path which potentially could be different than the filter)
	// then we collect all that plus any file specific info
	sqlStatement := fmt.Sprintf(`SELECT
	  %s
		FROM (
			SELECT
				restrictors.FileId,
				FIRST(fsm.FilePath) as FilePath,
				MIN(fsm.FirstSeen) AS FirstSeen,
				MAX(fsm.LastSeen) AS LastSeen,
				COUNT(*) as NumSystems
			FROM %s, systemSets ss, systems s, filetosystemmap fsm
			WHERE ss.CustomerID=:customerID
				AND ss.ID =s.SystemSetID
				AND s.ID=fsm.SystemID
				AND restrictors.FileId=fsm.FileId
			GROUP BY restrictors.FileId
			) v, ExecutableFiles f
			LEFT OUTER JOIN
			(SELECT * FROM FileToSignerMap ftsm, Signers s WHERE ftsm.SignerID = s.ID) fts
			ON f.id = fts.FileID
			WHERE v.FileId = f.id %s
			%s`, whatToSelect, ff.restrictedView, ff.outerRestrictionsStr, ordering)

	return sqlStatement
}

//
// SetFilterVar sets a variable to be used in the SQL
//
func (ff *FilterFiles) SetFilterVar(name string, value interface{}) {
	ff.filterVars[name] = value
}

//
// NewFilterFiles contructs a FilterFiles object
//
func NewFilterFiles(db *gorp.DbMap, customerID int64) *FilterFiles {
	restrictionsMaxSize := 0 // Arbitrary
	ff := &FilterFiles{db: db, restrictedView: "", restrictions: make([]string, restrictionsMaxSize), outerRestrictions: make([]string, restrictionsMaxSize)}
	ff.filterVars = make(map[string]interface{})
	ff.outerFilterVars = make(map[string]interface{})

	ff.filterVars["customerID"] = customerID
	ff.filterVars["limit"] = "10"
	ff.filterVars["offset"] = "0"
	ff.filterVars["sortColumn"] = "LastSeen"
	ff.filterVars["sortOrder"] = "ASC"

	return ff
}

//
// FilesJSON route
//
func (controller *Controller) FilesJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	// Read parameters
	dataTableParams := helpers.GetDatatableParams(r.URL.Query(), "LastSeen")
	dataTableParams.SortColumn = getExeSortColumn(dataTableParams.SortColumn)
	// TODO This filter fallback should be cleaner.  My goal is if the user types in unsupported characters, that I just show nothing.

	var ff = *NewFilterFiles(db, user.CustomerID)

	var filterString string
	filterString = ""
	if r.URL.Query()["filter"] != nil {
		filterString = r.URL.Query()["filter"][0]
	}

	if filterString != "" {
		type FilterElement struct {
			Category string `json:"category"`
			Operator string `json:"operator"`
			Value    string `json:"value"`
		}

		var filter []FilterElement
		if err := json.Unmarshal([]byte(filterString), &filter); err != nil {
			log.Warningf("Bad filter, could not be converted to json: %s", filterString)
		}

		// Apply filters
		if len(filter) != 0 {
			// TODO MUST check for sql injection
			for index, element := range filter {
				restrictionVar := fmt.Sprintf("restrictionVar%d", index)

				if element.Category == "ProductName" {
					ff.AddRestriction(fmt.Sprintf("productname %s :%s", utils.GetStringMatchOperator(element.Operator), restrictionVar), restrictionVar, element.Value)
				} else if element.Category == "CompanyName" {
					ff.AddRestriction(fmt.Sprintf("companyname %s :%s", utils.GetStringMatchOperator(element.Operator), restrictionVar), restrictionVar, element.Value)
				} else if element.Category == "Signer" {
					ff.AddOuterRestriction(fmt.Sprintf("fts.SubjectShortName %s :%s", utils.GetStringMatchOperator(element.Operator), restrictionVar), restrictionVar, element.Value)
				} else if element.Category == "Path" {
					value := element.Value
					if element.Operator == "contains" || element.Operator == "!contains" {
						value = fmt.Sprintf("%%%s%%", value)
					}
					ff.AddRestriction(fmt.Sprintf("FilePath %s :%s", utils.GetStringSubstrMatchOperator(element.Operator), restrictionVar), restrictionVar, value)
				} else if element.Category == "Count" {
					ff.AddOuterRestriction(fmt.Sprintf("NumSystems %s :%s", utils.GetIntMatchOperator(element.Operator), restrictionVar), restrictionVar, element.Value)
				} else if element.Category == "FirstSeen" {
					// TODO handle == and != (probably <= and => need to be fixed too)
					value := utils.ConvertYYYYMMDDtoUnix(element.Value)
					ff.AddDateRestriction("fsm.FirstSeen", element.Operator, value)
				} else if element.Category == "LastSeen" {
					value := utils.ConvertYYYYMMDDtoUnix(element.Value)
					ff.AddDateRestriction("fsm.LastSeen", element.Operator, value)
				}

			}
		}
	}

	sha256HexString := helpers.GetParam(r.URL.Query(), "sha256", "^[a-f0-9]{64}$", "")
	sha256, err := hex.DecodeString(sha256HexString)
	if err != nil {
		log.Errorf("Unable to decode Sha256: %v", err)
		return "", http.StatusBadRequest
	}

	if sha256HexString != "" {
		ff.AddRestriction("f.sha256 = :sha256", "sha256", sha256)
	}

	ff.SetFilterVar("limit", dataTableParams.Length)
	ff.SetFilterVar("offset", dataTableParams.Start)
	ff.SetFilterVar("sortColumn", dataTableParams.SortColumn)
	ff.SetFilterVar("sortOrder", dataTableParams.SortOrder)

	//
	// Get count
	//
	count, err := ff.GetCount()
	if err != nil {
		// TODO MUST This probably can happen if no files are in the DB
		log.Errorf("Unable to find files in DB, %v", err)
		return "", http.StatusBadRequest
	}

	files, err := ff.GetFilteredFiles()
	if err != nil {
		// TODO MUST This probably can happen if no files are in the DB
		log.Errorf("Unable to find files in DB, %v", err)
		return "", http.StatusBadRequest
	}

	type FileDataJSON struct {
		Sha256                 string
		FilePath               string
		FirstSeen              string
		LastSeen               string
		ProductName            string
		CompanyName            string
		NumSystems             int
		SignerSubjectShortName string
	}

	type DataTablesJSON struct {
		ITotalRecords        int            `json:"iTotalRecords"`
		ITotalDisplayRecords int            `json:"iTotalDisplayRecords"`
		SEcho                string         `json:"sEcho"`
		AaData               []FileDataJSON `json:"aaData"`
	}

	var dataTablesJSON DataTablesJSON
	dataTablesJSON.ITotalRecords = int(count)
	dataTablesJSON.ITotalDisplayRecords = len(files)
	dataTablesJSON.AaData = make([]FileDataJSON, len(files), len(files))

	for index, filedata := range files {
		var fileDataJSON FileDataJSON

		fileDataJSON.Sha256 = filedata.GetSha256()
		fileDataJSON.FilePath = filedata.FilePath
		fileDataJSON.FirstSeen = utils.Int64ToUnixTimeString(filedata.FirstSeen, true)
		fileDataJSON.LastSeen = utils.Int64ToUnixTimeString(filedata.LastSeen, true)
		fileDataJSON.ProductName = filedata.ProductName
		fileDataJSON.CompanyName = filedata.CompanyName
		fileDataJSON.NumSystems = filedata.NumSystems
		fileDataJSON.SignerSubjectShortName = filedata.GetSignerSubjectShortName()

		dataTablesJSON.AaData[index] = fileDataJSON
	}

	contents, err := json.Marshal(dataTablesJSON)
	if err != nil {
		log.Errorf("Unable to marshal json")
		return "", http.StatusBadRequest
	}

	return string(contents), http.StatusOK
}

//
// FileInfoJSON route
//
func (controller *Controller) FileInfoJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	sha256HexString := helpers.GetParam(r.URL.Query(), "sha256", "^[a-f0-9]{64}$", "")
	sha256, err := hex.DecodeString(sha256HexString)
	if err != nil {
		log.Errorf("Unable to decode Sha256: %v", err)
		return "", http.StatusBadRequest
	}

	var ff = *NewFilterFiles(db, user.CustomerID)

	// Apply filters
	if sha256HexString != "" {
		ff.AddRestriction("f.sha256 = :sha256", "sha256", sha256)
	}

	files, err := ff.GetFilteredFiles()
	if err != nil {
		// TODO MUST This probably can happen if no files are in the DB
		log.Errorf("Unable to find files in DB, %v", err)
		return "", http.StatusBadRequest
	}

	if len(files) > 1 {
		log.Errorf("BAD: More than one file found with this sha256. %v", err)
		return "", http.StatusBadRequest
	}

	if len(files) == 0 {
		log.Infof("Request made for non-existent sha256. %v", err)
		return "", http.StatusBadRequest
	}

	filteredfile := files[0]

	type DetailedFileData struct {
		Sha1             []byte
		Md5              []byte
		Size             int
		CompanyName      string
		ProductVersion   string
		ProductName      string
		FileDescription  string
		InternalName     string
		FileVersion      string
		OriginalFilename string

		Subject                   sql.NullString
		SerialNumber              []byte
		DigestAlgorithm           sql.NullString
		DigestEncryptionAlgorithm sql.NullString
	}

	var detailedFileData DetailedFileData

	err = ff.db.SelectOne(&detailedFileData, `SELECT
			Sha1,
			Md5,
			Size,
			CompanyName,
			ProductVersion,
			ProductName,
			FileDescription,
			InternalName,
			FileVersion,
			OriginalFilename,
			Subject,
			SerialNumber,
			DigestAlgorithm,
			DigestEncryptionAlgorithm
		FROM ExecutableFiles e
		LEFT JOIN (
				SELECT *
				FROM FileToSignerMap ftsmap, Signers s
				WHERE ftsmap.FileID=:fileID AND ftsmap.SignerID = s.ID
		) fts ON e.ID = fts.FileID
		WHERE e.id=:fileID`,
		map[string]interface{}{
			"fileID": filteredfile.FileID,
		})
	if err != nil {
		log.Errorf("BAD: Unable to find file in DB, %v", err)
		return "", http.StatusBadRequest
	}

	type FileDataJSON struct {
		Sha256 string
		Sha1   string
		Md5    string

		FilePath   string
		FirstSeen  string
		LastSeen   string
		NumSystems int

		Size             int
		CompanyName      string
		ProductVersion   string
		ProductName      string
		FileDescription  string
		InternalName     string
		FileVersion      string
		OriginalFilename string

		SubjectShortName          string
		Subject                   string
		SerialNumber              string
		DigestAlgorithm           string
		DigestEncryptionAlgorithm string
	}

	var fileDataJSON FileDataJSON
	fileDataJSON.Sha256 = filteredfile.GetSha256()
	fileDataJSON.Sha1 = hex.EncodeToString(detailedFileData.Sha1)
	fileDataJSON.Md5 = hex.EncodeToString(detailedFileData.Md5)

	fileDataJSON.FilePath = filteredfile.FilePath
	fileDataJSON.FirstSeen = utils.Int64ToUnixTimeString(filteredfile.FirstSeen, false)
	fileDataJSON.LastSeen = utils.Int64ToUnixTimeString(filteredfile.LastSeen, false)
	fileDataJSON.NumSystems = filteredfile.NumSystems

	fileDataJSON.Size = detailedFileData.Size
	fileDataJSON.CompanyName = detailedFileData.CompanyName
	fileDataJSON.ProductVersion = detailedFileData.ProductVersion
	fileDataJSON.ProductName = detailedFileData.ProductName
	fileDataJSON.FileDescription = detailedFileData.FileDescription
	fileDataJSON.InternalName = detailedFileData.InternalName
	fileDataJSON.FileVersion = detailedFileData.FileVersion
	fileDataJSON.OriginalFilename = detailedFileData.OriginalFilename

	fileDataJSON.SubjectShortName = filteredfile.GetSignerSubjectShortName()
	fileDataJSON.Subject = utils.GetNullString(detailedFileData.Subject, "")
	fileDataJSON.SerialNumber = hex.EncodeToString(detailedFileData.SerialNumber)
	fileDataJSON.DigestAlgorithm = utils.GetNullString(detailedFileData.DigestAlgorithm, "")
	fileDataJSON.DigestEncryptionAlgorithm = utils.GetNullString(detailedFileData.DigestEncryptionAlgorithm, "")

	contents, err := json.Marshal(fileDataJSON)
	if err != nil {
		log.Errorf("Unable to marshal json")
		return "", http.StatusBadRequest
	}

	return string(contents), http.StatusOK
}
