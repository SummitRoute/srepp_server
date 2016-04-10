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

	"github.com/justinas/nosurf"
	"github.com/zenazn/goji/web"
)

// Index route shows the home page
func (controller *Controller) Index(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)

	c.Env["csrf_token"] = nosurf.Token(r)

	return controller.Parse(t, "index", c.Env), http.StatusOK
}
