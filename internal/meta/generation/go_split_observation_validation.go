package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/gosplit"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func validateGoSplitObservations(plan Plan) error {
	if len(plan.GoSplitPreparations) == 0 {
		if plan.GoSplitPolicySHA256 != "" {
			return fmt.Errorf("split policy has no current-file observations")
		}
		return nil
	}
	if plan.GoSplitPolicySHA256 != goSplitSelectionSourceOrigin().SHA256 || plan.IndicatorDecisionLedger == nil ||
		plan.Decision == DecisionFixedPoint {
		return fmt.Errorf("split observations lack their current policy or metric population")
	}
	sources, err := refutedSourceIndicators(plan)
	if err != nil {
		return err
	}
	seen := map[string]GoSplitObservation{}
	previous := ""
	for _, observation := range plan.GoSplitPreparations {
		source, exists := sources[observation.IndicatorID]
		if !exists || source.Satisfied || source.Operation != sourcepolicy.OperationSplitGo ||
			source.SubjectKind != sourcepolicy.SubjectKindFile || observation.IndicatorID <= previous ||
			!validGoSplitObservation(observation) {
			return fmt.Errorf("split observation differs from its unresolved source indicator")
		}
		seen[observation.IndicatorID], previous = observation, observation.IndicatorID
	}
	for _, action := range plan.Selected {
		if action.Operation == sourcepolicy.OperationSplitGo {
			observation, exists := seen[action.IndicatorID]
			if !exists || !observation.readyForProjection() {
				return fmt.Errorf("selected split has no satisfied preparation")
			}
		}
	}
	return nil
}

func validGoSplitObservation(observed GoSplitObservation) bool {
	if !validActionIndicatorID(observed.IndicatorID) || observed.States[2] != PrerequisiteUnobserved {
		return false
	}
	first, next := observed.States[0], observed.States[1]
	if observed.BlockerKind == "" && observed.Cause != "" {
		return first == PrerequisiteUnknown && next == PrerequisiteUnobserved && observed.SourceSHA256 == ""
	}
	if !validDigest(observed.SourceSHA256) {
		return false
	}
	switch observed.BlockerKind {
	case "":
		return observed.Cause == "" && first == PrerequisiteObserved && next == PrerequisiteObserved
	case gosplit.RenderedPlanUnavailable, gosplit.NoDeclarationSplit:
		return observed.Cause != "" && first == PrerequisiteUnmet && next == PrerequisiteUnobserved
	case gosplit.DirectoryFull, gosplit.DirectoryMixed:
		return observed.Cause != "" && first == PrerequisiteObserved && next == PrerequisiteUnmet
	case gosplit.DirectoryUnobserved:
		return observed.Cause != "" && first == PrerequisiteObserved && next == PrerequisiteUnknown
	}
	return false
}
