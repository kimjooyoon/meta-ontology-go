package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}

func planExecutionDigest(plan Plan) string {
	activities := make([]string, 0, len(plan.programs))
	for activity := range plan.programs {
		activities = append(activities, activity)
	}
	slices.Sort(activities)
	return digestValue(struct {
		SourceDigest        string
		SemanticFingerprint string
		Activities          []string
		BindingCount        int
	}{
		SourceDigest: plan.SourceDigest, SemanticFingerprint: plan.SemanticFingerprint,
		Activities: activities, BindingCount: len(plan.bindings),
	})
}

func executionDigest(execution Execution) string {
	execution.ExecutionDigest = ""
	return digestValue(execution)
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || value[:len("sha256:")] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}
