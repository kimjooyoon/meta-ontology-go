package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var (
	shaPattern    = regexp.MustCompile("^[0-9a-f]{40}$")
	digestPattern = regexp.MustCompile("^[0-9a-f]{64}$")
	ledgerPattern = regexp.MustCompile("^sha256:[0-9a-f]{64}$")
)

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
