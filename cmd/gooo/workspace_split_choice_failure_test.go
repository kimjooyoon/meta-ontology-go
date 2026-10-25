package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceSplitChoiceRecomputedBeforePublication(t *testing.T) {
	for _, mutation := range []string{"count", "boundary", "limit-result", "no-boundary", "no-limit"} {
		t.Run(mutation, func(t *testing.T) {
			options, receipt, _ := chosenRegisteredSplit(t, 75)
			switch mutation {
			case "count":
				receipt.Selection.Candidates++
			case "boundary":
				receipt.Selection.Boundary.BeforeID = "urn:unselected"
			case "limit-result":
				receipt.Selection.LimitSatisfied = false
			case "no-boundary":
				receipt.Selection.Boundary = nil
			case "no-limit":
				receipt.Selection.LineLimit = 0
			}
			raw, err := json.Marshal(receipt)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(options.registryRoot, "proposal.json")
			if err := os.WriteFile(name, raw, 0600); err != nil {
				t.Fatal(err)
			}
			before, err := sourceSplitAllWrites(receipt)
			if err != nil {
				t.Fatal(err)
			}
			_, err = applyWorkspaceSourceSplit(workspaceSourceSplitOptions{
				manifestPath: options.manifestPath, applyReceipt: name}, defaultAtomicFileOps())
			if err == nil {
				t.Fatal("altered selection metadata reached publication")
			}
			if err := checkAtomicPreconditions(before); err != nil {
				t.Fatal("refused selection changed a publication target", err)
			}
		})
	}
}

func TestWorkspaceSplitChoiceRetainsSatisfiedSource(t *testing.T) {
	options := registeredSplitFixture(t)
	options.before, options.lineLimit = "", 82
	before, err := os.ReadFile(filepath.Join(filepath.Dir(options.manifestPath), options.source))
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := prepareWorkspaceSourceSplit(options, OSFileReader{})
	if err == nil || receipt.Result != nil || receipt.Selection == nil ||
		receipt.Selection.State != "ALREADY_WITHIN_LIMIT" || !receipt.Selection.LimitSatisfied {
		t.Fatal("satisfied file received a writable proposal", receipt, err)
	}
	after, err := os.ReadFile(filepath.Join(filepath.Dir(options.manifestPath), options.source))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("satisfied source changed", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(options.manifestPath), options.into)); !os.IsNotExist(err) {
		t.Fatal("unnecessary output created", err)
	}
}

type changingSplitChoiceReader struct {
	source string
	reads  int
}

func (reader *changingSplitChoiceReader) ReadFile(name string) ([]byte, error) {
	raw, err := os.ReadFile(name)
	if name == reader.source {
		reader.reads++
		if reader.reads > 1 {
			raw = append(raw, '\n')
		}
	}
	return raw, err
}

func TestWorkspaceSplitChoiceRejectsChangingSelectionInputs(t *testing.T) {
	options := registeredSplitFixture(t)
	options.before, options.registryRoot, options.lineLimit = "", "", 75
	reader := &changingSplitChoiceReader{source: filepath.Join(filepath.Dir(options.manifestPath), options.source)}
	receipt, err := prepareWorkspaceSourceSplit(options, reader)
	if err == nil || reader.reads < 2 || receipt.Result != nil || receipt.UpdatedWorkspace != nil ||
		receipt.Decision != "FAIL_CLOSED" || receipt.Selection == nil {
		t.Fatal("choice from earlier source bytes reached a writable proposal", receipt, err)
	}
}
