package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

// Field-choice provenance has its own identity; it does not impersonate an IR
// fill hole. Primitive field-location rows keep their existing 25-field shape.
type CompositionFieldChoiceSource struct {
	Schema          string                       `json:"schema"`
	ChoiceID        string                       `json:"choice_id"`
	OptionID        string                       `json:"option_id"`
	Kind            string                       `json:"kind"`
	RecordID        string                       `json:"record_id"`
	FieldID         string                       `json:"field_id"`
	SelectedBodySHA string                       `json:"selected_body_sha256"`
	SelectedDecoded bodycodegen.FillOriginRange  `json:"selected_decoded"`
	Declaration     CompositionLocalFileLocation `json:"declaration"`
	Value           CompositionLocalFileLocation `json:"value"`
}
