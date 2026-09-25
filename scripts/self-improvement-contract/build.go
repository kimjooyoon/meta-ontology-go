package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	reportSchema = "gooo/self-improvement-contract/v1"
	metaprogram  = "scripts/self-improvement-contract"
)

func project(path string, source []byte, commit string) analysis {
	registry := registrySnapshot()
	result := analysis{Report: Report{
		Schema: reportSchema, Metaprogram: metaprogram, CommitSHA: commit,
		ContractPath: path, SourceSHA256: sourceDigest(source),
		RegistryDigest: registryDigest(registry), Registry: registry,
		Errors: []string{}, Indicators: []Indicator{},
		PromotionAuthorized: false,
	}}
	model, err := compileContract(path, source)
	if err != nil {
		result.Report.Errors = append(result.Report.Errors, err.Error())
		result.Report.ExecutorCoverage = coverExecutors(model, registry)
		return result
	}
	result.Report.SemanticHash = model.SemanticHash
	result.Report.EntityCount = len(model.Entities)
	result.Report.ActivityCount = len(model.Activities)
	result.Report.ExecutorCoverage = coverExecutors(model, registry)
	result.SemanticOK = model.SemanticHash != ""
	result.LoopOK = closedLoop(model)
	result.ExecutorOK = completeCoverage(result.Report.ExecutorCoverage)
	result.TrilemmaOK = trilemmaChoice(model, registry)
	result.ObservationOK = readOnlyObservation(model)
	return result
}

func buildReport(path string, source []byte, commit string) Report {
	left, right := project(path, source, commit), project(path, source, commit)
	leftBytes, _ := json.Marshal(left.Report)
	rightBytes, _ := json.Marshal(right.Report)
	left.Report.Indicators = contractIndicators(
		left, bytes.Equal(leftBytes, rightBytes),
	)
	finishReport(&left.Report)
	return left.Report
}

func sourceDigest(source []byte) string {
	sum := sha256.Sum256(source)
	return hex.EncodeToString(sum[:])
}

func verdict(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}
