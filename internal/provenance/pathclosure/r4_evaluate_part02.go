package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// EvaluateR4 recomputes only the declared finite paths. It never discovers
// paths, accepts no expected/display label in the R4 input, and never emits
// promotion authorization.
func EvaluateR4(input R4Input) R4Result {
	required := sortedR4IDs(input.Boundary.RequiredPathIDs)
	cost := len(input.Records) + len(input.Receipts) + len(input.Paths)
	if input.Schema != R4SchemaVersion {
		return r4Fail(CodeInvalidPath, "invalid R4 schema", required, cost)
	}
	if duplicate := duplicateR4IDs(required); duplicate != "" {
		return r4Fail(CodeInvalidPath, "duplicate required path "+duplicate.String(), required, cost)
	}
	if len(required) == 0 {
		return r4Unknown(CodeMissingRequiredPaths, "no finite required paths were declared", required, cost)
	}
	if input.Boundary.OpenWorld {
		return r4Unknown(CodeOpenWorld, "open-world path closure is not a finite proof boundary", required, cost)
	}
	if !input.Boundary.Exhausted {
		return r4Unknown(CodeUnexhaustedBoundary, "finite path boundary was not explicitly exhausted", required, cost)
	}

	records, recordErr := indexR4Records(input.Records)
	if recordErr != nil {
		return r4Fail(CodeInvalidPath, recordErr.Error(), required, cost)
	}
	receipts, receiptErr := indexR4Receipts(input.Receipts)
	if receiptErr != nil {
		return r4Fail(CodeConflictingReceipt, receiptErr.Error(), required, cost)
	}
	paths, pathErr := indexR4Paths(input.Paths)
	if pathErr != nil {
		return r4Fail(CodeInvalidPath, pathErr.Error(), required, cost)
	}

	for _, pathID := range required {
		path, exists := paths[pathID]
		if !exists {
			return r4Unknown(CodeMissingRecord, "missing required path "+pathID.String(), required, cost)
		}
		status, code, reason := evaluateR4Path(path, records, receipts)
		if status == FAIL_CLOSED {
			return r4Fail(code, reason, required, cost)
		}
		if status == UNKNOWN {
			return r4Unknown(code, reason, required, cost)
		}
	}
	result := r4Result(PASS, CodeR4ProofValid, "all finite declared paths are covered", required, cost)
	result.CoveredPathIDs = append([]semantic.ID(nil), required...)
	result.ProofValid = true

	result.PromotionAuthorized = false
	return result
}
