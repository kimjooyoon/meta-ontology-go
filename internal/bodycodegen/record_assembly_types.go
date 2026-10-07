package bodycodegen

import (
	"encoding/json"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

const recordAssemblySchema = "gooo/record-field-assembly/v1"

type RecordAssemblyReceipt struct {
	Schema               string                         `json:"schema"`
	OriginalSourceSHA256 string                         `json:"original_source_sha256"`
	SelectedSourceSHA256 string                         `json:"selected_source_sha256"`
	ContractSHA256       string                         `json:"contract_sha256"`
	TestSuiteSHA256      string                         `json:"test_suite_sha256"`
	SelectedMask         uint16                         `json:"selected_mask"`
	Choices              []RecordValueChoice            `json:"choices"`
	Attempts             []RecordAssemblyAttempt        `json:"attempts"`
	Cases                []RecordAssemblyCase           `json:"cases"`
	Passed               int                            `json:"passed"`
	Total                int                            `json:"total"`
	FieldsPassed         int                            `json:"fields_passed"`
	FieldsTotal          int                            `json:"fields_total"`
	Status               string                         `json:"status"`
	ModelRequested       bool                           `json:"model_requested"`
	ModelCalls           int                            `json:"model_calls"`
	PredictNS            int64                          `json:"predict_ns"`
	Model                *RetainedModelInfo             `json:"model,omitempty"`
	Context              *RecordOrdinalContext          `json:"model_context,omitempty"`
	Prediction           *jointdecision.ThreePrediction `json:"prediction,omitempty"`
	Ranking              []uint16                       `json:"ranking"`
	GenerationNS         int64                          `json:"generation_ns"`
	Scope                string                         `json:"scope"`
}

type RecordValueChoice struct {
	ID         string `json:"id"`
	Kind       string `json:"kind,omitempty"`
	RecordID   string `json:"record_id"`
	FieldID    string `json:"field_id"`
	Field      string `json:"field"`
	TypeID     string `json:"type_id,omitempty"`
	Presence   string `json:"presence,omitempty"`
	Occurrence int    `json:"occurrence"`
	Intent     string `json:"intent"`
	First      string `json:"first"`
	Second     string `json:"second"`
	Picked     string `json:"picked"`
}

type RecordAssemblyAttempt struct {
	Mask         uint16 `json:"mask"`
	Status       string `json:"status,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Passed       int    `json:"passed"`
	Total        int    `json:"total"`
	FieldsPassed int    `json:"fields_passed"`
	FieldsTotal  int    `json:"fields_total"`
}

type RecordAssemblyCase struct {
	Inputs   json.RawMessage       `json:"inputs"`
	Expected json.RawMessage       `json:"expected"`
	Actual   json.RawMessage       `json:"actual"`
	Passed   bool                  `json:"passed"`
	Fields   []RecordAssemblyField `json:"fields"`
}

type RecordAssemblyField struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	TypeID          string `json:"type_id,omitempty"`
	Presence        string `json:"presence,omitempty"`
	ExpectedPresent *bool  `json:"expected_present,omitempty"`
	ActualPresent   *bool  `json:"actual_present,omitempty"`
	Expected        string `json:"expected"`
	Actual          string `json:"actual"`
	Passed          bool   `json:"passed"`
}

// Context records the complete source input for its explicitly named model
// contract. Legacy ordinal observations preserve their original JSON shape.
type RecordOrdinalContext struct {
	Schema          string   `json:"schema"`
	Status          string   `json:"status"`
	Reason          string   `json:"reason,omitempty"`
	Text            string   `json:"text,omitempty"`
	SHA256          string   `json:"sha256,omitempty"`
	Parts           []string `json:"parts,omitempty"`
	FeatureVersion  string   `json:"feature_version,omitempty"`
	ValueFlowSHA256 string   `json:"value_flow_sha256,omitempty"`
	Scope           string   `json:"scope"`
}
