package main

import (
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
	predicate := flag.String("predicate", "", "UNKNOWN, EVIDENCE_ACCEPTED, or EXPLICIT_CONTRADICTION")
	evidence := flag.String("evidence", "", "deterministic evidence text for accepted/refuted observations")
	priorPath := flag.String("prior-receipt", "", "prior receipt to extend with append-only recovery")
	observationPath := flag.String("observation", "", "observation JSON output path")
	outputPath := flag.String("output", "", "receipt output path")
	check := flag.Bool("check", false, "run the independent raw-input judge after producing the receipt")
	flag.Parse()
	if *sourcePath == "" || *predicate == "" || *outputPath == "" {
		fail("-source, -predicate, and -output are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	obs, err := claimdependency.ObservationForSource(source, *sourcePath, claimdependency.ObservationPredicate(*predicate), *evidence)
	if err != nil {
		fail(err.Error())
	}
	var prior *claimdependency.Receipt
	var priorBytes []byte
	if *priorPath != "" {
		priorBytes, err = os.ReadFile(*priorPath)
		if err != nil {
			fail(err.Error())
		}
		var value claimdependency.Receipt
		if err := json.Unmarshal(priorBytes, &value); err != nil {
			fail(err.Error())
		}
		prior = &value
	}
	receipt, err := claimdependency.Evaluate(source, *sourcePath, obs, prior)
	if err != nil {
		fail(err.Error())
	}
	if *observationPath == "" {
		*observationPath = *outputPath + ".observation.json"
	}
	writeJSON(*observationPath, obs)
	writeJSON(*outputPath, receipt)
	if *check {
		observationBytes, err := os.ReadFile(*observationPath)
		if err != nil {
			fail(err.Error())
		}
		receiptBytes, err := os.ReadFile(*outputPath)
		if err != nil {
			fail(err.Error())
		}
		if _, err := claimdependencyjudge.Judge(source, *sourcePath, priorBytes, observationBytes, receiptBytes); err != nil {
			fail(err.Error())
		}
	}
	fmt.Printf("claim dependency predicate=%s claims=%d/%d edges=%d open=%d discharged=%d refuted=%d direct_unknown=%d blocked=%d direct_refuted=%d dependency_refuted=%d recovery_edges=%d transition_total=%d read_only=%t repository_writes=%d\n", obs.Predicate, receipt.Metrics.ClassifiedClaimTotal, receipt.Metrics.FixedClaimTotal, receipt.Metrics.FixedEdgeTotal, receipt.Metrics.OpenClaimTotal, receipt.Metrics.DischargedClaimTotal, receipt.Metrics.RefutedClaimTotal, receipt.Metrics.DirectUnknownClaimTotal, receipt.Metrics.DependencyBlockedClaimTotal, receipt.Metrics.DirectRefutedClaimTotal, receipt.Metrics.DependencyRefutedClaimTotal, receipt.Metrics.ObservedRecoveryEdgeTotal, receipt.Metrics.TransitionTotal, receipt.Subject.ReadOnly, receipt.Subject.RepositoryWrites)
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

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
