package pressureshadow

import (
	"testing"
)

func TestValidateRequiredVectors(t *testing.T) {
	cases := []pathVector{
		{
			name: "empty selector paths", mutate: func(input *Input) {
				input.Selector.Paths, input.PathCoverage = nil, nil
			}, decision: DecisionUnknown, reason: ReasonRequiredInputMissing,
		},
		{
			name: "missing row", mutate: func(input *Input) {
				input.PathCoverage = input.PathCoverage[:1]
			}, decision: DecisionUnknown, reason: ReasonMissingPathCoverage,
			missing: []string{"path/a"},
		},
		{
			name: "orphan row", mutate: func(input *Input) {
				row := input.PathCoverage[0]
				row.PathID = "path/orphan"
				input.PathCoverage = append(input.PathCoverage, row)
			}, decision: DecisionFailClosed, reason: ReasonOrphanPathCoverage,
			orphan: []string{"path/orphan"},
		},
		{
			name: "blank selector tuple component", mutate: func(input *Input) {
				input.Selector.SnapshotDigest = ""
			}, decision: DecisionUnknown, reason: ReasonRequiredInputMissing,
			missingBinding: []string{"path/a", "path/b"},
		},
		{
			name: "blank row tuple component", mutate: func(input *Input) {
				input.PathCoverage[0].PolicyDigest = ""
			}, decision: DecisionUnknown, reason: ReasonRequiredInputMissing,
			missingBinding: []string{"path/b"},
		},
	}
	runPathVectors(t, cases)
}
func TestValidatePrecedenceVectors(t *testing.T) {
	cases := []pathVector{
		{
			name: "binding mismatch", mutate: func(input *Input) {
				input.PathCoverage[0].RegistryDigest = "stale"
			}, decision: DecisionUnknown, reason: ReasonBindingMismatch,
			mismatch: []string{"path/b"},
		},
		{
			name: "mixed precedence", mutate: func(input *Input) {
				input.PathCoverage = input.PathCoverage[:1]
				row := input.PathCoverage[0]
				row.PathID = "path/orphan"
				input.PathCoverage = append(input.PathCoverage, row)
				input.PathCoverage[0].PolicyDigest = ""
				input.PathCoverage[0].RegistryDigest = "stale"
			}, decision: DecisionFailClosed, reason: ReasonOrphanPathCoverage,
			missing: []string{"path/a"}, orphan: []string{"path/orphan"},
			missingBinding: []string{"path/b"}, mismatch: []string{"path/b"},
		},
		{
			name: "invalid A1 syntax", mutate: func(input *Input) {
				input.Selector.Paths[0].StableID = "path a"
			}, decision: DecisionFailClosed, reason: ReasonInvalidInput,
		},
	}
	runPathVectors(t, cases)
}
