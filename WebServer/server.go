////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/goji/glogrus"
	"github.com/unrolled/secure"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	gojiweb "github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"qdserver/WebServer/controllers/api"
	"qdserver/WebServer/controllers/web"
	"qdserver/WebServer/system"
)

// Code for this project is based on the MIT licensed: https://github.com/elcct/defaultproject/

func main() {
	configfile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	var application = &system.Application{}

	application.Init(configfile)
	application.LoadTemplates()
	application.ConnectToDatabase()

	// Setup static files
	static := gojiweb.New()
	static.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(application.Configuration.PublicPath))))

	http.Handle("/assets/", static)

	//
	// Figure out log type to use (if debug, use text, else use json)
	//
	type LogInterface func() log.Formatter
	var getLogger LogInterface

	getLogger = func() log.Formatter { return new(log.JSONFormatter) }

	if application.Configuration.Environment == "debug" {
		getLogger = func() log.Formatter { return new(log.TextFormatter) }
	}

	//
	// Apply middleware
	//

	// User defined logs
	log.SetFormatter(getLogger())

	// Goji logs
	goji.Abandon(middleware.Logger)
	logr := log.New()
	logr.Formatter = getLogger()
	appName := "webserver"
	goji.Use(glogrus.NewGlogrus(logr, appName))

	secureMiddleware := secure.New(secure.Options{
		STSSeconds:              315360000,            // STSSeconds is the max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.
		STSIncludeSubdomains:    true,                 // If STSIncludeSubdomains is set to true, the `includeSubdomains` will be appended to the Strict-Transport-Security header. Default is false.
		FrameDeny:               true,                 // If FrameDeny is set to true, adds the X-Frame-Options header with the value of `DENY`. Default is false.
		CustomFrameOptionsValue: "SAMEORIGIN",         // CustomFrameOptionsValue allows the X-Frame-Options header value to be set with a custom value. This overrides the FrameDeny option.
		ContentTypeNosniff:      true,                 // If ContentTypeNosniff is true, adds the X-Content-Type-Options header with the value `nosniff`. Default is false.
		BrowserXssFilter:        true,                 // If BrowserXssFilter is true, adds the X-XSS-Protection header with the value `1; mode=block`. Default is false.
		ContentSecurityPolicy:   "default-src 'self'", // ContentSecurityPolicy allows the Content-Security-Policy header value to be set with a custom value. Default is "".
	})
	goji.Use(secureMiddleware.Handler)

	goji.Use(application.ApplyTemplates)
	goji.Use(application.ApplySessions)
	goji.Use(application.ApplyDatabase)
	goji.Use(application.ApplyAuth)
	goji.Use(application.ApplyProtectionFromCSRF)

	controller := &web.Controller{}
	apiController := &api.Controller{}

	//
	// Setup routes
	//

	// Set some static file
	// TODO: Should use nginx to serve these
	goji.Get("/robots.txt", http.FileServer(http.Dir(application.Configuration.PublicPath)))
	goji.Get("/favicon.ico", http.FileServer(http.Dir(application.Configuration.PublicPath+"/images")))

	// Home page
	goji.Get("/", application.Route(controller, "Index", system.RouteProtected))

	// Sign In routes
	goji.Get("/signin", application.Route(controller, "SignIn", system.RoutePublic))
	goji.Post("/signin", application.Route(controller, "SignInPost", system.RoutePublic))
	goji.Get("/forgot_password", application.Route(controller, "ForgotPassword", system.RoutePublic))
	goji.Post("/forgot_password", application.Route(controller, "ForgotPasswordPost", system.RoutePublic))
	goji.Get("/password_reset/:data", application.Route(controller, "PasswordReset", system.RoutePublic))

	// Register routes
	goji.Get("/register", application.Route(controller, "Register", system.RoutePublic))
	goji.Post("/register", application.Route(controller, "RegisterPost", system.RoutePublic))

	// Static
	goji.Get("/terms_and_conditions", application.Route(controller, "TermsAndConditions", system.RoutePublic))
	goji.Get("/privacy_policy", application.Route(controller, "Privacy", system.RoutePublic))
	goji.Get("/help", application.Route(controller, "Help", system.RoutePublic))

	// Logout
	goji.Get("/logout", application.Route(controller, "Logout", system.RouteProtected))

	// Download
	goji.Get("/download/SREPP.exe", application.Route(controller, "DownloadInstaller", system.RouteProtected))

	//
	// API
	//
	goji.Get("/api/systems.json", application.Route(apiController, "SystemsJSON", system.RouteProtected))
	goji.Get("/api/systeminfo.json", application.Route(apiController, "SystemInfoJSON", system.RouteProtected))
	goji.Get("/api/processes.json", application.Route(apiController, "ProcessesJSON", system.RouteProtected))
	goji.Get("/api/files.json", application.Route(apiController, "FilesJSON", system.RouteProtected))
	goji.Get("/api/fileinfo.json", application.Route(apiController, "FileInfoJSON", system.RouteProtected))

	goji.Get("/api/privacy_policy", application.Route(apiController, "PrivacyAPI", system.RouteProtected))
	goji.Get("/api/terms_and_conditions", application.Route(apiController, "TermsAPI", system.RouteProtected))
	goji.Get("/api/help", application.Route(apiController, "HelpAPI", system.RouteProtected))

	goji.Get("/api/profile.json", application.Route(apiController, "ProfileJSON", system.RouteProtected))
	goji.Post("/api/profile.json", application.Route(apiController, "PostProfileJSON", system.RouteProtected))
	goji.Post("/api/change_password.json", application.Route(apiController, "PostChangePasswordJSON", system.RouteProtected))
	goji.Post("/api/reset_password.json", application.Route(apiController, "PostResetPasswordJSON", system.RouteProtected))
	// Reset password is the same as change password, except it doesn't require you to type in your old password

	// Don't show 404's
	goji.NotFound(application.Route(controller, "Index", system.RouteProtected))

	//
	// Graceful shutodown
	//
	graceful.PostHook(func() {
		application.Close()
	})

	flag.Set("bind", fmt.Sprintf(":%s", application.Configuration.ListeningPort))
	goji.Serve()
}
