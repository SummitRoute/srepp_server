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
	"net/http"

	"github.com/zenazn/goji/web"
)

// Logout logs the user out
func (controller *Controller) Logout(c web.C, r *http.Request) (string, int) {
	session := controller.GetSession(c)

	session.Values["SessionID"] = nil
	session.Values["SessionNonce"] = nil

	return "/", http.StatusSeeOther
}
