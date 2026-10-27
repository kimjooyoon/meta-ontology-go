package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceRegisteredSplitRestoresJointWrites(t *testing.T) {
	for _, fault := range []string{"rename-4", "rename-5", "directory-sync"} {
		t.Run(fault, func(t *testing.T) {
			options, proposal := saveRegisteredSplit(t, registeredSplitFixture(t))
			writes, err := sourceSplitAllWrites(proposal)
			if err != nil {
				t.Fatal(err)
			}
			ops, count := defaultAtomicFileOps(), 0
			ops.rename = func(from, to string) error {
				count++
				if fault == fmt.Sprintf("rename-%d", count) {
					return errors.New("injected joint publication failure")
				}
				return os.Rename(from, to)
			}
			if fault == "directory-sync" {
				ops.syncDir = func(string) error { return errors.New("injected directory sync failure") }
			}
			result, err := applyWorkspaceSourceSplit(options, ops)
			if err == nil || result.Application == nil || result.Application.State != "PRE_STATE_RETAINED" || result.Decision != "FAIL_CLOSED" {
				t.Fatal("joint failure did not retain pre-state", result.Application, err)
			}
			if err := checkAtomicPreconditions(writes); err != nil {
				t.Fatal("joint source/corpus/native publication was not restored", err)
			}
		})
	}
}

func TestWorkspaceRegisteredSplitRefusesChangedInputsAndSavedOutputs(t *testing.T) {
	for _, fault := range []string{"native", "corpus", "operation", "added-operation", "added-native", "ancestor", "output", "contract"} {
		t.Run(fault, func(t *testing.T) {
			request := registeredSplitFixture(t)
			options, proposal := saveRegisteredSplit(t, request)
			switch fault {
			case "native", "corpus", "operation":
				name := map[string]string{"native": splitRegistryDirectory + "/case_sources.go", "corpus": splitRegistryCorpus,
					"operation": "internal/meta/generation/operation-input-contract.gooo"}[fault]
				raw, err := os.ReadFile(filepath.Join(request.registryRoot, name))
				if err != nil {
					t.Fatal(err)
				}
				writeRegisteredSplitInput(t, request.registryRoot, name, append(raw, '\n'))
			case "added-native":
				writeRegisteredSplitInput(t, request.registryRoot, splitRegistryDirectory+"/later.go", []byte("package languagesyntax\n"))
			case "added-operation":
				writeRegisteredSplitInput(t, request.registryRoot, "internal/meta/generation/operation-input-contract-later.gooo", []byte("package operationinputcontract\nnamespace operationinputcontract\n"))
			case "ancestor":
				writeRegisteredSplitInput(t, request.registryRoot, "consumer.go", []byte("package fixture\n"))
			case "output", "contract":
				if fault == "output" {
					proposal.Registration.Outputs[0].Source += "\n# edited after planning\n"
					proposal.Registration.Outputs[0].SHA256 = workspaceDigest([]byte(proposal.Registration.Outputs[0].Source))
				} else {
					proposal.Registration.Operation.ActivityID = "urn:invented:operation"
				}
				raw, err := json.Marshal(sourceSplitPublicReceipt(proposal))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			renamed := 0
			ops := defaultAtomicFileOps()
			ops.rename = func(_, _ string) error { renamed++; return errors.New("unexpected rename") }
			result, err := applyWorkspaceSourceSplit(options, ops)
			if err == nil || renamed != 0 || result.Decision == "APPLIED" {
				t.Fatal("stale or altered registered proposal reached publication", err, renamed)
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(options.manifestPath), "inspect.gooo")); !os.IsNotExist(err) {
				t.Fatal("refused registration installed a fragment", err)
			}
		})
	}
}

func TestWorkspaceRegisteredSplitRechecksUnchangedNativeAfterStaging(t *testing.T) {
	request := registeredSplitFixture(t)
	options, _ := saveRegisteredSplit(t, request)
	name := filepath.Join(request.registryRoot, splitRegistryDirectory, "case_sources.go")
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	changed := append(bytes.Clone(raw), []byte("\n// independent edit\n")...)
	ops, written, renamed := defaultAtomicFileOps(), false, 0
	ops.syncFile = func(file *os.File) error {
		if !written {
			written = true
			if err := os.WriteFile(name, changed, 0o640); err != nil {
				return err
			}
		}
		return file.Sync()
	}
	ops.rename = func(_, _ string) error { renamed++; return errors.New("unexpected rename") }
	result, err := applyWorkspaceSourceSplit(options, ops)
	if err == nil || !written || renamed != 0 || result.Application.State != "PRE_STATE_RETAINED" {
		t.Fatal("unchanged registry input was not rechecked before publication", err)
	}
	current, err := os.ReadFile(name)
	if err != nil || !bytes.Equal(current, changed) {
		t.Fatal("refusal overwrote an independent edit", err)
	}
}

func TestWorkspaceRegisteredSplitAddsExternalManifestOnlyOnce(t *testing.T) {
	request := registeredSplitFixture(t)
	peer := filepath.Join(filepath.Dir(request.manifestPath), "peer.json")
	raw, err := os.ReadFile(request.manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(peer, raw, 0o640); err != nil {
		t.Fatal(err)
	}
	request.consumerWorkspaces = []string{peer}
	options, proposal := saveRegisteredSplit(t, request)
	writes, err := sourceSplitAllWrites(proposal)
	if err != nil || len(writes) != 6 {
		t.Fatal("external workspace did not join the shared write set", len(writes), err)
	}
	result, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
	if err != nil || result.Decision != "APPLIED" {
		t.Fatal("shared registered application failed", err)
	}
	current, err := os.ReadFile(peer)
	if err != nil || !strings.Contains(string(current), "inspect.gooo") {
		t.Fatal("external manifest did not receive new membership", err)
	}
}
