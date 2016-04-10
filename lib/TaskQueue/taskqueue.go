////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////


package TaskQueue

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
)

// analyzeFile used for creating a task to analyze a file
type analyzeFile struct {
	FileID int64
}

// createAnalyzeFileMessage returns message to be sent to the queue for analyzing file messages
func createAnalyzeFileMessage(fileID int64) ([]byte, error) {
	var task analyzeFile
	task.FileID = fileID
	return json.Marshal(task)
}

// CreateTaskToAnalyzeFile creates a rabbitMQ task for a worker to pick up
func CreateTaskToAnalyzeFile(ch *amqp.Channel, fileID int64) (err error) {
	taskMsg, err := createAnalyzeFileMessage(fileID)
	if err != nil {
		log.Errorf("Failed to create task")
		panic(err)
	}

	err = QueueTask(ch, "analyzefile", taskMsg)
	if err != nil {
		log.Errorf("Failed to queue task")
		panic(err)
	}
	return nil
}

// analyzeCatalog used for creating a task to analyze a catalog
type analyzeCatalog struct {
	CatalogID int64
}

// createAnalyzeCatalogMessage returns message to be sent to the queue for analyzing catalog messages
func createAnalyzeCatalogMessage(catalogID int64) ([]byte, error) {
	var task analyzeCatalog
	task.CatalogID = catalogID
	return json.Marshal(task)
}

// CreateTaskToAnalyzeCatalog creates a rabbitMQ task for a worker to pick up
func CreateTaskToAnalyzeCatalog(ch *amqp.Channel, catalogID int64) (err error) {
	taskMsg, err := createAnalyzeCatalogMessage(catalogID)
	if err != nil {
		log.Errorf("Failed to create task")
		panic(err)
	}

	err = QueueTask(ch, "analyzecatalog", taskMsg)
	if err != nil {
		log.Errorf("Failed to queue task")
		panic(err)
	}
	return nil
}

// QueueTask sends the message to the queueing server
func QueueTask(ch *amqp.Channel, channelName string, msg []byte) error {
	err := ch.Publish(
		"",          // exchange
		channelName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	return err
}

// CreateQueue declares durable queues
func CreateQueue(ch *amqp.Channel, queueName string) (err error) {
	// TODO Need to deal with re-connects.  Should probably look at: https://github.com/richardiux/rabbitmq-to-resque

	queueArguments := amqp.Table{"x-ha-policy": "all"}
	_, err = ch.QueueDeclare(
		queueName,      // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		queueArguments, // arguments
	)

	return err
}
