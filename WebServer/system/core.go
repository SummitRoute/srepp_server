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
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq" // Needed for gorp
	"github.com/zenazn/goji/web"

	"qdserver/lib/models"
)

// Application holds our app's global variables
type Application struct {
	Configuration *Configuration
	Template      *template.Template
	Store         *sessions.CookieStore
	DBSession     *gorp.DbMap
}

const (
	// RouteProtected means the user has to have logged in in order to access this route
	RouteProtected = 1
	// RoutePublic means this route is publicly accessible (no need for a log in)
	RoutePublic = 0
)

// Init initializes our globals
func (application *Application) Init(filename *string) {
	application.Configuration = &Configuration{}

	// Load our configuration file
	err := application.Configuration.Load(*filename)
	if err != nil {
		log.Fatalf("Can't read configuration file: %s", err)
		panic(err)
	}

	// Set up cookie store
	application.Store = sessions.NewCookieStore([]byte(application.Configuration.Secret))
}

// LoadTemplates initializes the template engine
func (application *Application) LoadTemplates() error {
	var templates []string

	// Create function to collect our template files
	fn := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() != true && strings.HasSuffix(f.Name(), ".html") {
			templates = append(templates, path)
		}
		return nil
	}

	// Look for all the template files
	err := filepath.Walk(application.Configuration.TemplatePath, fn)
	if err != nil {
		return err
	}

	// Make sure we can parse all the template files
	application.Template = template.Must(template.ParseFiles(templates...))
	return nil
}

// ConnectToDatabase initializes our database connection
func (application *Application) ConnectToDatabase() {
	// Initialize the database
	dbmap, err := models.InitDB(application.Configuration.Database.ConnectionString)
	if err != nil {
		log.Fatalf("Unable to initialize the database: %v", err)
		panic(err)
	}

	// Store our session
	application.DBSession = dbmap
}

// Close cleans up anything nicely before exiting the process
func (application *Application) Close() {
	log.Info("Bye!")
	application.DBSession.Db.Close()
}

// Route defines a web route
func (application *Application) Route(controller interface{}, route string, protected int) interface{} {
	fn := func(c web.C, w http.ResponseWriter, r *http.Request) {
		c.Env["Content-Type"] = "text/html"

		if protected == RouteProtected && c.Env["User"] == nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		methodValue := reflect.ValueOf(controller).MethodByName(route)
		methodInterface := methodValue.Interface()
		method := methodInterface.(func(c web.C, r *http.Request) (string, int))

		body, code := method(c, r)

		if session, exists := c.Env["Session"]; exists {
			err := session.(*sessions.Session).Save(r, w)
			if err != nil {
				log.Errorf("Can't save session: %v", err)
			}
		}

		switch code {
		case http.StatusOK:
			if _, exists := c.Env["Content-Type"]; exists {
				w.Header().Set("Content-Type", c.Env["Content-Type"].(string))
			}
			if _, exists := c.Env["Content-Length"]; exists {
				w.Header().Set("Content-Length", c.Env["Content-Length"].(string))
			}
			io.WriteString(w, body)
		case http.StatusSeeOther, http.StatusFound:
			http.Redirect(w, r, body, code)
		default:
			w.WriteHeader(code)
			io.WriteString(w, body)
		}
	}
	return fn
}
