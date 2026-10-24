package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func sourceSplitApplicationFixture(t *testing.T) (workspaceSourceSplitOptions, workspaceSourceSplitReceipt, map[string][]byte) {
	t.Helper()
	root := t.TempDir()
	inputs := map[string][]byte{}
	for _, name := range []string{"main.gooo", "workspace.json", "cases.json"} {
		raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "named-inputs", name))
		if err != nil {
			t.Fatal(err)
		}
		inputs[name] = raw
	}
	manifest, err := decodeWorkspaceManifest(inputs["workspace.json"])
	if err != nil {
		t.Fatal(err)
	}
	manifest.Packages[0].Sources = append(manifest.Packages[0].Sources, "peer.gooo")
	inputs["workspace.json"], err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inputs["peer.gooo"] = []byte("package namedinputs\nnamespace namedinputs\nactivity Dormant(Integer) -> Integer id \"urn:split:unused-peer\" computes \"return input\"\n")
	for name, raw := range inputs {
		if err := os.WriteFile(filepath.Join(root, name), raw, 0o640); err != nil {
			t.Fatal(err)
		}
	}
	options := workspaceSourceSplitOptions{packagePath: "named/inputs", source: "main.gooo",
		before: "urn:gooo:example:named-inputs:net", into: "net.gooo", manifestPath: filepath.Join(root, "workspace.json")}
	proposal, err := prepareWorkspaceSourceSplit(options, OSFileReader{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(sourceSplitPublicReceipt(proposal))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "proposal.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return workspaceSourceSplitOptions{applyReceipt: path, manifestPath: options.manifestPath}, proposal, inputs
}

func TestWorkspaceSourceSplitApplyCLIExecutesInstalledWorkspace(t *testing.T) {
	options, _, _ := sourceSplitApplicationFixture(t)
	var stdout, stderr bytes.Buffer
	args := []string{"package", "split-source", "--apply", options.applyReceipt, "--json", options.manifestPath}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var applied workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil {
		t.Fatal(err)
	}
	if applied.Decision != "APPLIED" || applied.Application == nil || applied.Application.State != "OUTPUT_VERIFIED" {
		t.Fatal("application not verified", applied)
	}
	raw, err := os.ReadFile(options.applyReceipt)
	if err != nil || applied.Application.ProposalSHA256 != workspaceDigest(raw) {
		t.Fatal("proposal origin lost", err)
	}
	root := filepath.Dir(options.manifestPath)
	if runtime.GOOS != "windows" {
		for _, name := range []string{"main.gooo", "net.gooo", "workspace.json"} {
			info, err := os.Stat(filepath.Join(root, name))
			if err != nil || info.Mode().Perm() != 0o640 {
				t.Fatal("source mode changed", name, err)
			}
		}
	}
	stdout.Reset()
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), options.manifestPath}, &stdout, &stderr); code != exitOK {
		t.Fatal("installed workspace did not execute", code, stdout.String(), stderr.String())
	}
	var execution packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &execution); err != nil {
		t.Fatal(err)
	}
	if execution.Result == nil || execution.Result.Runtime.FinitePassed != 4 || execution.Result.Runtime.FiniteTotal != 4 ||
		!execution.Result.Runtime.RuntimeReplayed || execution.Result.Runtime.ModelCalls != 0 {
		t.Fatal("independent installed-workspace oracle failed")
	}
	stdout.Reset()
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("old proposal applied twice", code)
	}
	if err := verifyAppliedSourceSplit(applied); err != nil {
		t.Fatal("refused repeat changed installed output", err)
	}
}

func TestWorkspaceSourceSplitApplyRejectsStaleOrAlteredProposal(t *testing.T) {
	for _, kind := range []string{"source", "dependency", "manifest", "forged-output", "duplicate-key", "occupied"} {
		t.Run(kind, func(t *testing.T) {
			options, proposal, inputs := sourceSplitApplicationFixture(t)
			root := filepath.Dir(options.manifestPath)
			switch kind {
			case "source":
				inputs["main.gooo"] = append(inputs["main.gooo"], []byte("\n# later edit\n")...)
			case "dependency":
				inputs["peer.gooo"] = append(inputs["peer.gooo"], []byte("\n# dependency edit\n")...)
			case "manifest":
				inputs["workspace.json"] = append(inputs["workspace.json"], '\n')
			case "occupied":
				inputs["net.gooo"] = []byte("unlisted destination must survive")
			case "forged-output":
				proposal.Result.Files[1].Source += "\n# altered proposal\n"
				raw, _ := json.Marshal(proposal)
				if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
					t.Fatal(err)
				}
			case "duplicate-key":
				raw, _ := json.Marshal(proposal)
				raw = append([]byte(`{"decision":"APPLIED",`), raw[1:]...)
				if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			for name, raw := range inputs {
				if err := os.WriteFile(filepath.Join(root, name), raw, 0o640); err != nil {
					t.Fatal(err)
				}
			}
			receipt, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
			if err == nil || receipt.Decision == "APPLIED" {
				t.Fatal("invalid proposal applied", kind)
			}
			checkSourceSplitApplicationInputs(t, options, inputs)
		})
	}
}

