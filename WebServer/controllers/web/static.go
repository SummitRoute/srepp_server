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
	"html/template"
	"net/http"

	"github.com/zenazn/goji/web"
)

// Help shows the help page
func (controller *Controller) Help(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)

	var widgets = controller.Parse(t, "help", c.Env)

	c.Env["Title"] = "Summit Route - Help"
	c.Env["Content"] = template.HTML(widgets)

	return controller.Parse(t, "main", c.Env), http.StatusOK
}

// TermsAndConditions shows the terms and conditions page
func (controller *Controller) TermsAndConditions(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)

	var widgets = controller.Parse(t, "terms_and_conditions", c.Env)

	c.Env["Title"] = "Summit Route - Terms and Conditions"
	c.Env["Content"] = template.HTML(widgets)

	return controller.Parse(t, "main", c.Env), http.StatusOK
}

// Privacy shows privacy policy
func (controller *Controller) Privacy(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)

	var widgets = controller.Parse(t, "privacy", c.Env)

	c.Env["Title"] = "Summit Route - Privacy Policy"
	c.Env["Content"] = template.HTML(widgets)

	return controller.Parse(t, "main", c.Env), http.StatusOK
}
