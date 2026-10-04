package bodycodegen

import (
	"encoding/json"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func recordModelContext(choices []RecordValueChoice, version string) *RecordOrdinalContext {
	if version != jointdecision.RecordFieldFeatureVersion && version != jointdecision.RecordSharedFeatureVersion {
		return recordOrdinalContext(choices)
	}
	r := &RecordOrdinalContext{Schema: "gooo/record-field-expression-context/v1", Status: "ENCODED",
		FeatureVersion: version,
		Scope:          "complete source-bound field names, ordered expressions and intent; actual string-field structural features; cases and expected outputs excluded"}
	var input [3]jointdecision.RecordChoice
	for i, choice := range choices {
		part := jointdecision.RecordChoice{Field: choice.Field, First: choice.First, Second: choice.Second, Intent: choice.Intent}
		raw, _ := json.Marshal(part)
		r.Parts = append(r.Parts, string(raw))
		if i < len(input) {
			input[i] = part
		}
	}
	if len(choices) != len(input) {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "THREE_FIELD_CHOICES_REQUIRED"
		return r
	}
	var err error
	r.Text, err = jointdecision.EncodeRecordThree(input)
	if err != nil {
		r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "COMPLETE_FIELD_CONTEXT_EXCEEDS_MODEL_BOUND"
		return r
	}
	r.SHA256 = digest([]byte(r.Text))
	return r
}