func TestWorkspaceSourceSplitApplyRestoresEveryOutputOnFailure(t *testing.T) {
	for _, fault := range []string{"rename-2", "rename-3", "directory-sync"} {
		t.Run(fault, func(t *testing.T) {
			options, _, inputs := sourceSplitApplicationFixture(t)
			ops := defaultAtomicFileOps()
			count := 0
			ops.rename = func(from, to string) error {
				count++
				if fault == "rename-2" && count == 2 || fault == "rename-3" && count == 3 {
					return errors.New("injected publication failure")
				}
				return os.Rename(from, to)
			}
			if fault == "directory-sync" {
				ops.syncDir = func(string) error { return errors.New("injected directory sync failure") }
			}
			receipt, err := applyWorkspaceSourceSplit(options, ops)
			if err == nil || receipt.Decision != "FAIL_CLOSED" || receipt.Application.State != "PRE_STATE_RETAINED" {
				t.Fatal("failed application lost pre-state observation", receipt, err)
			}
			checkSourceSplitApplicationInputs(t, options, inputs)
		})
	}
}

func TestWorkspaceSourceSplitApplyRechecksInputsAfterStaging(t *testing.T) {
	for _, name := range []string{"main.gooo", "peer.gooo"} {
		t.Run(name, func(t *testing.T) {
			options, _, inputs := sourceSplitApplicationFixture(t)
			inputs[name] = append(inputs[name], []byte("\n# changed during staging\n")...)
			ops := defaultAtomicFileOps()
			changed, renames := false, 0
			ops.syncFile = func(file *os.File) error {
				if !changed {
					changed = true
					if err := os.WriteFile(filepath.Join(filepath.Dir(options.manifestPath), name), inputs[name], 0o640); err != nil {
						return err
					}
				}
				return file.Sync()
			}
			ops.rename = func(from, to string) error { renames++; return os.Rename(from, to) }
			receipt, err := applyWorkspaceSourceSplit(options, ops)
			if err == nil || renames != 0 || receipt.Decision != "FAIL_CLOSED" {
				t.Fatal("stale staged outputs published", receipt, renames, err)
			}
			checkSourceSplitApplicationInputs(t, options, inputs)
		})
	}
}

func TestWorkspaceSourceSplitApplyReadsLegacyAndCompressedProposals(t *testing.T) {
	for _, kind := range []string{"legacy", "gzip"} {
		t.Run(kind, func(t *testing.T) {
			options, proposal, _ := sourceSplitApplicationFixture(t)
			value := proposal
			if kind == "gzip" {
				value = sourceSplitPublicReceipt(proposal)
			}
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			wire := raw
			if kind == "gzip" {
				var compressed bytes.Buffer
				writer := gzip.NewWriter(&compressed)
				if _, err := writer.Write(raw); err != nil {
					t.Fatal(err)
				}
				if err := writer.Close(); err != nil {
					t.Fatal(err)
				}
				wire = compressed.Bytes()
			}
			if err := os.WriteFile(options.applyReceipt, wire, 0o600); err != nil {
				t.Fatal(err)
			}
			applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
			if err != nil || applied.Decision != "APPLIED" || applied.Application.ProposalSHA256 != workspaceDigest(raw) {
				t.Fatal("receipt transport changed application/original JSON identity", applied, err)
			}
		})
	}
}

func TestWorkspaceSourceSplitWireOmitsUnchangedBodyButBindsItsIdentity(t *testing.T) {
	options, _, _ := sourceSplitApplicationFixture(t)
	peer := filepath.Join(filepath.Dir(options.manifestPath), "peer.gooo")
	raw, err := os.ReadFile(peer)
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte("#"+strings.Repeat("unchanged", 64<<10)+"\n")...)
	if err := os.WriteFile(peer, raw, 0o640); err != nil {
		t.Fatal(err)
	}
	proposal, err := prepareWorkspaceSourceSplit(workspaceSourceSplitOptions{
		packagePath: "named/inputs", source: "main.gooo", before: "urn:gooo:example:named-inputs:net", into: "net.gooo", manifestPath: options.manifestPath}, OSFileReader{})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := json.Marshal(sourceSplitPublicReceipt(proposal))
	if err != nil {
		t.Fatal(err)
	}
	if len(wire) >= 16<<10 || bytes.Contains(wire, []byte(`"manifest":`)) || !bytes.Contains(wire, []byte("input_image_digest")) {
		t.Fatal("unchanged source leaked into wire or input identity missing", len(wire))
	}
	if err := os.WriteFile(options.applyReceipt, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
	if err != nil || applied.Decision != "APPLIED" {
		t.Fatal("compact proposal failed to bind full current package", err)
	}
}

func checkSourceSplitApplicationInputs(t *testing.T, options workspaceSourceSplitOptions, inputs map[string][]byte) {
	t.Helper()
	root := filepath.Dir(options.manifestPath)
	for name, before := range inputs {
		after, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal("input changed on refusal/recovery", name, err)
		}
	}
	if _, occupied := inputs["net.gooo"]; !occupied {
		if _, err := os.Stat(filepath.Join(root, "net.gooo")); !os.IsNotExist(err) {
			t.Fatal("unexpected split output", err)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			t.Fatal("temporary output remained", entry.Name())
		}
	}
}
