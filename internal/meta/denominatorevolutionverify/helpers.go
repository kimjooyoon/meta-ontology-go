package denominatorevolutionverify

import "strconv"

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
func intText(value int) string { return strconv.Itoa(value) }
func receiptText(value *Receipt) string {
	if value == nil {
		return "missing"
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
