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

	"code.google.com/p/go.crypto/bcrypt"
	log "github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web"

	"qdserver/lib/models"
)

// ProfileJSON route
func (controller *Controller) ProfileJSON(c web.C, r *http.Request) (string, int) {
	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	type UserJSON struct {
		FirstName    string
		LastName     string
		Email        string
		CreationDate int64
		LastLogin    int64
	}

	var userJSON UserJSON

	userJSON.FirstName = user.FirstName
	userJSON.LastName = user.LastName
	userJSON.Email = user.Email
	userJSON.Email = user.Email
	userJSON.LastLogin = user.LastLogin

	contents, err := json.Marshal(userJSON)
	if err != nil {
		log.Errorf("Unable to marshal json")
		return "", http.StatusBadRequest
	}

	return string(contents), http.StatusOK
}

// PostProfileJSON route
func (controller *Controller) PostProfileJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	firstname := r.FormValue("FirstName")
	lastname := r.FormValue("LastName")
	email := r.FormValue("Email")

	// TODO Should sanity check the incoming data
	// Sanity check
	if len(email) <= 3 {
		return "Email too short", http.StatusBadRequest
	}

	// Ensure email address is not already in the DB
	count, err := db.SelectInt(`SELECT count(*)
			FROM Users
			WHERE Email=:email and ID!=:id`,
		map[string]interface{}{
			"email": email,
			"id":    user.ID,
		})
	if err != nil {
		log.Errorf("Error searching for user with email %s, %v", email, err)
		return "", http.StatusBadRequest
	}

	if count != 0 {
		log.Infof("User %d tried to set email address to an existing accont's email %v", user.ID, err)
		return "email not unique", http.StatusBadRequest
	}

	// Set the database user to use our new data
	user.FirstName = firstname
	user.LastName = lastname
	user.Email = email

	// Update it in the DB
	_, err = db.Update(&user)
	if err != nil {
		log.Warningf("Can't update user: %v", err)
		return "", http.StatusBadRequest
	}

	return "", http.StatusOK
}

// PostChangePasswordJSON route
func (controller *Controller) PostChangePasswordJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	NewPassword := r.FormValue("NewPassword")
	CurrentPassword := r.FormValue("CurrentPassword")

	// Check old password matches
	err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(CurrentPassword))
	if err != nil {
		log.Info("Password hash did not match")
		return "old password incorrect", http.StatusBadRequest
	}

	// Set our password hash
	err = user.HashPassword(NewPassword)
	if err != nil {
		log.Errorf("Can't hash password: %v", err)
		return "", http.StatusBadRequest
	}

	// Update it in the DB
	_, err = db.Update(&user)
	if err != nil {
		log.Warningf("Can't update user: %v", err)
		return "", http.StatusBadRequest
	}

	return "", http.StatusOK
}

// PostResetPasswordJSON route
func (controller *Controller) PostResetPasswordJSON(c web.C, r *http.Request) (string, int) {
	db := controller.GetDatabase(c)

	log.Infof("Reset password") // TODO REMOVE

	// Get our user object
	var user models.User
	user, ok := c.Env["User"].(models.User)
	if !ok {
		// This should not happen
		log.Errorf("User was not of the expected form")
		return "", http.StatusBadRequest
	}

	log.Infof("user obtained") // TODO REMOVE

	NewPassword := r.FormValue("NewPassword")

	// Set our password hash
	err := user.HashPassword(NewPassword)
	if err != nil {
		log.Errorf("Can't hash password: %v", err)
		return "", http.StatusBadRequest
	}

	// Update it in the DB
	_, err = db.Update(&user)
	if err != nil {
		log.Warningf("Can't update user: %v", err)
		return "", http.StatusBadRequest
	}

	return "", http.StatusOK
}
