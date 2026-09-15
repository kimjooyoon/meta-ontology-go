package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

var (
	shaPattern    = regexp.MustCompile("^[0-9a-f]{40}$")
	digestPattern = regexp.MustCompile("^[0-9a-f]{64}$")
	ledgerPattern = regexp.MustCompile("^sha256:[0-9a-f]{64}$")
)

type contractIndicator struct {
	Route   string `json:"route"`
	Verdict string `json:"verdict"`
}
type contractCoverage struct {
	Covered bool `json:"covered"`
}
type contractDocument struct {
	Schema              string              `json:"schema"`
	CommitSHA           string              `json:"commit_sha"`
	SourceSHA256        string              `json:"source_sha256"`
	SemanticHash        string              `json:"semantic_hash"`
	RegistryDigest      string              `json:"registry_digest"`
	Status              string              `json:"status"`
	PromotionAuthorized bool                `json:"promotion_authorized"`
	Indicators          []contractIndicator `json:"indicators"`
	ExecutorCoverage    []contractCoverage  `json:"executor_coverage"`
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}

func validSHA(value string) bool {
	return shaPattern.MatchString(value)
}

func validDigest(value string) bool {
	return digestPattern.MatchString(value)
}

func validLedgerDigest(value string) bool {
	return ledgerPattern.MatchString(value)
}

func validFileDigests(in inputs) bool {
	return validDigest(in.Metrics.FileSHA256) && validDigest(in.Plan.FileSHA256) &&
		validDigest(in.Execution.FileSHA256) && validDigest(in.Receipts.FileSHA256) &&
		validDigest(in.Provenance.FileSHA256) && validDigest(in.Contract.FileSHA256)
}

func validContractCoverage(contract contractDocument) bool {
	if len(contract.ExecutorCoverage) != len(generation.DefaultRegistry()) {
		return false
	}
	for _, coverage := range contract.ExecutorCoverage {
		if !coverage.Covered {
			return false
		}
	}
	return true
}
