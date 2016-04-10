////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package api

import (
	"net/http"

	"github.com/zenazn/goji/web"

	"qdserver/WebServer/system"
)

// Controller blah
type Controller struct {
	system.Controller
}

// TermsAPI route
func (controller *Controller) TermsAPI(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)
	return controller.Parse(t, "terms_and_conditions", c.Env), http.StatusOK
}

// PrivacyAPI route
func (controller *Controller) PrivacyAPI(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)
	return controller.Parse(t, "privacy", c.Env), http.StatusOK
}

// HelpAPI route
func (controller *Controller) HelpAPI(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)
	return controller.Parse(t, "help", c.Env), http.StatusOK
}
