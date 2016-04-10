////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package system

import (
	"encoding/json"
	"io/ioutil"

	"qdserver/lib/utils"
)

// ConfigurationDatabase is a sub-element of Configuration
type ConfigurationDatabase struct {
	ConnectionString string `json:"connection_string"`
}

// ConfigurationAWS is a sub-element of Configuration
type ConfigurationAWS struct {
	Ses utils.AwsSes `json:"ses"`
}

// Configuration is the main structure of our config.json file
type Configuration struct {
	Environment   string                `json:"environment"`
	ListeningPort string                `json:"listening_port"`
	Secret        string                `json:"secret"`        // Secret value for crypto of cookies.  Change this periodically.
	PublicPath    string                `json:"public_path"`   // Public web files
	TemplatePath  string                `json:"template_path"` // Template directory (the views)
	BaseURL       string                `json:"base_url"`      // In production this is "https://app.summitroute.com"
	Database      ConfigurationDatabase `json:"database"`
	Aws           ConfigurationAWS      `json:"aws"`
}

// Load parses our configuration file
func (configuration *Configuration) Load(filename string) (err error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return
	}

	err = configuration.Parse(data)

	return
}

// Parse parses the configuration file into a structure
func (configuration *Configuration) Parse(data []byte) (err error) {
	err = json.Unmarshal(data, &configuration)

	return
}
