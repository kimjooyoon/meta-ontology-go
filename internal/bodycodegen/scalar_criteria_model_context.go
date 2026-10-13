package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func exportScalarCriteriaModelContext(ctx context.Context, filename string, source []byte, activity string,
	info RetainedModelInfo, options RecordModelContextOptions) (RecordAssemblyContextExport, error) {
	if options.FeatureVersion != "" && options.FeatureVersion != info.FeatureVersion {
		return RecordAssemblyContextExport{}, fmt.Errorf("explicit feature version differs from scalar model feature %q", info.FeatureVersion)
	}
	plan, err := prepareRecordAssembly(ctx, filename, source, activity)
	if err != nil {
		return RecordAssemblyContextExport{}, err
	}
	features, sourceContext, err := scalarCriteriaModelInput(ctx, plan)
	if err != nil {
		return RecordAssemblyContextExport{}, err
	}
	if !assemblyspec.IsScalarCriteria(plan.spec) {
		return RecordAssemblyContextExport{}, fmt.Errorf("scalar source context requires Integer choices and fixed predicates")
	}
	contract, err := plan.spec.Canonical()
	if err != nil {
		return RecordAssemblyContextExport{}, err
	}
	r := RecordAssemblyContextExport{Schema: "gooo/scalar-criteria-input-export/v1",
		OriginalSourceSHA256: plan.sourceSHA256, ContractSHA256: digest([]byte(contract)),
		ActivityID: plan.body.activityID, Choices: append([]RecordValueChoice(nil), plan.choices...),
		Context: sourceContext, ScalarFeatures: features,
		Scope: "source scalar model input and artifact compatibility; no prediction, finite evaluation, selected emission or writes"}
	if options.IncludePlan {
		r.ExpandedPlan = plan.spec.Clone()
	}
	if options.ValueFlow {
		r.CriteriaFlow = criteriaSourceFlow(ctx, filename, source, plan)
	}
	status, reason := sourceContext.Status, sourceContext.Reason
	if status == "ENCODED" {
		status, reason = "READY_FOR_RANKING", "SOURCE_INPUT_ENCODED"
	}
	r.ModelCompatibility = &RecordModelCompatibility{Schema: "gooo/record-model-compatibility/v1", Status: status,
		Reason: reason, Model: info, Scope: "loaded scalar artifact and source-only numerical representation; model quality requires separate observations"}
	return r, ctx.Err()
}
