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
	"net/http"

	"github.com/zenazn/goji/web"
)

// ApplyConfig makes sure controllers can have access to the database and other variables
func (application *Application) ApplyConfig(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c.Env["DBSession"] = application.DBSession
		c.Env["Config"] = application.Configuration
		c.Env["QueueChannel"] = application.QueueChannel

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
