package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/gosplit"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// GoSplitObservation is the compact current-file observation retained in a
// generation plan. The registry already binds the prerequisite contract; the
// separate policy hash names the Gooo program selecting the next step.
type GoSplitObservation struct {
	IndicatorID  string                                           `json:"indicator_id"`
	SourceSHA256 string                                           `json:"source_sha256"`
	States       [len(nativeGoSplitObligations)]PrerequisiteState `json:"states"`
	BlockerKind  gosplit.BlockerKind                              `json:"blocker_kind,omitempty"`
	Cause        string                                           `json:"cause,omitempty"`
}

type splitPrerequisiteResolver func(sourcepolicy.Indicator) GoSplitObservation

func (observation GoSplitObservation) NextStep() R_SplitNext {
	return Evaluate(R_SplitObservations{F_rendered_state: string(observation.States[0]),
		F_directory_state: string(observation.States[1]), F_native_blocker: string(observation.BlockerKind)})
}

func (observation GoSplitObservation) readyForProjection() bool {
	return observation.NextStep().F_ready_for_projection && validDigest(observation.SourceSHA256)
}

// BuildWithSourcePrerequisites inspects only eligible split candidates during
// selection. It preserves every raw indicator and never runs a rewrite or a
// projected package test during metric collection.
func BuildWithSourcePrerequisites(root, baseSHA, headSHA string, report sourcepolicy.Report,
	topology metricevidence.Report, inputs OperationInputs) (Plan, error) {
	if topology.CommitSHA != headSHA || topology.Meta.Policy.MaxFileLines != report.Policy.MaxFileLines ||
		topology.Meta.Policy.MaxDirectEntries != report.Policy.MaxDirectDirectoryIn || report.Policy.Validate() != nil {
		return Plan{}, fmt.Errorf("split prerequisite metrics do not bind the source revision and policy")
	}
	resolve := func(indicator sourcepolicy.Indicator) GoSplitObservation {
		preparation, err := InspectGoSplit(root, indicator.Subject, topology)
		observed := GoSplitObservation{IndicatorID: indicatorID(indicator), SourceSHA256: preparation.SourceSHA256,
			BlockerKind: preparation.BlockerKind}
		for index, requirement := range preparation.Requirements {
			observed.States[index] = requirement.State
		}
		if err != nil {
			observed.Cause = err.Error()
		}
		return observed
	}
	return buildOperationInputs(baseSHA, headSHA, report, inputs, resolve), nil
}
