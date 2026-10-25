package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func operationContractFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := operationInputContractFiles.ReadFile("operation-input-contract.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func contractTestSources(raw []byte) []sourcebundle.Source {
	return []sourcebundle.Source{{Filename: "operation-input-contract.gooo", Content: string(raw)}}
}

func parseOperationInputContract(raw []byte) (operationInputContract, error) {
	return parseOperationInputContractSources(contractTestSources(raw))
}
