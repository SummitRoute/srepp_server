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
	"bytes"
	"html/template"

	"github.com/coopernurse/gorp"
	"github.com/gorilla/sessions"
	"github.com/zenazn/goji/web"
)

// Controller blah
type Controller struct {
}

// GetSession blah
func (controller *Controller) GetSession(c web.C) *sessions.Session {
	return c.Env["Session"].(*sessions.Session)
}

// GetTemplate blah
func (controller *Controller) GetTemplate(c web.C) *template.Template {
	return c.Env["Template"].(*template.Template)
}

// GetDatabase blah
func (controller *Controller) GetDatabase(c web.C) *gorp.DbMap {
	if db, ok := c.Env["DBSession"].(*gorp.DbMap); ok {
		return db
	}

	return nil
}

// Parse blah
func (controller *Controller) Parse(t *template.Template, name string, data interface{}) string {
	var doc bytes.Buffer
	t.ExecuteTemplate(&doc, name, data)
	return doc.String()
}
