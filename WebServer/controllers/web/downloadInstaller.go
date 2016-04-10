////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	log "github.com/Sirupsen/logrus"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/zenazn/goji/web"

	"qdserver/WebServer/helpers"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// DownloadInstaller route
func (controller *Controller) DownloadInstaller(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "/signin", http.StatusSeeOther
	}

	// Get the Customer UUID
	var customer models.Customer
	err := db.SelectOne(&customer, "select * from customers where ID=:id",
		map[string]interface{}{
			"id": user.CustomerID,
		})
	if err != nil {
		log.Errorf("Unable to find customer in DB")
		return "/signin", http.StatusSeeOther
	}

	CustomerUUID, err := uuid.Parse(customer.UUID)
	if err != nil {
		log.Errorf("Unable to parse UUID from DB, %v", err)
		return "/signin", http.StatusSeeOther
	}

	// Check for personalized installer
	// TODO Need to put this version number somewhere configurable
	version := "0.0.0.1"
	guid := CustomerUUID.String()
	installerName := fmt.Sprintf("srepp%s-%s.exe", version, guid)
	installerPath := path.Join("downloads", installerName)

	// TODO this is racy
	if !utils.CheckExists(installerPath) {
		log.Infof("Installer does not exist so creating it")

		if err := helpers.ConfigureInstaller(path.Join("downloads", "srepp_installer.exe"), installerPath, guid); err != nil {
			log.Errorf("Unable to configure installer, %v", err)
			return "/signin", http.StatusSeeOther
		}

		// TODO need to periodically delete these installers
	}

	c.Env["Content-Type"] = "application/octet-stream"
	contents, err := ioutil.ReadFile(installerPath)
	if err != nil {
		log.Errorf("Unable to read file, %v", err)
		return "/signin", http.StatusSeeOther
	}

	c.Env["Content-Length"] = fmt.Sprintf("%d", len(contents))

	return string(contents), http.StatusOK
}
