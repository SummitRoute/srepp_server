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
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/zenazn/goji/web"

	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// ApplyTemplates makes sure templates are stored in the context
func (application *Application) ApplyTemplates(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c.Env["Template"] = application.Template
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ApplySessions makes sure controllers can have access to session
func (application *Application) ApplySessions(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session, _ := application.Store.Get(r, "session")
		c.Env["Session"] = session
		h.ServeHTTP(w, r)
		context.Clear(r)
	}
	return http.HandlerFunc(fn)
}

// ApplyDatabase makes sure controllers can have access to the database
func (application *Application) ApplyDatabase(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c.Env["DBSession"] = application.DBSession
		c.Env["Config"] = application.Configuration

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ApplyAuth makes sure controllers can check if the user is authorized
func (application *Application) ApplyAuth(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Get session ID and nonce from the cookie
		session := c.Env["Session"].(*sessions.Session)
		if sessionID, ok := session.Values["SessionID"]; ok {
			if sessionNonce, ok := session.Values["SessionNonce"]; ok {
				if sessionID == nil || sessionNonce == nil {
					// No session values so get out of here
					c.Env["User"] = nil
					session.Values["SessionID"] = nil
					session.Values["SessionNonce"] = nil
				} else {

					// Sanity check it is the correct form
					sessionNonceBytes, ok := sessionNonce.([]byte)
					if !ok {
						log.Warningf("Session nonce was not a byte array... that should not happen")
						c.Env["User"] = nil
						session.Values["SessionID"] = nil
						session.Values["SessionNonce"] = nil
					} else {

						hasher := sha256.New()
						hasher.Write(sessionNonceBytes)
						sessionNonceHash := hasher.Sum(nil)

						// Look for the session ID in the database
						db := application.DBSession
						var browserSession models.BrowserSession
						err := db.SelectOne(&browserSession, "select * from browserSessions where ID=:id",
							map[string]interface{}{
								"id": sessionID,
							})
						if err != nil {
							log.Warningf("Session values not found")
							c.Env["User"] = nil
							session.Values["SessionID"] = nil
							session.Values["SessionNonce"] = nil
						} else {

							// Check our nonce matches
							if subtle.ConstantTimeCompare(browserSession.NonceHash, sessionNonceHash) != 1 {
								log.Errorf("Nonce does not match!  Possible attack!?")
								c.Env["User"] = nil
								session.Values["SessionID"] = nil
								session.Values["SessionNonce"] = nil
							} else {

								// Ensure this session is not stale
								var timeout int64 = 86400 // Seconds in a day
								if utils.DBTimeNow()-browserSession.CreationDate > timeout {
									log.Warningf("Session expired")
									c.Env["User"] = nil
									session.Values["SessionID"] = nil
									session.Values["SessionNonce"] = nil
								} else {

									// Valid browser session, so look for the user
									var user models.User
									err = db.SelectOne(&user, "select * from users where ID=:id",
										map[string]interface{}{
											"id": browserSession.UserID,
										})
									if err != nil {
										log.Warningf("Problem finding the user: %v", err)
										c.Env["User"] = nil
										session.Values["SessionID"] = nil
										session.Values["SessionNonce"] = nil
									} else {
										c.Env["User"] = user
									}
								}
							}
						}
					}
				}
			}
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ApplyProtectionFromCSRF makes all POST messages check for a csrf_token
func (application *Application) ApplyProtectionFromCSRF(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		protected := nosurf.New(h)

		failureHandler := func(w http.ResponseWriter, r *http.Request) {
			log.Errorf("Possible CSRF attack")
			w.Write([]byte("400: Request could not be handled"))
			w.WriteHeader(400)
		}

		protected.SetFailureHandler(http.HandlerFunc(failureHandler))
		protected.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
