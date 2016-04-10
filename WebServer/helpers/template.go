////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package helpers

import (
	"bytes"
	"html/template"
	"net/url"
	"regexp"
	"strconv"
)

// Parse parses the template file
func Parse(t *template.Template, name string, data interface{}) string {
	var doc bytes.Buffer
	t.ExecuteTemplate(&doc, name, data)
	return doc.String()
}

// GetParam looks for a given parameter
// urlValues should be set to r.URL.Query()
// param is the name of the parameter
// regexMatcher is a regex string to ensure the string matches against, or "" for no regex
func GetParam(urlValues url.Values, param string, regexMatcher string, defaultStr string) string {
	if urlValues[param] == nil {
		return defaultStr
	}
	response := urlValues[param][0]

	// Check if we have a regex to match against
	if regexMatcher == "" {
		// No pattern to match against so return what we have
		return response
	}

	match, _ := regexp.MatchString(regexMatcher, response)
	if !match {
		// Regex didn't match!
		return defaultStr
	}

	// Regex matched, return it
	return response
}

// DatatableParams sanitized parameters for SQL statements for returning info to datatables
type DatatableParams struct {
	Start      string
	Length     string
	SortColumn string
	SortOrder  string
}

// GetDatatableParams given query parameters, it finds the values relevant to the datatable and sets the response struct with those values
func GetDatatableParams(urlValues url.Values, defaultSortColumn string) DatatableParams {
	var response DatatableParams

	response.Start = GetParam(urlValues, "start", "^[0-9]*$", "0")
	response.Length = GetParam(urlValues, "length", "^[0-9]*$", "25")
	ilength, err := strconv.Atoi(response.Length)
	if err != nil || ilength > 100 {
		response.Length = "100"
	}

	response.SortColumn = GetParam(urlValues, "sort", "^[a-zA-Z 0-9]*$", defaultSortColumn)
	if response.SortColumn == "" {
		response.SortColumn = defaultSortColumn
	}

	sortAscending := GetParam(urlValues, "sa", "^[a-z]*$", "true")

	if sortAscending == "true" {
		response.SortOrder = "ASC"
	} else {
		response.SortOrder = "DESC"
	}

	return response
}
