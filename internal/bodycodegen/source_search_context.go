package bodycodegen

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// SourceIRSearchContextExport describes a source-owned grammar before candidate
// enumeration, residual probes, scoring or model work. It is not a model encoding.
type SourceIRSearchContextExport struct {
	Schema               string             `json:"schema"`
	OriginalSourceSHA256 string             `json:"original_source_sha256"`
	ContractSHA256       string             `json:"contract_sha256"`
	ActivityID           string             `json:"activity_id"`
	ExpandedPlan         *assemblyspec.Spec `json:"expanded_plan,omitempty"`
	ModelPredictions     int                `json:"model_predictions"`
	CandidateTests       int                `json:"candidate_tests"`
	Scope                string             `json:"scope"`
}

func ExportSourceIRSearchContext(ctx context.Context, filename string, source []byte,
	activity string, includePlan bool) (SourceIRSearchContextExport, error) {
	var result SourceIRSearchContextExport
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return result, err
	}
	plan, err := sourceIRSearchPlan(spec)
	if err != nil {
		return result, err
	}
	if err = validateIRBodySearchPlan(plan); err != nil {
		return result, err
	}
	_, _, activityID, _, err := prepareBodySearch(filename, source, activity, plan.HoleID)
	if err != nil {
		return result, err
	}
	contract, err := spec.Canonical()
	if err != nil {
		return result, err
	}
	result = SourceIRSearchContextExport{Schema: "gooo/source-search-input-export/v1",
		OriginalSourceSHA256: digest(source), ContractSHA256: digest([]byte(contract)), ActivityID: activityID,
		Scope: "source contract, integer profile and hole validation only; no candidate enumeration, residual probes, candidate typechecking, model encoding, predictions or outcomes"}
	if includePlan {
		result.ExpandedPlan = spec.Clone()
	}
	return result, ctx.Err()
}
