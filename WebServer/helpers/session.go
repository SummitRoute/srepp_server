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
	"crypto/rand"
	"crypto/sha256"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"

	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// GetIP finds the IP address of the request
func GetIP(r *http.Request) []byte {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return []byte(ipProxy)
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return []byte(ip)
}

// CreateBrowserSession creates a browser session object and stores it in the DB
func CreateBrowserSession(database *gorp.DbMap, r *http.Request, user models.User) (SessionID int64, SessionNonce []byte, err error) {
	// Create a random hash
	nonce := make([]byte, 32)
	n, err := rand.Read(nonce)
	if n != 32 || err != nil {
		log.Errorf("Problem obtaining random data")
	}

	// Get a hash of the nonce
	hasher := sha256.New()
	hasher.Write(nonce)
	sessionNonceHash := hasher.Sum(nil)

	// Create the DB object
	browserSession := &models.BrowserSession{
		NonceHash:    sessionNonceHash,
		UserID:       user.ID,
		CreationDate: utils.DBTimeNow(),
		LastActive:   utils.DBTimeNow(),
		IP:           GetIP(r),
		UserAgent:    []byte(r.UserAgent()),
	}

	// Save it to the DB
	if err := models.InsertSession(database, browserSession); err != nil {
		log.Errorf("Error while adding session to DB: %v", err)
		return 0, nil, err
	}

	// Return
	SessionID = browserSession.ID
	SessionNonce = nonce

	return SessionID, SessionNonce, nil
}
