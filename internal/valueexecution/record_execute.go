package valueexecution

import (
	"maps"
	"slices"
	"unicode/utf8"
)

type RecordExecution struct {
	Scope      string                    `json:"scope"`
	ApplyCalls int                       `json:"apply_calls"`
	Deliveries int                       `json:"deliveries"`
	Activities []string                  `json:"activities"`
	Results    map[string]RecordEvidence `json:"results"`
}

func (plan RecordPlan) Execute(rootInputs map[string]RecordFields) (RecordExecution, error) {
	execution := RecordExecution{Scope: RecordTransportScope, Results: map[string]RecordEvidence{}}
	if err := plan.validateAuthority(); err != nil {
		return execution, err
	}
	if err := plan.validateInputs(rootInputs); err != nil {
		return execution, err
	}
	values := map[string]ProducedRecord{}
	for _, name := range plan.order {
		program := plan.programs[name]
		fields, rootActivity, rootDigest, parentDigest, err := plan.recordInput(name, rootInputs, values)
		if err != nil {
			return execution, err
		}
		result := issueProducedRecord(program.authority, fields, rootActivity, rootDigest, parentDigest)
		if err := result.validate(); err != nil {
			return execution, err
		}
		values[name] = result
		execution.ApplyCalls++
		execution.Activities = append(execution.Activities, name)
		if parentDigest != "" {
			execution.Deliveries++
		}
		execution.Results[name] = result.Evidence()
	}
	return execution, nil
}

func (plan RecordPlan) validateInputs(inputs map[string]RecordFields) error {
	names := make([]string, 0, len(inputs))
	for name := range inputs {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if _, found := plan.programs[name]; !found {
			return failAt(ReasonExternalInputUnexpected, "INPUT", "reject-unknown-record-root", name)
		}
		if _, bound := plan.incoming[name]; bound {
			return failAt(ReasonExternalInputUnexpected, "INPUT", "reject-bound-record-root", name)
		}
	}
	for _, name := range plan.order {
		if _, bound := plan.incoming[name]; bound {
			continue
		}
		fields, found := inputs[name]
		if !found {
			return failAt(ReasonExternalInputMissing, "INPUT", "require-record-root", name)
		}
		if err := validateRecordFields(plan.programs[name].fields, fields); err != nil {
			return err
		}
	}
	return nil
}

func validateRecordFields(schema []RecordField, fields RecordFields) error {
	if fields == nil || len(fields) != len(schema) {
		return failAt(ReasonSignatureTypeMismatch, "TYPECHECK", "validate-record-fields", "record must contain exactly the declared fields")
	}
	for _, field := range schema {
		value, present := fields[field.Name]
		if !present || !utf8.ValidString(value) {
			return failAt(ReasonSignatureTypeMismatch, "TYPECHECK", "validate-record-fields", field.Name)
		}
	}
	return nil
}

func (plan RecordPlan) recordInput(name string, roots map[string]RecordFields, values map[string]ProducedRecord) (RecordFields, string, string, string, error) {
	producer, bound := plan.incoming[name]
	if !bound {
		fields := maps.Clone(roots[name])
		return fields, name, digestValue(fields), "", nil
	}
	result, found := values[producer]
	if !found || result.validate() != nil || result.authority != plan.programs[producer].authority ||
		result.authority.outputEntityID != plan.programs[name].authority.outputEntityID {
		return nil, "", "", "", failAt(ReasonBindingResultInvalid, "EXECUTE", "validate-bound-record", name)
	}
	return maps.Clone(result.fields), result.rootActivity, result.rootDigest, result.digest, nil
}
