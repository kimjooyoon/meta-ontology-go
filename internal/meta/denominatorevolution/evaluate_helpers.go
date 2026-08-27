package denominatorevolution

import (
	"strconv"
	"strings"
)

func check(id, proof, operation string, spec CaseSpec, state, expected, observed string) CheckResult {
	return CheckResult{ID: id, Status: state, ProofChoice: proof, MetaOperation: operation, Coordinate: Coordinate{Stage: spec.Stage, Step: spec.Step, Reason: spec.Reason}, Expected: expected, Observed: observed}
}

func receiptStatus(predKnown, receiptValid bool, receipt *MigrationReceipt) string {
	if receiptValid {
		return "PASS"
	}
	if !predKnown {
		return "UNKNOWN"
	}
	return "FAIL"
}

func receiptText(receipt *MigrationReceipt) string {
	if receipt == nil {
		return "missing"
	}
	return receipt.Decision + "/" + receipt.Reason
}

func status(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}
func boolText(value bool) string { return strconv.FormatBool(value) }
func repositoryWritesText(receipt *MigrationReceipt) string {
	if receipt == nil {
		return "not applicable"
	}
	return strconv.Itoa(receipt.RepositoryWrites)
}
func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func containsFold(value, wanted string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(wanted))
}
