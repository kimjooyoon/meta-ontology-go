package bodyexecution

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReturnProfileLegacy21737(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	file, err := os.Open(filepath.Join(root, "internal", "bodyexecution", "testdata", "legacy-return21511.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var prior Composition
	if err := json.NewDecoder(reader).Decode(&prior); err != nil {
		t.Fatal(err)
	}
	for _, step := range prior.Steps {
		if step.Generation.Report.BodyReturns != nil {
			t.Fatal("legacy fixture acquired return annotations")
		}
	}
	inventory, err := compositionFieldBindings(prior)
	if err != nil || len(inventory.bindings) != 0 || len(inventory.returns) != 0 {
		t.Fatal("legacy absence acquired new return journal subjects", inventory, err)
	}
	t.Log("saved legacy record only;missing return annotations retain old field inventory;no generation,case evaluation,producer execution or model call")
}
