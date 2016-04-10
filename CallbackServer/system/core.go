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
	"io"
	"net/http"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq" // Needed for gorp
	"github.com/streadway/amqp"
	"github.com/zenazn/goji/web"

	"qdserver/lib/models"
	"qdserver/lib/taskqueue"
)

// Application holds our app's global variables
type Application struct {
	Configuration   *Configuration
	DBSession       *gorp.DbMap
	QueueConnection *amqp.Connection
	QueueChannel    *amqp.Channel
}

// Init initializes our globals
func (application *Application) Init(filename *string) {
	application.Configuration = &Configuration{}

	// Load our configuration file
	err := application.Configuration.Load(*filename)
	if err != nil {
		log.Fatalf("Can't read configuration file: %s", err)
		panic(err)
	}
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

// ConnectToQueues initializes our AMQP connection
func (application *Application) ConnectToQueues() {
	conn, err := amqp.Dial(application.Configuration.AMQPServer)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel: %v", err)
		panic(err)
	}

	// Create queues
	for _, queueName := range []string{"analyzefile", "analyzecatalog"} {
		err = TaskQueue.CreateQueue(ch, queueName)
		if err != nil {
			conn.Close()
			ch.Close()
			log.Errorf("Failed to create queue %s; %v", queueName, err)
		}
	}

	application.QueueChannel = ch
	application.QueueConnection = conn
}

// Close cleans up anything nicely before exiting the process
func (application *Application) Close() {
	log.Info("Bye!")
	application.DBSession.Db.Close()
	application.QueueConnection.Close()
	application.QueueChannel.Close()
}

// Route defines a web route
func (application *Application) Route(controller interface{}, route string) interface{} {
	fn := func(c web.C, w http.ResponseWriter, r *http.Request) {
		c.Env["Content-Type"] = "application/json"

		methodValue := reflect.ValueOf(controller).MethodByName(route)
		methodInterface := methodValue.Interface()
		method := methodInterface.(func(c web.C, r *http.Request) (string, int))

		body, code := method(c, r)

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
