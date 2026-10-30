package generation

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/gosplit"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

type PrerequisiteState string

const (
	PrerequisiteObserved   PrerequisiteState = "OBSERVED"
	PrerequisiteUnmet      PrerequisiteState = "UNMET"
	PrerequisiteUnknown    PrerequisiteState = "UNKNOWN"
	PrerequisiteUnobserved PrerequisiteState = "UNOBSERVED"
)

type GoSplitRequirement struct {
	OperationInputContractObligationEvidence
	Origin sourcebundle.DeclarationOrigin
	State  PrerequisiteState
}

// GoSplitPreparation describes current-host readonly observations. Contract
// facts describe the authored chain; requirement states describe this file.
// Projected package conformance remains unobserved until its actual consumer.
type GoSplitPreparation struct {
	Contract     OperationInputContractEvidence
	PolicySource sourcebundle.Member
	Subject      string
	SourceSHA256 string
	BlockerKind  gosplit.BlockerKind
	Requirements [len(nativeGoSplitObligations)]GoSplitRequirement
	Plan         *gosplit.Plan
}

func (preparation GoSplitPreparation) ReadyForProjection() bool {
	return preparation.Plan != nil && preparation.NextStep().F_ready_for_projection
}

// NextStep is projected from go-split-selection.gooo. It suggests a prerequisite
// operation and never authorizes a source write or claims package conformance.
func (preparation GoSplitPreparation) NextStep() R_SplitNext {
	return Evaluate(R_SplitObservations{
		F_rendered_state:  string(preparation.Requirements[0].State),
		F_directory_state: string(preparation.Requirements[1].State),
		F_native_blocker:  string(preparation.BlockerKind),
	})
}

// InspectGoSplit does not collect metrics or write source. Supplied directory
// observations retain their own scope; SourceSHA256 binds only parsed bytes.
func InspectGoSplit(root, subject string, report metricevidence.Report) (GoSplitPreparation, error) {
	contract, err := GoSplitInputContractEvidence()
	if err != nil {
		return GoSplitPreparation{}, err
	}
	result := newGoSplitPreparation(subject, contract)
	plan, err := NewGoSplitPlanner().PlanSource(root, subject, report.Meta.Policy.MaxFileLines)
	if err != nil {
		result.Requirements[0].State = PrerequisiteUnknown
		if blocker, ok := errors.AsType[*gosplit.Blocker](err); ok {
			result.SourceSHA256 = blocker.SourceSHA256
			result.BlockerKind = blocker.Kind
			result.Requirements[0].State = PrerequisiteUnmet
		}
		return result, err
	}
	result.SourceSHA256 = plan.SourceSHA256
	result.Requirements[0].State = PrerequisiteObserved
	if err := gosplit.ValidateTopology(report, plan); err != nil {
		result.Requirements[1].State = PrerequisiteUnknown
		if blocker, ok := errors.AsType[*gosplit.Blocker](err); ok {
			result.BlockerKind = blocker.Kind
			if blocker.Kind != gosplit.DirectoryUnobserved {
				result.Requirements[1].State = PrerequisiteUnmet
			}
		}
		return result, err
	}
	result.Requirements[1].State = PrerequisiteObserved
	result.Plan = &plan
	return result, nil
}

func newGoSplitPreparation(subject string, contract OperationInputContractEvidence) GoSplitPreparation {
	result := GoSplitPreparation{Subject: subject, Contract: contract, PolicySource: goSplitSelectionSourceOrigin()}
	for index, native := range nativeGoSplitObligations {
		result.Requirements[index] = GoSplitRequirement{
			OperationInputContractObligationEvidence: contract.Obligations[index],
			State:                                    PrerequisiteUnobserved,
		}
		for _, origin := range contract.SourceOrigins.Declarations {
			if origin.ID == native.ActivityID {
				result.Requirements[index].Origin = origin
				break
			}
		}
	}
	return result
}
