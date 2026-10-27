package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSharedConsumerSplitRejectsStaleOrForgedConsumers(t *testing.T) {
	for _, mutation := range []string{"manifest", "dependency", "output", "missing-result", "duplicate"} {
		t.Run(mutation, func(t *testing.T) {
			options, proposal, inputs := sharedConsumerApplicationFixture(t)
			root := filepath.Dir(options.manifestPath)
			switch mutation {
			case "manifest", "dependency":
				name := "second.workspace.json"
				if mutation == "dependency" {
					name = "other.gooo"
				}
				inputs[name] = append(inputs[name], '\n')
				if err := os.WriteFile(filepath.Join(root, name), inputs[name], 0o640); err != nil {
					t.Fatal(err)
				}
			case "output":
				proposal.Consumers[0].Result.Files[1].Source += "# changed\n"
			case "missing-result":
				proposal.Consumers[0].Result = nil
			case "duplicate":
				proposal.Consumers = append(proposal.Consumers, proposal.Consumers[0])
			}
			raw, err := json.Marshal(proposal)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
			if err == nil || applied.Decision == "APPLIED" {
				t.Fatal("invalid consumer accepted", mutation, err)
			}
			checkSourceSplitApplicationInputs(t, options, inputs)
		})
	}
}

func TestSharedConsumerSplitRollsBackAllManifests(t *testing.T) {
	for _, fault := range []string{"last-manifest", "directory-sync", "changed-dependency"} {
		t.Run(fault, func(t *testing.T) {
			options, _, inputs := sharedConsumerApplicationFixture(t)
			ops := defaultAtomicFileOps()
			renames := 0
			ops.rename = func(from, to string) error {
				renames++
				if fault == "last-manifest" && renames == 4 {
					return errors.New("injected second manifest failure")
				}
				return os.Rename(from, to)
			}
			switch fault {
			case "directory-sync":
				ops.syncDir = func(string) error { return errors.New("injected directory sync failure") }
			case "changed-dependency":
				ops.beforePublish = func() error {
					inputs["other.gooo"] = append(inputs["other.gooo"], '\n')
					return os.WriteFile(filepath.Join(filepath.Dir(options.manifestPath), "other.gooo"), inputs["other.gooo"], 0o640)
				}
			}
			applied, err := applyWorkspaceSourceSplit(options, ops)
			if err == nil || applied.Decision != "FAIL_CLOSED" || applied.Application.State != "PRE_STATE_RETAINED" ||
				fault == "changed-dependency" && renames != 0 {
				t.Fatal("incomplete group publication was accepted", fault, err, applied)
			}
			checkSourceSplitApplicationInputs(t, options, inputs)
		})
	}
}
