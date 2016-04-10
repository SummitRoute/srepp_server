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

	log "github.com/Sirupsen/logrus"
	"github.com/justinas/nosurf"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/zenazn/goji/web"

	"qdserver/WebServer/helpers"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

// Register route
func (controller *Controller) Register(c web.C, r *http.Request) (string, int) {
	t := controller.GetTemplate(c)
	session := controller.GetSession(c)

	// With that kind of flags template can "figure out" what route is being rendered
	c.Env["IsRegister"] = true

	c.Env["Flash"] = session.Flashes("auth")
	c.Env["csrf_token"] = nosurf.Token(r)
	var widgets = controller.Parse(t, "auth/register", c.Env)

	c.Env["Title"] = "Summit Route - Register"
	c.Env["Content"] = template.HTML(widgets)

	return controller.Parse(t, "main", c.Env), http.StatusOK
}

// RegisterPost registers new user or shows Register route with appropriate messages set in session
func (controller *Controller) RegisterPost(c web.C, r *http.Request) (string, int) {
	// TODO Need to validate these
	firstname, lastname := r.FormValue("firstname"), r.FormValue("lastname")
	email, password := r.FormValue("email"), r.FormValue("password")

	session := controller.GetSession(c)
	database := controller.GetDatabase(c)

	// Check if this user already exists (based on email)
	user := models.GetUserByEmail(database, email)
	if user != nil {
		session.AddFlash("A user with that email address already exists", "auth")
		return controller.Register(c, r)
	}

	// Create Customer UUID
	CustomerUUID, err := uuid.NewV4()
	if err != nil {
		// This is bad
		log.Errorf("Problem getting a new uuid, %v", err)
		session.AddFlash("The server is experiencing a problem, try again later", "auth")
		return controller.Register(c, r)
	}

	// Convert UUID to slice for DB
	var CustomerUUIDslice []byte
	CustomerUUIDslice = CustomerUUID[0:16]

	// Create customer DB object
	customer := &models.Customer{
		UUID:         CustomerUUIDslice,
		Active:       true,
		CreationDate: utils.DBTimeNow(),
	}
	// TODO Check for collisions on the UUID

	// Save it
	if err = database.Insert(customer); err != nil {
		session.AddFlash("Error while registering user.")
		log.Errorf("Error while creating customer: %v", err)
		return controller.Register(c, r)
	}

	// Create user DB object
	user = &models.User{
		CustomerID:   customer.ID,
		FirstName:    firstname,
		LastName:     lastname,
		Email:        email,
		Active:       true,
		CreationDate: utils.DBTimeNow(),
	}
	err = user.HashPassword(password)
	if err != nil {
		session.AddFlash("Error while registering user.")
		log.Errorf("Can't hash password: %v", err)
		return controller.Register(c, r)
	}

	// Save it
	if err = database.Insert(user); err != nil {
		session.AddFlash("Error while registering user.")
		log.Errorf("Error while registering user: %v", err)
		return controller.Register(c, r)
	}

	// Create browser session (so the user is logged in)
	SessionID, SessionNonce, err := helpers.CreateBrowserSession(database, r, *user)
	if err != nil {
		session.AddFlash("Unexpected error, please come back later", "auth")
		return controller.Register(c, r)
	}

	session.Values["SessionID"] = SessionID
	session.Values["SessionNonce"] = SessionNonce
	// Values automatically get saved due to a wrapper

	return "/", http.StatusSeeOther
}
