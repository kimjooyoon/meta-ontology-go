package audienceresolution

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func fixtureInput(t *testing.T) Input {
	t.Helper()
	root := filepath.Join("..", "..", "..", "examples", "audience-resolution")
	var contract Contract
	var ledger Ledger
	contractData, err := os.ReadFile(filepath.Join(root, "contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(contractData, &contract); err != nil {
		t.Fatal(err)
	}
	ledgerData, err := os.ReadFile(filepath.Join(root, "ledger.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(ledgerData, &ledger); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(root, "main.gooo")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	return Input{Contract: contract, Ledger: ledger,
		SourcePath: "examples/audience-resolution/main.gooo", Source: source}
}
