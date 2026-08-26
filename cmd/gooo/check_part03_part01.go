package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

func runSemanticCheck(options checkOptions, jsonMode bool, source []byte, file *syntax.File, diagnostics syntax.Diagnostics, deadline time.Time, stdout, stderr io.Writer) (string, *provenancePublishResponse, int) {
	if !options.semantic {
		return "", nil, exitOK
	}
	ir, err := semanticCheckIR(file, remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return "", nil, reportFailure(true, stdout, stderr, "check", options.filename, "semantic.lowering", err.Error(), syntaxFileSpan(file))
		}
		if !reportSemanticDiagnostic(options.filename, file, err, stderr) {
			return "", nil, exitFailure
		}
		return "", nil, exitFailure
	}
	semanticHash := ir.StableHash()
	if options.provenanceStore == "" {
		if !jsonMode {
			if _, err := fmt.Fprintln(stderr, deferredCheckProvenance); err != nil {
				return semanticHash, nil, exitFailure
			}
		}
		return semanticHash, nil, exitOK
	}
	response, err := publishSemanticCheckProvenance(options.filename, source, file, ir, options.provenanceStore)
	if err == nil {
		return semanticHash, &response, exitOK
	}
	response, sealErr := rejectSemanticCheckProvenance(response, err)
	if sealErr != nil {
		return semanticHash, nil, exitFailure
	}
	if jsonMode {
		report := newJSONReport("check", "error", options.filename, syntaxCLIDiagnostics(diagnostics))
		report.SemanticHash = semanticHash
		report.Provenance = &response
		if writeErr := writeJSONReport(stdout, report); writeErr != nil {
			return semanticHash, nil, exitFailure
		}
		return semanticHash, nil, exitFailure
	}
	fmt.Fprintf(stdout, "ok: %s\nprovenance: rejected\n", options.filename)
	fmt.Fprintf(stderr, "gooo: %s: %s: %v\n", options.filename, provenanceErrorCode(err), err)
	return semanticHash, nil, exitFailure
}
func writeCheckResult(jsonMode bool, filename, semanticHash string, provenanceResponse *provenancePublishResponse, diagnostics syntax.Diagnostics, stdout io.Writer) int {
	if jsonMode {
		report := newJSONReport("check", "ok", filename, syntaxCLIDiagnostics(diagnostics))
		report.SemanticHash = semanticHash
		report.Provenance = provenanceResponse
		if err := writeJSONReport(stdout, report); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "ok: %s\n", filename)
	if provenanceResponse != nil {
		fmt.Fprintf(stdout, "provenance: %s records=%d store_digest=%s\n", provenanceResponse.Status, len(provenanceResponse.Records), provenanceResponse.StoreDigest)
	}
	return exitOK
}

const checkUsage = "usage: gooo check [--semantic] [--provenance-store <ledger.jsonl>] [--json] <file.gooo>"
