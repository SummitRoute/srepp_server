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
	"code.google.com/p/go.crypto/bcrypt"
	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"

	"qdserver/lib/models"
)

// Login attempts to login with the provided credentials.
// On success, returns the user model, else nil and the error
func Login(db *gorp.DbMap, email string, password string) (user *models.User, err error) {
	err = db.SelectOne(&user, "select * from users where Email=:Email",
		map[string]interface{}{
			"Email": email,
		})
	if err != nil {
		log.Errorf("Failed to find user")
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		log.Info("Password hash did not match")
		return nil, err
	}
	return user, nil
}
