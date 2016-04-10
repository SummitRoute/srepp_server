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
	ExeUpload     utils.AwsS3 `json:"exe_upload"`
	CatalogUpload utils.AwsS3 `json:"catalog_upload"`
}

// Configuration is the main structure of our config.json file
type Configuration struct {
	ListeningPort string                `json:"listening_port"`
	UploadPath    string                `json:"upload_path"`
	AMQPServer    string                `json:"amqp_server"`
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

// Parse our configuration file
func (configuration *Configuration) Parse(data []byte) (err error) {
	err = json.Unmarshal(data, &configuration)

	return
}
