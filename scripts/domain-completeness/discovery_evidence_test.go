package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"
)

func TestCapabilityDiscoveryEvidenceBindsExactReplayAndSource(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(capabilityDiscoverySourcePath)))
	if err != nil {
		t.Fatal(err)
	}
	contract, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(capabilityDiscoveryContractPath)))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "run", "./cmd/gooo", "discover", "--json", "--query", capabilityDiscoveryQuery,
		"--domain-contract", capabilityDiscoveryContractPath, capabilityDiscoverySourcePath)
	command.Dir = root
	reportRaw, err := command.Output()
	if err != nil {
		t.Fatalf("generate capability discovery receipt: %v", err)
	}
	replayRaw := bytes.Clone(reportRaw)
	validated := validateCapabilityDiscoveryEvidence(reportRaw, replayRaw, source, contract)
	if validated.State != "PASS" || validated.Reason != "CAPABILITY_DISCOVERY_EXACT_AND_REPLAYED" || len(validated.Refs) != 4 {
		t.Fatalf("valid capability discovery evidence = %#v", validated)
	}
	observation := capabilityDiscoveryObservation(reportRaw, replayRaw)
	if state, reason := validateCapabilityDiscoveryCells(observation, reportRaw, replayRaw); state != "PASS" {
		t.Fatalf("valid capability discovery cells = %q / %q", state, reason)
	}
	t.Run("unobserved capability stages do not gain credit", func(t *testing.T) {
		utilityContract := languageutility.Contract{
			UseCases: []languageutility.UseCaseSpec{{ID: "capability-discovery"}},
			Stages:   languageutility.CanonicalStages,
		}
		utilityReport := languageutility.Report{
			Cells: []languageutility.CellResult{{UseCaseID: "capability-discovery", StageID: "USER_ARTIFACT_VERIFIED",
				State: languageutility.StateClosed, EvidencePath: capabilityDiscoveryReportPath, EvidenceDigest: digestBytes(reportRaw)}},
			UseCases: []languageutility.UseCaseSummary{{ID: "capability-discovery", TotalCells: 7, ClosedCells: 6, RemainingCells: 1}},
			Summary:  languageutility.Summary{CellsTotal: 7, UseCasesTotal: 1},
		}
		inputs := loadedInputs{discovery: validated, reportRaw: []byte(`{"schema":"test"}`),
			programRaw: []byte("program"), inventoryRaw: []byte("{}"), observationRaw: []byte("{}")}
		generation := measureGenerationCoverage(dimensions[1], utilityContract, utilityReport, inputs)
		if generation.Status != "UNKNOWN" || generation.Numerator != 0 || generation.Denominator != 1 || generation.UnknownUnits != 1 {
			t.Fatalf("discovery report was counted as generated code: %#v", generation)
		}
		useCases := measureUseCaseCoverage(dimensions[3], utilityContract, utilityReport, inputs)
		if useCases.Status != "UNKNOWN" || useCases.Numerator != 0 || useCases.Denominator != 1 || useCases.UnknownUnits != 1 {
			t.Fatalf("discovery report was counted as independent behavior: %#v", useCases)
		}
		graph := graphSnapshot{GraphHash: "sha256:equal"}
		graphRaw, err := json.Marshal(graph)
		if err != nil {
			t.Fatal(err)
		}
		inputs.graphRaw, inputs.finalGraphRaw = graphRaw, bytes.Clone(graphRaw)
		inputs.graph, inputs.finalGraph = graph, graph
		reverse := measureReverseObservationCoverage(dimensions[2], inputs)
		if reverse.Status != "UNKNOWN" || reverse.Numerator != 1 || reverse.Denominator != 2 || reverse.UnknownUnits != 1 {
			t.Fatalf("discovery reverse observation was overstated: %#v", reverse)
		}
		boundary := measureBoundaryCoverage(dimensions[4], inputs)
		if boundary.Status != "PASS" || boundary.Numerator != 6 || boundary.Denominator != 6 {
			t.Fatalf("validated non-executing boundary was not counted: %#v", boundary)
		}
	})

	t.Run("changed replay bytes fail closed", func(t *testing.T) {
		changed := append(bytes.Clone(replayRaw), '\n')
		got := validateCapabilityDiscoveryEvidence(reportRaw, changed, source, contract)
		if got.State != "FAIL_CLOSED" || got.Reason != "CAPABILITY_DISCOVERY_REPLAY_MISMATCH" {
			t.Fatalf("changed replay = %#v", got)
		}
	})
	t.Run("changed source fails closed", func(t *testing.T) {
		changed := bytes.Replace(bytes.Clone(source), []byte("package "), []byte("package x"), 1)
		got := validateCapabilityDiscoveryEvidence(reportRaw, replayRaw, changed, contract)
		if got.State != "FAIL_CLOSED" || got.Reason != "CAPABILITY_DISCOVERY_SOURCE_DIGEST_MISMATCH" {
			t.Fatalf("changed source = %#v", got)
		}
	})
	t.Run("missing evidence remains unknown", func(t *testing.T) {
		got := validateCapabilityDiscoveryEvidence(nil, replayRaw, source, contract)
		if got.State != "UNKNOWN" || got.Reason != "CAPABILITY_DISCOVERY_EVIDENCE_MISSING" {
			t.Fatalf("missing report = %#v", got)
		}
	})
	t.Run("loader binds the evidence and counts each input once", func(t *testing.T) {
		evidenceDir := t.TempDir()
		files := map[string][]byte{
			capabilityDiscoveryReportPath:   reportRaw,
			capabilityDiscoveryReplayPath:   replayRaw,
			capabilityDiscoverySourceCopy:   source,
			capabilityDiscoveryContractCopy: contract,
			"observation.json":              mustJSON(t, capabilityDiscoveryObservation(reportRaw, replayRaw)),
		}
		for path, value := range files {
			fullPath := filepath.Join(evidenceDir, filepath.FromSlash(path))
			if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(fullPath, value, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		utilityReportRaw := mustJSON(t, languageutility.Report{Cells: []languageutility.CellResult{{
			UseCaseID: "debugging", StageID: "USER_ARTIFACT_VERIFIED", State: languageutility.StateClosed,
			EvidencePath: capabilityDiscoveryReportPath, EvidenceDigest: digestBytes(reportRaw),
		}}})
		if err := os.WriteFile(filepath.Join(evidenceDir, "report.json"), utilityReportRaw, 0o600); err != nil {
			t.Fatal(err)
		}
		wantBytes := int64(len(reportRaw) + len(replayRaw) + len(source) + len(contract) + len(files["observation.json"]) + len(utilityReportRaw))
		inputs := loadInputs(filepath.Join(t.TempDir(), "missing-contract.json"), evidenceDir)
		if inputs.discovery.State != "PASS" || inputs.discovery.Reason != "CAPABILITY_DISCOVERY_EXACT_AND_REPLAYED" {
			t.Fatalf("loaded discovery = %#v", inputs.discovery)
		}
		if inputs.inputFiles != 6 || inputs.inputBytes != wantBytes {
			t.Fatalf("unique input cost = %d files / %d bytes, want 6 / %d", inputs.inputFiles, inputs.inputBytes, wantBytes)
		}
	})
	t.Run("unsupported resource claim is rejected", func(t *testing.T) {
		observation := capabilityDiscoveryObservation(reportRaw, replayRaw)
		for index := range observation.Cells {
			if observation.Cells[index].UseCaseID == "capability-discovery" && observation.Cells[index].StageID == "RESOURCE_OBSERVED" {
				observation.Cells[index].State = languageutility.StateClosed
				observation.Cells[index].EvidenceKey = "capability-discovery.report"
				observation.Cells[index].EvidencePath = capabilityDiscoveryReportPath
				observation.Cells[index].EvidenceDigest = digestBytes(reportRaw)
			}
		}
		if state, reason := validateCapabilityDiscoveryCells(observation, reportRaw, replayRaw); state != "FAIL_CLOSED" ||
			reason != "CAPABILITY_DISCOVERY_RESOURCE_CLAIM_UNSUPPORTED" {
			t.Fatalf("unsupported resource claim = %q / %q", state, reason)
		}
	})
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func capabilityDiscoveryObservation(reportRaw, replayRaw []byte) languageutility.Observation {
	observation := languageutility.Observation{}
	for _, stage := range languageutility.CanonicalStages {
		cell := languageutility.CellObservation{UseCaseID: "capability-discovery", StageID: stage.ID}
		if stage.ID == "RESOURCE_OBSERVED" {
			cell.State = languageutility.StateOpen
			cell.Producer = "ci:capability-discovery"
			cell.Step = "COLLECT_RESOURCE_OBSERVED"
			cell.Reason = "DISCOVERY_RESOURCES_NOT_MEASURED"
		} else {
			cell.State = languageutility.StateClosed
			cell.Producer = "scripts/language-utility-evidence"
			cell.Step = "VERIFY_" + stage.ID
			cell.Reason = "EVIDENCE_ACCEPTED"
			cell.EvidenceKey = "capability-discovery.report"
			cell.EvidencePath = capabilityDiscoveryReportPath
			cell.EvidenceDigest = digestBytes(reportRaw)
			if stage.ID == "DETERMINISTIC_REPLAY" {
				cell.EvidencePath = capabilityDiscoveryReplayPath
				cell.EvidenceDigest = digestBytes(replayRaw)
			}
		}
		observation.Cells = append(observation.Cells, cell)
	}
	return observation
}
