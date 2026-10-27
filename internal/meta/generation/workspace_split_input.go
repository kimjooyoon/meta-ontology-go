package generation

import (
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func cloneWorkspaceSplitRequest(request *workspacesplit.Request) *workspacesplit.Request {
	if request == nil {
		return nil
	}
	copy := *request
	return &copy
}

func validWorkspaceSplitIndicator(indicator sourcepolicy.Indicator, request workspacesplit.Request) bool {
	return request.Validate() == nil && indicator.Operation == sourcepolicy.OperationSplitGooo &&
		indicator.SubjectKind == sourcepolicy.SubjectKindFile && indicator.Subject == request.Subject() &&
		indicator.MetricID == sourcepolicy.DimensionGoooFileLines && indicator.Limit == request.LineLimit &&
		indicator.Relation == sourcepolicy.RelationLessOrEqual && indicator.Role == sourcepolicy.IndicatorRoleDriver &&
		indicator.OperationInputDigest == request.Digest()
}

func validWorkspaceSplitActionInput(action Action) bool {
	if action.WorkspaceSplitRequest == nil {
		// An unbound logical plan can still describe pressure. Execution reports
		// the missing consumer scope explicitly; it cannot install a partition.
		return action.SourceIndicator.OperationInputDigest == ""
	}
	return validWorkspaceSplitIndicator(action.SourceIndicator, *action.WorkspaceSplitRequest)
}

// BindWorkspaceSplitInputs binds existing observed metrics, never manufactures
// a satisfied/unsatisfied line count. Each supplied request must match one driver.
func BindWorkspaceSplitInputs(report sourcepolicy.Report, requests []workspacesplit.Request) (sourcepolicy.Report, OperationInputs, error) {
	report.Indicators = slices.Clone(report.Indicators)
	inputs := OperationInputs{WorkspaceSplits: map[string]workspacesplit.Request{}}
	seen := map[string]bool{}
	for _, request := range requests {
		if err := request.Validate(); err != nil {
			return report, inputs, err
		}
		if seen[request.Subject()] {
			return report, inputs, fmt.Errorf("duplicate workspace split subject: %s", request.Subject())
		}
		seen[request.Subject()] = true
		matched := 0
		for index := range report.Indicators {
			indicator := &report.Indicators[index]
			if indicator.Operation != sourcepolicy.OperationSplitGooo || indicator.Subject != request.Subject() {
				continue
			}
			if indicator.OperationInputDigest != "" && indicator.OperationInputDigest != request.Digest() {
				return report, inputs, fmt.Errorf("workspace split metric already binds a different request")
			}
			indicator.OperationInputDigest = request.Digest()
			if !validWorkspaceSplitIndicator(*indicator, request) {
				return report, inputs, fmt.Errorf("workspace split request differs from its observed metric")
			}
			matched++
		}
		if matched != 1 {
			return report, inputs, fmt.Errorf("workspace split request matched %d metric drivers", matched)
		}
		inputs.WorkspaceSplits[request.Digest()] = request
	}
	return report, inputs, nil
}
