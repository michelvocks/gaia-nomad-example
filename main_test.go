package main

import (
	"testing"

	sdk "github.com/gaia-pipeline/gosdk"
)

var args = sdk.Arguments{
	sdk.Argument{
		Key: "MYAPP_HOST",
		Value: "127.0.0.1:3306",
	},
	sdk.Argument{
		Key: "MYAPP_USER",
		Value: "root",
	},
	sdk.Argument{
		Key: "MYAPP_PASS",
		Value: "mysecretpw",
	},
}

func TestDeployApplication(t *testing.T) {
	if err := DeployApplication(args); err != nil {
		t.Fatal(err)
	}
}

func TestDBImportTestData(t *testing.T) {
	if err := DBImportTestData(args); err != nil {
		t.Fatal(err)
	}
}
