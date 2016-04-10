////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/exp/ses"

	"qdserver/lib/models"
)

// AwsSes is used in config.json files
type AwsSes struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	RegionURL string `json:"region_url"`
}

// PasswordResetEmailData is sent in password reset emails
type PasswordResetEmailData struct {
	DatabaseID int64
	Nonce      []byte
}

// SendPasswordResetEmail checks the email is a known user, and if so, send an email to them to allow them to change their password and be logged in
func SendPasswordResetEmail(db *gorp.DbMap, awsSes AwsSes, emailAddress string, baseURL string) (err error) {

	// Ensure user exists in the DB
	var user *models.User
	err = db.SelectOne(&user, "select * from users where Email=:Email",
		map[string]interface{}{
			"Email": emailAddress,
		})
	if err != nil {
		// User does not exist
		// TODO EVENTUALLY: If user does not exist, send an email to them anyway saying this email does not exist in our DB
		//   in case they gave you a different email than the one they signed up with
		log.Infof("Attempt to reset password for user with email that was not found in our DB: %s", emailAddress)
		return err
	}

	// TODO Check and set user.LastPasswordResetEmailDate

	log.Infof("Reset email being sent to %s for user %d", user.Email, user.ID)

	// Create passwordReset data for database
	var passwordReset models.PasswordReset

	nonceSize := 64
	nonce := make([]byte, nonceSize)
	n, err := rand.Read(nonce)
	if n != nonceSize || err != nil {
		log.Errorf("Problem obtaining random data")
		return err
	}

	passwordReset.Nonce = nonce
	passwordReset.CreationDate = DBTimeNow()
	passwordReset.Valid = true
	passwordReset.UserID = user.ID
	err = db.Insert(&passwordReset)
	if err != nil {
		log.Errorf("Failed to add passwordReset info to DB, %v", err)
		return err
	}

	// TODO Give the session a time limit
	// TODO Send an email after the password has been reset using this technique

	passwordResetEmailData := PasswordResetEmailData{
		DatabaseID: passwordReset.ID,
		Nonce:      nonce}
	// Convert struct to byte array
	var passwordResetBytes bytes.Buffer
	encoder := gob.NewEncoder(&passwordResetBytes)
	err = encoder.Encode(passwordResetEmailData)
	if err != nil {
		log.Errorf("Unable to encode data")
		return err
	}

	encryptedPasswordResetData, err := SymmetricEncrypt(KeyForObfuscation, passwordResetBytes.Bytes())
	if err != nil {
		log.Errorf("Unable to encrypt data")
		return err
	}

	// Convert string to base64 then replace weird characters so it works in an URL
	passwordResetString := base64.StdEncoding.EncodeToString(encryptedPasswordResetData)
	passwordResetString = strings.Replace(passwordResetString, "+", "-", -1)
	passwordResetString = strings.Replace(passwordResetString, "/", "_", -1)

	url := fmt.Sprintf("%s/password_reset/%s", baseURL, passwordResetString)
	body := fmt.Sprintf("You requested a password reset. Please visit this link to enter your new password:\n\n%s", url)
	response, err := sendMail(awsSes, "Summit Route <do_not_reply@summitroute.com>", emailAddress, "Password Reset: Summit Route", body)

	// Do something with this response
	log.Infof("Password reset email response: %s", response)

	return err
}

func sendMail(awsSes AwsSes, from string, to string, subject string, body string) (string, error) {
	auth := aws.Auth{AccessKey: awsSes.AccessKey, SecretKey: awsSes.SecretKey}
	sesService := ses.New(auth, aws.Region{SESEndpoint: awsSes.RegionURL})
	resp, err := sesService.SendEmail(from,
		ses.NewDestination([]string{to}, nil, nil),
		ses.NewMessage(subject, body, body))
	if err != nil || resp == nil {
		return "", err
	}
	return resp.SendEmailResult.MessageId, err
}
