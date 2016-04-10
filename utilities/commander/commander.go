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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"regexp"
	"strconv"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq" // Needed for gorp

	"qdserver/CallbackServer/command"
	"qdserver/lib/models"
	"qdserver/lib/utils"
)

var (
	// IDRegex define
	IDRegex = "^[0-9]+$"
	// UUIDRegex define
	UUIDRegex = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	// DB session
	DB *gorp.DbMap
)

// PrintUsage shows the command-line parameters for this tool
func PrintUsage() {
	fmt.Printf("Usage: commander\n")
	fmt.Printf("Usage: list task (<system_ID>|system_UUID>)\n")
	fmt.Printf("Usage: remove task <task_ID>\n")
}

// GetArg returns the argument at the given position, or exits the process
func GetArg(n int) string {
	if len(os.Args) <= n {
		PrintUsage()
		os.Exit(-1)
	}
	return os.Args[n]
}

// GetSystem receives either a system ID or UUID, figures out which it recieved,
// searches the DB and returns the other half
func GetSystem(identifier string) (ID int64, UUID string) {
	var err error
	if match, _ := regexp.MatchString(IDRegex, identifier); match {
		ID, err = strconv.ParseInt(identifier, 10, 64)
		if err != nil {
			panic(err)
		}
		var system *models.System
		err = DB.SelectOne(&system, "select * from systems where ID=:ID",
			map[string]interface{}{
				"ID": ID,
			})
		if err != nil {
			fmt.Printf("System not found\n")
			os.Exit(-1)
		}

		UUID, err = utils.ByteArrayToUUIDString(system.SystemUUID)
		if err != nil {
			panic(err)
		}

	} else if match, _ := regexp.MatchString(UUIDRegex, identifier); match {
		UUID = identifier
		SystemUUIDBytes, err := utils.UUIDStringToBytes(UUID)
		if err != nil {
			panic(err)
		}
		var system *models.System
		err = DB.SelectOne(&system, "select * from systems where SystemUUID=:UUID",
			map[string]interface{}{
				"UUID": SystemUUIDBytes,
			})
		if err != nil {
			panic(err)
		}

		ID = system.ID

	} else {
		panic("Malformed system identifier")
	}
	return ID, UUID
}

// MatchString given a string (command-line argument) looks to see if it matches one of the given values
func MatchString(needle string, haystacks ...string) bool {
	for _, haystack := range haystacks {
		if needle == haystack {
			return true
		}
	}

	return false
}

// Configuration is the main structure of our config.json file
type Configuration struct {
	DBConnectionString string `json:"db_connection_string"`
}

// Load parses our configuration file
func (configuration *Configuration) Load(filename string) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = configuration.Parse(data)
	return
}

// Parse parses the configuration file into a structure
func (configuration *Configuration) Parse(data []byte) (err error) {
	err = json.Unmarshal(data, &configuration)

	return
}

func main() {
	configfile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	config := &Configuration{}

	// Load our configuration file
	err := config.Load(*configfile)
	if err != nil {
		panic(err)
	}

	DB, err = models.InitDB(config.DBConnectionString)
	if err != nil {
		log.Fatalf("Unable to initialize the database: %v", err)
		panic(err)
	}
	defer DB.Db.Close()

	if MatchString(GetArg(1), "list", "ls") {
		if MatchString(GetArg(2), "task", "tasks") {
			systemID, systemUUID := GetSystem(GetArg(3))
			fmt.Printf("Listing tasks for system %d (%s)\n", systemID, systemUUID)

			var tasks []*models.Task
			_, err = DB.Select(&tasks, "select * from tasks where systemid=:systemid and deployedtoagentdate=0",
				map[string]interface{}{
					"systemid": systemID,
				})
			if err != nil {
				panic(err)
			}

			for _, task := range tasks {
				fmt.Printf("Task %d (%s): %s\n",
					task.ID,
					utils.Int64ToUnixTimeString(task.CreationDate, true),
					task.Command)
			}

			return
		}
	} else if MatchString(GetArg(1), "add") {
		if MatchString(GetArg(2), "task", "tasks") {
			systemID, systemUUID := GetSystem(GetArg(3))
			fmt.Printf("Adding task for system %d (%s)\n", systemID, systemUUID)

			if MatchString(GetArg(4), "update") {
				newVersion := GetArg(5)
				fmt.Printf("Creating update task for new version: %s\n", newVersion)

				agentTask := command.Update(newVersion)
				err = command.AddTask(DB, systemID, agentTask)
				if err != nil {
					panic(err)
				}
				return

			}
			panic("Unknown command")
		}
	} else if MatchString(GetArg(1), "remove", "rm") {
		if MatchString(GetArg(2), "task", "tasks") {
			taskID := GetArg(3)
			taskID64, err := strconv.ParseInt(taskID, 10, 64)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Removing task %d\n", taskID64)

			var task models.Task
			err = DB.SelectOne(&task, "select * from tasks where ID=:ID",
				map[string]interface{}{
					"ID": taskID64,
				})
			if err != nil {
				fmt.Printf("Error search db\n")
				panic(err)
			}

			task.DeployedToAgentDate = utils.DBTimeNow()

			_, err = DB.Update(&task)
			if err != nil {
				fmt.Printf("Update failure\n")
				panic(err)
			}

			return
		}
	}

	// No commands matched so print Usage
	fmt.Errorf("Arguments did not match available commands")
	PrintUsage()
	os.Exit(-1)
}
