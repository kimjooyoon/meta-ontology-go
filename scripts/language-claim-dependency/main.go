package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependencyjudge"
)

func main() {
	sourcePath := flag.String("source", "", "claim dependency Gooo source")
	evidencePath := flag.String("evidence", "", "CI CURRENT_EVIDENCE receipt")
	priorPath := flag.String("prior-receipt", "", "prior UNKNOWN receipt to extend")
	contractPath := flag.String("contract", "", "external validator contract for the independent judge")
	manifestPath := flag.String("structural-manifest", "", "external structural inventory oracle for the independent judge")
	outputPath := flag.String("output", "", "claim receipt output path")
	check := flag.Bool("check", false, "run the independent raw-input judge")
	flag.Parse()
	if *sourcePath == "" || *evidencePath == "" || *outputPath == "" {
		fail("-source, -evidence, and -output are required")
	}
	if *check && (*contractPath == "" || *manifestPath == "") {
		fail("-check requires -contract and -structural-manifest")
	}
	source := read(*sourcePath)
	evidenceBytes := read(*evidencePath)
	var evidence claimdependency.EvidenceReceipt
	if err := json.Unmarshal(evidenceBytes, &evidence); err != nil {
		fail(err.Error())
	}
	if *contractPath != "" || *manifestPath != "" {
		if *contractPath == "" || *manifestPath == "" {
			fail("external validator contract and structural manifest must be supplied together")
		}
		contractBytes := read(*contractPath)
		manifestBytes := read(*manifestPath)
		if evidence.ValidatorContractPath != *contractPath || !bytes.Equal(evidence.ValidatorContractRaw, contractBytes) || evidence.ValidatorContractDigest != claimdependency.DigestBytesForObservation(contractBytes) {
			fail("EXTERNAL_VALIDATOR_CONTRACT_MISMATCH: producer evidence is not bound to external contract bytes")
		}
		if evidence.StructuralManifestPath != *manifestPath || !bytes.Equal(evidence.StructuralManifestRaw, manifestBytes) || evidence.StructuralManifestDigest != claimdependency.DigestBytesForObservation(manifestBytes) {
			fail("EXTERNAL_STRUCTURAL_MANIFEST_MISMATCH: producer evidence is not bound to external manifest bytes")
		}
	}
	var prior *claimdependency.Receipt
	var priorBytes []byte
	if *priorPath != "" {
		priorBytes = read(*priorPath)
		var value claimdependency.Receipt
		if err := json.Unmarshal(priorBytes, &value); err != nil {
			fail(err.Error())
		}
		prior = &value
	}
	receipt, err := claimdependency.Evaluate(source, *sourcePath, evidence, prior)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*outputPath, receipt)
	if *check {
		receiptBytes := read(*outputPath)
		if _, err := claimdependencyjudge.JudgeWithExternalMaterials(source, *sourcePath, priorBytes, evidenceBytes, receiptBytes, *contractPath, *manifestPath); err != nil {
			fail(err.Error())
		}
	}
	fmt.Printf("claim dependency operation=%s request_status=%s procedure=%s current=%d historical=%d unknown_evidence=%d distinct_propositions=%d/%d edges=%d/%d decision=%s direct_unknown=%d blocked=%d direct_refuted=%d dependency_refuted=%d discharged=%d refuted=%d causal_edges=%d/%d shortest_path_edge_union=%d edge_depth=%d authority=%s writes=%d\n", evidence.Operation, evidence.RequestStatus, evidence.Procedure, receipt.Metrics.CurrentEvidenceTotal, receipt.Metrics.HistoricalEvidenceTotal, receipt.Metrics.UnknownEvidenceTotal, receipt.Metrics.DistinctPropositionTotal, receipt.Metrics.FixedClaimTotal, receipt.Metrics.ObservedCausalEdgeTotal, receipt.Metrics.EligibleEdgeTotal, receipt.Decision.Value, receipt.Metrics.DirectUnknownClaimTotal, receipt.Metrics.DependencyBlockedClaimTotal, receipt.Metrics.DirectRefutedClaimTotal, receipt.Metrics.DependencyRefutedClaimTotal, receipt.Metrics.DischargedClaimTotal, receipt.Metrics.RefutedClaimTotal, receipt.Metrics.ObservedCausalEdgeTotal, receipt.Metrics.EligibleEdgeTotal, receipt.Metrics.ShortestPathEdgeUnionTotal, receipt.Metrics.MaximumCausePathDepth, receipt.Subject.AuthorityResolution, receipt.Subject.RepositoryWrites)
}

func read(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err.Error())
	}
	return data
}
func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err.Error())
	}
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(2) }
