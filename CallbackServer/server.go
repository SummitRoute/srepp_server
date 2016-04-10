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

	log "github.com/Sirupsen/logrus"
	"github.com/goji/glogrus"
	_ "github.com/lib/pq" // Needed for gorp
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web/middleware"

	"qdserver/CallbackServer/controllers/api"
	"qdserver/CallbackServer/system"
)

func main() {
	configfile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	var application = &system.Application{}
	application.Init(configfile)
	application.ConnectToDatabase()
	application.ConnectToQueues()

	//
	// Apply middleware
	//

	goji.Use(application.ApplyConfig)

	// User defined logs
	log.SetFormatter(new(log.JSONFormatter))

	// Goji logs
	goji.Abandon(middleware.Logger)
	logr := log.New()
	logr.Formatter = new(log.JSONFormatter)
	goji.Use(glogrus.NewGlogrus(logr, "callbackserver"))

	controller := &api.Controller{}

	//
	// Setup routes
	//

	goji.Get("/", application.Route(controller, "Status"))

	// Clients call getRegisterSystem to inform the server they've been installed and to
	// get a system UUID
	goji.Post("/api/v1/Register", application.Route(controller, "RegisterAgent"))
	goji.Post("/api/v1/ProcessEvent", application.Route(controller, "ProcessEvent"))
	goji.Post("/api/v1/CatalogFileEvent", application.Route(controller, "CatalogFileEvent"))
	goji.Post("/api/v1/UploadFile", application.Route(controller, "UploadFile"))
	goji.Post("/api/v1/Heartbeat", application.Route(controller, "Heartbeat"))
	goji.Post("/api/v1/GetUpdate", application.Route(controller, "GetUpdate"))

	graceful.PostHook(func() {
		application.Close()
	})

	flag.Set("bind", fmt.Sprintf(":%s", application.Configuration.ListeningPort))
	goji.Serve()
}
