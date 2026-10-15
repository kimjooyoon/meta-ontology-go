package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func writeRegistrationInventory(root string, stdout io.Writer) error {
	report, err := syntaxregistration.InspectInventory(os.DirFS(root))
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(report)
}
