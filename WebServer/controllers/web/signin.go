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
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"html/template"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/justinas/nosurf"
	"github.com/zenazn/goji/web"

	"qdserver/WebServer/helpers"
	"qdserver/WebServer/system"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// SignIn is used for displaying the sign-in page
func (controller *Controller) SignIn(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)
	session := controller.GetSession(c)

	c.Env["IsSignIn"] = true
	c.Env["Flash"] = session.Flashes("auth")
	c.Env["csrf_token"] = nosurf.Token(r)
	var widgets = controller.Parse(t, "auth/signin", c.Env)

	c.Env["Title"] = "Summit Route - Sign In"
	c.Env["Content"] = template.HTML(widgets)

	return controller.Parse(t, "main", c.Env), http.StatusOK
}

// SignInPost is used to log the user in or set appropriate message in session if login was not succesful
func (controller *Controller) SignInPost(c web.C, r *http.Request) (string, int) {
	// TODO need to do form validation to ensure these values exist
	email, password := r.FormValue("email"), r.FormValue("password")

	session := controller.GetSession(c)
	database := controller.GetDatabase(c)

	user, err := helpers.Login(database, email, password)
	if err != nil {
		session.AddFlash("Invalid Email or Password", "auth")
		return controller.SignIn(c, r)
	}

	SessionID, SessionNonce, err := helpers.CreateBrowserSession(database, r, *user)
	if err != nil {
		session.AddFlash("Unexpected error, please come back later", "auth")
		return controller.SignIn(c, r)
	}

	session.Values["SessionID"] = SessionID
	session.Values["SessionNonce"] = SessionNonce

	return "/", http.StatusSeeOther
}

// ForgotPassword route
func (controller *Controller) ForgotPassword(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)
	session := controller.GetSession(c)

	c.Env["Flash"] = session.Flashes("auth")
	c.Env["csrf_token"] = nosurf.Token(r)
	var widgets = controller.Parse(t, "auth/forgot_password", c.Env)

	c.Env["Title"] = "Summit Route - Forgot password"
	c.Env["Content"] = template.HTML(widgets)

	return controller.Parse(t, "main", c.Env), http.StatusOK
}

// ForgotPasswordPost route
func (controller *Controller) ForgotPasswordPost(c web.C, r *http.Request) (string, int) {
	email := r.FormValue("email")
	email = strings.ToLower(email)
	log.Infof("Password reset for user with email: %s", email)

	// TODO Should: Make this a task for a worker so we aren't waiting for Amazon's server response
	database := controller.GetDatabase(c)

	// TODO EVENTUALLY: Require a captcha to send this

	err := utils.SendPasswordResetEmail(database, c.Env["Config"].(*system.Configuration).Aws.Ses, email, c.Env["Config"].(*system.Configuration).BaseURL)
	if err != nil {
		log.Errorf("Error sending password reset email: %v", err)
	}

	// No matter what email they gave, we'll just send a "success" back to the user
	t := controller.GetTemplate(c)

	c.Env["Email"] = email
	var widgets = controller.Parse(t, "auth/forgot_password_post", c.Env)

	c.Env["Title"] = "Summit Route - Forgot password"
	c.Env["Content"] = template.HTML(widgets)

	return controller.Parse(t, "main", c.Env), http.StatusOK
}

// PasswordReset is called after a user clicks the link from their password reset email
// It checks the token from the URL and if it's all legit, then it logs the user in and makes them reset their password
func (controller *Controller) PasswordReset(c web.C, r *http.Request) (string, int) {
	session := controller.GetSession(c)
	db := controller.GetDatabase(c)

	data := c.URLParams["data"]

	// Convert string to base64 then replace weird characters so it works in an URL
	data = strings.Replace(data, "-", "+", -1)
	data = strings.Replace(data, "_", "/", -1)
	passwordResetEncryptedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Errorf("Can't decode base64 password reset string")
		// TODO Should inform the user in case they've copied in the link wrong or their email client broke it
		return "/", http.StatusSeeOther
	}

	// Decrypt it
	passwordResetData, err := utils.SymmetricDecrypt(utils.KeyForObfuscation, passwordResetEncryptedBytes)
	if err != nil {
		log.Errorf("Unable to encrypt password reset data")
		return "/", http.StatusSeeOther
	}

	// Unmarshal it
	var passwordResetEmailData utils.PasswordResetEmailData
	buffer := bytes.NewBuffer(passwordResetData)
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&passwordResetEmailData)
	if err != nil {
		log.Errorf("Unable to decode data to struct, %v", err)
		return "/", http.StatusSeeOther
	}

	//
	// Decoded data, now check that it's legit
	//

	log.Infof("ID: %d", passwordResetEmailData.DatabaseID)

	var passwordReset models.PasswordReset
	err = db.SelectOne(&passwordReset, `SELECT *
		FROM passwordresets
		WHERE ID=:ID and Nonce=:Nonce`,
		map[string]interface{}{
			"ID":    passwordResetEmailData.DatabaseID,
			"Nonce": passwordResetEmailData.Nonce,
		})
	if err != nil {
		log.Errorf("Password reset data not found in DB.  Possible hacking attempt.")
		return "/", http.StatusSeeOther
	}

	log.Infof("User: %d", passwordReset.UserID)

	if !passwordReset.Valid {
		log.Warningf("Password reset credentials are no longer valid")
		// TODO I Should warn the user about this
		return "/", http.StatusSeeOther
	}

	var timeout int64 = 7200 // Seconds in 2 hours
	if time.Now().UTC().Unix()-passwordReset.CreationDate > timeout {
		log.Warningf("Password reset credentials have expired")
		// TODO I should warn the user about this
		return "/", http.StatusSeeOther
	}

	//
	// Data is legit so generate a session for the user
	//

	// Find the user
	var user models.User
	err = db.SelectOne(&user, `SELECT *
		FROM users
		WHERE ID=:ID`,
		map[string]interface{}{
			"ID": passwordReset.UserID,
		})
	if err != nil {
		log.Errorf("BAD: User not found in DB for the password reset.")
		return "/", http.StatusSeeOther
	}

	// Update the user so we have to reset our password
	user.MustSetPassword = true
	_, err = db.Update(&user)
	if err != nil {
		log.Warningf("Can't update user: %v", err)
		return "/", http.StatusSeeOther
	}

	// Update passwordReset in the DB
	passwordReset.Valid = false
	_, err = db.Update(&passwordReset)
	if err != nil {
		log.Warningf("Can't update passwordReset: %v", err)
		return "/", http.StatusSeeOther
	}

	// Generate new session as if we've logged in
	SessionID, SessionNonce, err := helpers.CreateBrowserSession(db, r, user)
	if err != nil {
		log.Errorf("Unable to create the browser session")
		// TODO Should inform user
		return controller.SignIn(c, r)
	}

	session.Values["SessionID"] = SessionID
	session.Values["SessionNonce"] = SessionNonce

	// Logged in so send the user to the password reset screen
	return "/password_reset", http.StatusSeeOther
}
