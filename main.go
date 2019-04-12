package main

import (
	"database/sql"
	"fmt"
	"log"

	sdk "github.com/gaia-pipeline/gosdk"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/nomad/api"
)

var (
	taskGroup = "myAppTaskGroup"
	count     = 1
	memoryDB  = 800

	testData = []string{
		"Friedrich",
		"Hans",
		"Anna",
		"Bertha",
		"Heinrich",
		"Hermann",
		"Maria",
		"Martha",
		"Otto",
		"Walter",
		"Sieglinde",
		"Emma",
	}
)

func DBImportTestData(args sdk.Arguments) error {
	// Convert args
	argsMap := convArgsToMap(args)

	// Open DB connection
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/myappdb", argsMap["MYAPP_USER"], argsMap["MYAPP_PASS"], argsMap["MYAPP_HOST"]))
	if err != nil {
		log.Printf("failed to connect to database. Error: %s", err.Error())
		return err
	}
	defer db.Close()

	// Drop table if exists
	dropTable, err := db.Query("DROP TABLE IF EXISTS names;")
	if err != nil {
		log.Printf("failed to drop table. Error: %s", err.Error())
		return err
	}
	defer dropTable.Close()

	// Create table
	insertTable, err := db.Query("CREATE TABLE names (name VARCHAR(20));")
	if err != nil {
		log.Printf("failed to create table. Error: %s", err.Error())
		return err
	}
	defer insertTable.Close()

	// Insert test data
	insertData, err := db.Prepare("INSERT INTO names VALUES( ? )")
	if err != nil {
		log.Printf("failed to prepare statement. Error: %s", err.Error())
		return err
	}
	defer insertData.Close()

	for _, name := range testData {
		_, err = insertData.Exec(name)
		if err != nil {
			log.Printf("failed to insert test data. Error: %s", err.Error())
			return err
		}
	}

	return nil
}

func DeployApplication(args sdk.Arguments) error {
	// Setup config
	conf := &api.Config{Address: "http://127.0.0.1:4646"}

	// Convert args
	argsMap := convArgsToMap(args)

	// Create client instance
	client, err := api.NewClient(conf)
	if err != nil {
		log.Printf("failed to connect to Nomad API. Error: %s", err.Error())
		return err
	}

	// Create service job
	job := api.NewServiceJob("myapp", "myapp", "eu", 50)
	job.Datacenters = []string{"dc1"}
	job.TaskGroups = []*api.TaskGroup{
		&api.TaskGroup{
			Name:  &taskGroup,
			Count: &count,
			Tasks: []*api.Task{
				&api.Task{
					Name:   "myapp",
					Driver: "docker",
					Config: map[string]interface{}{
						"image": "michelvocks/myapp:latest",
					},
					Env: map[string]string{
						"MYAPP_DB_HOST":     argsMap["MYAPP_HOST"],
						"MYAPP_DB_USERNAME": argsMap["MYAPP_USER"],
						"MYAPP_DB_PASSWORD": argsMap["MYAPP_PASS"],
					},
					Services: []*api.Service{
						&api.Service{
							Name:      "myapp-frontend",
							PortLabel: "frontend",
						},
					},
					Resources: &api.Resources{
						Networks: []*api.NetworkResource{{
							ReservedPorts: []api.Port{{
								Label: "frontend",
								Value: 9090,
							}},
						}},
					},
				},
				&api.Task{
					Name:   "db",
					Driver: "docker",
					Config: map[string]interface{}{
						"image": "mysql:latest",
					},
					Env: map[string]string{
						"MYSQL_ROOT_PASSWORD": argsMap["MYAPP_PASS"],
						"MYSQL_DATABASE":      "myappdb",
					},
					Services: []*api.Service{
						&api.Service{
							Name:      "db-backend",
							PortLabel: "backend",
						},
					},
					Resources: &api.Resources{
						MemoryMB: &memoryDB,
						Networks: []*api.NetworkResource{{
							ReservedPorts: []api.Port{{
								Label: "backend",
								Value: 3306,
							}},
						}},
					},
				},
			},
		},
	}

	// Get jobs and register job
	jobs := client.Jobs()
	_, _, err = jobs.Register(job, &api.WriteOptions{})
	if err != nil {
		log.Printf("failed to register job. Error: %s", err.Error())
		return err
	}
	return nil
}

func convArgsToMap(args sdk.Arguments) map[string]string {
	// Extract arguments
	argsMap := map[string]string{}
	for _, arg := range args {
		argsMap[arg.Key] = arg.Value
	}
	return argsMap
}

func main() {
	jobs := sdk.Jobs{
		sdk.Job{
			Title: "Deploy Application",
			Handler: DeployApplication,
			Description: "deploy the application with database",
		},
		sdk.Job{
			Title: "Import test data",
			Handler: DBImportTestData,
			Description: "import test data into application database",
		},
	}

	if err := sdk.Serve(jobs); err != nil {
		panic(err)
	}
}
