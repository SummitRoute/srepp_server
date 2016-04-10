////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"
)

// Customer is a collection of systems.  Likely an entire business.
// When a new customer is creater
type Customer struct {
	ID   int64  // Internal ID for the DB
	UUID []byte // Customer ID the client knows

	Active       bool // In case we want to disable the customer
	CreationDate int64
}

// User of the webapp
//
// Example: insert into users (customerid, firstname, lastname, email, passwordhash, verified, active, mustsetpassword, creationdate, lastlogin) values (1, 'scott', 'piper', 'scott@summitroute.com', '\x24326124313024355566457671715563332f79436653447a32646a434f477278434b466c5861536f476b4f4e465a353141774c55376342345a777a2e', TRUE, TRUE, FALSE,  1414698750, 0);
//   Password is "abc"
type User struct {
	ID           int64
	CustomerID   int64
	FirstName    string
	LastName     string
	Email        string // Used for both login and contacting
	PasswordHash []byte // PasswordHash (bcrypt hash, so it includes a salt)

	Verified                   bool  // True when the user has verified their email
	Active                     bool  // In case we want to disable the user
	MustSetPassword            bool  // True when the user needs to set their password (from password resets)
	LastPasswordResetEmailDate int64 // Last time we sent them a password reset email, so we don't flood them
	CreationDate               int64
	LastLogin                  int64
}

// BrowserSession keeps track of browser sessions where a user has logged in
// TODO ONEDAY I tried researching best practices regarding what to keep in the cookie,
// and what to store in the DB, and couldn't find anything.  So I've taken a best guess here at what to do.
type BrowserSession struct {
	ID        int64  // TODO Need to use random, non-repeating numbers for the ID
	NonceHash []byte // Random value to avoid session hijacking, since the session ID really should be at least 128-bits
	UserID    int64  // Ties this back to the user that logged in

	CreationDate int64  // For expiration purposes
	LastActive   int64  // Need to keep track of this so we don't log the user out while they are doing something
	IP           []byte // Keep track of where this was originally logged in from
	UserAgent    []byte // Keep track of what browser was used to login with
}

// PasswordReset Used by password reset emails to re-login a user
type PasswordReset struct {
	ID     int64
	Nonce  []byte // Random value 256-bit value
	UserID int64  // Ties this back to the user that logged in

	CreationDate int64 // For expiration purposes
	Valid        bool  // Set to zero after first use so this can't be re-used
}

// HashPassword generates a bcrypt hash (includes salt) of the password
func (user *User) HashPassword(password string) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("Couldn't hash password: %v", err)
		return err
	}
	user.PasswordHash = hash
	return nil
}

// GetUserByEmail looks up the user in the database by their email
func GetUserByEmail(db *gorp.DbMap, email string) (user *User) {
	db.SelectOne(&user, "select * from users where Email=:email",
		map[string]interface{}{
			"email": email,
		})
	// No error checking because it throws errors when there are no matches
	/*
		if err != nil {
			log.Warningf("Can't get user by email: %v", err)
		}
	*/
	return
}

// InsertUser adds a user to the DB
func InsertUser(db *gorp.DbMap, user *User) error {
	return db.Insert(user)
}

// InsertSession adds a broser session to the DB
func InsertSession(db *gorp.DbMap, browserSession *BrowserSession) error {
	return db.Insert(browserSession)
}
