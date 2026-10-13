package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/criteriainput"
)

// The carrier binds finite declarations by their contract digest. Numerical rows
// read only source choices and fixed predicate graphs; case values and outcomes
// are absent. Each invocation derives it anew from the supplied source.
func scalarCriteriaModelInput(ctx context.Context, plan recordAssemblyPlan) (*criteriainput.Features, *RecordOrdinalContext, error) {
	r := &RecordOrdinalContext{Schema: "gooo/compiler-scalar-criteria-context/v1", Status: "DECLINED_TO_DETERMINISTIC",
		FeatureVersion: criteriainput.FeatureSchema,
		Scope:          "source-owned scalar choices and fixed predicate graphs; finite inputs, expected outputs and observations excluded from numerical features"}
	if ctx == nil {
		return nil, r, fmt.Errorf("scalar source model context required")
	}
	if err := ctx.Err(); err != nil {
		return nil, r, err
	}
	if plan.scalar == nil {
		r.Reason = "SCALAR_CRITERIA_SOURCE_REQUIRED"
		return nil, r, nil
	}
	if plan.sourceSHA256 != digest(plan.sourceCarrier) {
		return nil, r, fmt.Errorf("scalar model source carrier differs from prepared source")
	}
	contract, err := plan.spec.Canonical()
	if err != nil {
		return nil, r, err
	}
	export := RecordAssemblyContextExport{Schema: "gooo/scalar-criteria-input-export/v1",
		OriginalSourceSHA256: plan.sourceSHA256, ContractSHA256: digest([]byte(contract)),
		ActivityID:   plan.body.activityID,
		CriteriaFlow: criteriaSourceFlow(ctx, plan.sourceFilename, plan.sourceCarrier, plan)}
	raw, err := json.Marshal(export)
	if err != nil {
		return nil, r, err
	}
	prepared, err := criteriainput.FromCompilerExport(raw)
	if err != nil {
		r.Reason = err.Error()
		return nil, r, ctx.Err()
	}
	features := new(criteriainput.Features)
	if err := prepared.FeaturesInto(features); err != nil {
		r.Reason = err.Error()
		return nil, r, ctx.Err()
	}
	r.Text, r.SHA256, r.Status = string(prepared.Canonical()), features.InputSHA256, "ENCODED"
	return features, r, ctx.Err()
}
