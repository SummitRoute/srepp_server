////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package web

import "github.com/gorilla/sessions"

// AddFlash adds a flash message to a page
func AddFlash(session *sessions.Session, t string, msg string) {
	// TODO MUST Need to be able to show errors or success messages
	session.AddFlash(msg, "auth")
}
