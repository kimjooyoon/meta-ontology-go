package denominatorevolutionverify

import (
	"strconv"
	"strings"
)

func verifyCheck(id, proof, operation, stage, step string, ok bool, expected, observed string) Check {
	state := "FAIL"
	if ok {
		state = "PASS"
	}
	return Check{ID: id, Status: state, ProofChoice: proof, MetaOperation: operation, Coordinate: Coordinate{Stage: stage, Step: step, Reason: expected}, Expected: expected, Observed: observed}
}

func hasFailure(values []Check) bool {
	for _, value := range values {
		if value.Status != "PASS" {
			return true
		}
	}
	return false
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

func receiptWrites(value *Receipt) string {
	if value == nil {
		return "snapshot-bound"
	}
	return intText(value.RepositoryWrites)
}

func receiptText(value *Receipt) string {
	if value == nil {
		return "missing"
	}
	if value.Decision == "" {
		return "material-only"
	}
	return value.Decision + " / " + value.Reason
}

func guardrailText(value *Guardrail) string {
	if value == nil {
		return "missing"
	}
	return value.ID + " direction=" + value.Direction + " observed=" + intText(value.Observed) + " allowed_max=" + intText(value.AllowedMax) + " conformance=" + intText(value.ConformanceNumerator) + "/" + intText(value.ConformanceDenominator) + " conforms=" + boolText(value.Conforms)
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func intText(value int) string { return strconv.Itoa(value) }

func snapshotText(value RepositorySnapshot) string {
	return value.BeforeDigest + " -> " + value.AfterDigest + " changed_paths=" + intText(value.ChangedPaths)
}

func sourceProjectionText(value SourceProjection) string {
	return "entities=" + intText(value.EntityCount) + " activities=" + intText(value.ActivityCount) + " obligations=" + intText(value.ObligationCount) + " cases=" + intText(value.CaseCount)
}

func sourceErrorText(value SourceProjection, err error) string {
	if err != nil {
		return err.Error()
	}
	return sourceProjectionText(value)
}
