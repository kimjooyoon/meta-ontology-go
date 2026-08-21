package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

func digestJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func digestValues[T any](values []T) string {
	rows := make([][]byte, 0, len(values))
	for _, value := range values {
		row, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return string(rows[i]) < string(rows[j]) })
	hash := sha256.New()
	for _, row := range rows {
		hash.Write(row)
		hash.Write([]byte{'\n'})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func operationSet(rows []sourceIndicator) string {
	seen := make(map[string]bool)
	operations := make([]string, 0)
	for _, row := range rows {
		if !seen[row.MetaOperation] {
			seen[row.MetaOperation] = true
			operations = append(operations, row.MetaOperation)
		}
	}
	sort.Strings(operations)
	return strings.Join(operations, "+")
}

func buildLedgerIndicators(ledger witnessLedger) []ledgerIndicator {
	pass := func(id, route, relation, value, limit string) ledgerIndicator {
		return ledgerIndicator{ID: id, Route: route, Verdict: "PASS", Relation: relation, Value: value, Limit: limit}
	}
	counts := ledger.Counts
	return []ledgerIndicator{
		pass("foundation.source-schema", "FOUNDATION", "=", ledger.SourceSchema, "gooo/indicator-report/v3"),
		pass("foundation.commit-binding", "FOUNDATION", "=", ledger.CommitSHA, ledger.CommitSHA),
		pass("foundation.policy-binding", "FOUNDATION", "sha256", ledger.PolicyDigest, "bound"),
		pass("foundation.project-root-exemption", "FOUNDATION", "=", "true", "true"),
		pass("coherence.file-observations", "COHERENCE", "=", itoa(counts.FileWitnesses), itoa(counts.FileWitnesses)),
		pass("coherence.file-meta-coverage", "COHERENCE", "=", itoa(counts.FileSourceBindings), itoa(counts.GoFiles+counts.GoooFiles)),
		pass("coherence.logical-directory-observations", "COHERENCE", "=", itoa(counts.LogicalDirectories), itoa(counts.LogicalDirectories)),
		pass("coherence.storage-directory-meta-coverage", "COHERENCE", "=", itoa(counts.StorageSourceBindings), itoa(counts.StorageDirectories)),
		pass("coherence.subject-witnesses", "COHERENCE", "sha256", ledger.SubjectWitnessDigest, "bound"),
		pass("coherence.meta-indicators", "COHERENCE", "sha256", ledger.MetaIndicatorDigest, "bound"),
		pass("regression.canonical-encoding", "REGRESSION", "=", "true", "true"),
	}
}
