package denominatorevolution

import (
	"strconv"
	"strings"
)

func check(id, proof, operation string, coordinate Coordinate, state, expected, observed string) CheckResult {
	return CheckResult{ID: id, Status: state, ProofChoice: proof, MetaOperation: operation, Coordinate: coordinate, Expected: expected, Observed: observed}
}

func receiptStatus(predKnown, receiptValid bool) string {
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
	if receipt.Decision == "" {
		return "material-only"
	}
	return receipt.Decision + "/" + receipt.Reason
}

func status(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}

func claimProof(decision string) string {
	if decision == "FAIL_CLOSED" {
		return "FOUNDATION"
	}
	if decision == "ADVANCE" {
		return "COHERENCE"
	}
	return "REGRESSION"
}

func claimOperation(decision string) string {
	if decision == "FAIL_CLOSED" {
		return "fail-closed-unknown-predecessor"
	}
	if decision == "ADVANCE" {
		return "accept-authorized-denominator-advance"
	}
	return "reject-invalid-denominator-change"
}

func changeText(values []Change) string {
	if len(values) == 0 {
		return "none"
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ObligationID+"/"+value.Reason)
	}
	return strings.Join(result, ",")
}

func repositoryWritesText(receipt *MigrationReceipt) string {
	if receipt == nil {
		return "snapshot-bound"
	}
	return strconv.Itoa(receipt.RepositoryWrites)
}

func boolText(value bool) string { return strconv.FormatBool(value) }

func hasUnsatisfied(values []Indicator) bool {
	for _, value := range values {
		if !value.Satisfied {
			return true
		}
	}
	return false
}
