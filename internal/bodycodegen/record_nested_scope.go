package bodycodegen

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

func bindNominalRecordAssemblyScope(common *CompletenessReceipt, result Report) {
	if !recordTypesHaveNominalFields(result.RecordTypes) {
		return
	}
	common.Scope["domain_scope"] = "one source-owned pure record assembly with 1..16 typed inputs and bounded scalar or nominal record fields"
	common.Scope["record_body_scope"] = "bounded immutable required and optional record values and ordered collections; source-declared finite field synthesis with explicit presence"
	common.Scope["nominal_record_limits"] = map[string]any{
		"reference_depth":     syntax.RecordReferenceDepthLimit,
		"logical_value_slots": syntax.RecordValueSlotLimit,
		"selected_types":      recordTypeLimit,
		"fields_per_record":   RecordFieldLimit,
		"collection_elements": RepeatedFieldLimit,
		"default_values":      "shared immutable nominal layout DAG; each field/index occurrence is charged",
	}
	common.NotClaimed = append(common.NotClaimed,
		"full dynamic nested-field dependency provenance",
		"direct nested record field assignment",
		"identity comparison between two present optional pointers")
}
