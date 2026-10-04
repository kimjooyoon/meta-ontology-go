package bodycodegen

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type RecordAssemblyContextExport struct {
	Schema               string                `json:"schema"`
	OriginalSourceSHA256 string                `json:"original_source_sha256"`
	ContractSHA256       string                `json:"contract_sha256"`
	ActivityID           string                `json:"activity_id"`
	Choices              []RecordValueChoice   `json:"choices"`
	Context              *RecordOrdinalContext `json:"context"`
	ExpandedPlan         *assemblyspec.Spec    `json:"expanded_plan,omitempty"`
	ModelPredictions     int                   `json:"model_predictions"`
	CandidateTests       int                   `json:"candidate_tests"`
	Scope                string                `json:"scope"`
}

func ExportRecordAssemblyContext(ctx context.Context, filename string, source []byte, activity string, includePlan bool) (RecordAssemblyContextExport, error) {
	plan, err := prepareRecordAssembly(ctx, filename, source, activity)
	if err != nil {
		return RecordAssemblyContextExport{}, err
	}
	contract, _ := plan.spec.Canonical()
	result := RecordAssemblyContextExport{Schema: "gooo/record-assembly-input-export/v1", OriginalSourceSHA256: digest(source),
		ContractSHA256: digest([]byte(contract)), ActivityID: plan.body.activityID, Choices: append([]RecordValueChoice(nil), plan.choices...),
		Context: recordOrdinalContext(plan.choices), Scope: "source-bound ordinal model context; shape and type preflight; zero predictions or candidate outcomes; optional expanded plan carries the separate finite cases"}
	if includePlan {
		result.ExpandedPlan = plan.spec.Clone()
	}
	return result, nil
}
