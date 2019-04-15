package main

import (
	"testing"

	sdk "github.com/gaia-pipeline/gosdk"
)

var args = sdk.Arguments{
	sdk.Argument{
		Key:   "MYAPP_HOST",
		Value: "host.docker.internal:3306",
	},
	sdk.Argument{
		Key:   "MYAPP_USER",
		Value: "root",
	},
	sdk.Argument{
		Key:   "MYAPP_PASS",
		Value: "mysecretpw",
	},
	sdk.Argument{
		Key:   "NOMAD_API",
		Value: "127.0.0.1",
	},
}

func TestDeployApplication(t *testing.T) {
	if err := DeployApplication(args); err != nil {
		t.Fatal(err)
	}
}

func TestDBImportTestData(t *testing.T) {
	// Deploy application and DB
	if err := DeployApplication(args); err != nil {
		t.Fatal(err)
	}

	// Wait for DB
	for id := range args {
		if args[id].Key == "MYAPP_HOST" {
			args[id].Value = "127.0.0.1:3306"
			break
		}
	}
	if err := WaitForDB(args); err != nil {
		t.Fatal(err)
	}

	// Insert test data
	if err := DBImportTestData(args); err != nil {
		t.Fatal(err)
	}
}
