package bodycodegen

import (
	"context"
	"fmt"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type RecordAssemblyContextExport struct {
	Schema               string                    `json:"schema"`
	OriginalSourceSHA256 string                    `json:"original_source_sha256"`
	ContractSHA256       string                    `json:"contract_sha256"`
	ActivityID           string                    `json:"activity_id"`
	Choices              []RecordValueChoice       `json:"choices"`
	Context              *RecordOrdinalContext     `json:"context"`
	ValueFlow            *RecordValueFlow          `json:"value_flow,omitempty"`
	ExpandedPlan         *assemblyspec.Spec        `json:"expanded_plan,omitempty"`
	ModelCompatibility   *RecordModelCompatibility `json:"model_compatibility,omitempty"`
	ModelPredictions     int                       `json:"model_predictions"`
	CandidateTests       int                       `json:"candidate_tests"`
	Scope                string                    `json:"scope"`
}

func ExportRecordAssemblyContext(ctx context.Context, filename string, source []byte, activity string, includePlan bool) (RecordAssemblyContextExport, error) {
	return ExportRecordAssemblyContextWithFeature(ctx, filename, source, activity, includePlan, "")
}

func ExportRecordAssemblyContextWithFeature(ctx context.Context, filename string, source []byte, activity string,
	includePlan bool, version string) (RecordAssemblyContextExport, error) {
	return exportRecordAssemblyContext(ctx, filename, source, activity, includePlan, version, false)
}

// ExportRecordAssemblyContextWithFlow adds source-derived definitions without
// predicting, evaluating cases or changing an existing model feature contract.
func ExportRecordAssemblyContextWithFlow(ctx context.Context, filename string, source []byte, activity string,
	includePlan bool, version string) (RecordAssemblyContextExport, error) {
	return exportRecordAssemblyContext(ctx, filename, source, activity, includePlan, version, true)
}

func exportRecordAssemblyContext(ctx context.Context, filename string, source []byte, activity string,
	includePlan bool, version string, includeFlow bool) (RecordAssemblyContextExport, error) {
	if version != "" && version != decision.SplitContextIntentFeatureVersion &&
		version != decision.SemanticContextIntentFeatureVersion && version != jointdecision.RecordFieldFeatureVersion &&
		version != jointdecision.RecordSharedFeatureVersion && version != jointdecision.RecordOriginSharedFeatureVersion &&
		version != jointdecision.RecordGraphSharedFeatureVersion {
		return RecordAssemblyContextExport{}, fmt.Errorf("unsupported record context feature version")
	}
	plan, err := prepareRecordAssembly(ctx, filename, source, activity)
	if err != nil {
		return RecordAssemblyContextExport{}, err
	}
	contract, _ := plan.spec.Canonical()
	var flow *RecordValueFlow
	if includeFlow || version == jointdecision.RecordOriginSharedFeatureVersion || version == jointdecision.RecordGraphSharedFeatureVersion {
		flow = recordValueFlow(plan)
	}
	var modelContext *RecordOrdinalContext
	if version == jointdecision.RecordOriginSharedFeatureVersion {
		modelContext = recordOriginContext(plan.choices, flow)
	} else if version == jointdecision.RecordGraphSharedFeatureVersion {
		modelContext = recordGraphContext(plan, flow)
	} else {
		modelContext = recordModelContext(plan.choices, version)
	}
	result := RecordAssemblyContextExport{Schema: "gooo/record-assembly-input-export/v1", OriginalSourceSHA256: digest(source),
		ContractSHA256: digest([]byte(contract)), ActivityID: plan.body.activityID, Choices: append([]RecordValueChoice(nil), plan.choices...),
		Context: modelContext, Scope: "explicit source-bound model context; shape and type preflight; zero predictions or candidate outcomes; optional expanded plan carries the separate finite cases"}
	if includePlan {
		result.ExpandedPlan = plan.spec.Clone()
	}
	if includeFlow {
		result.ValueFlow = flow
	}
	return result, ctx.Err()
}
