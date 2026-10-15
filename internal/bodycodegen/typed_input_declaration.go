package bodycodegen

import "github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"

// TypedInputDeclaration locates a type reference in an authored activity
// signature. The port is positional within this source; it is not a role ID.
type TypedInputDeclaration struct {
	Status       string `json:"status"`
	Reason       string `json:"reason,omitempty"`
	View         string `json:"view"`
	SourceSHA256 string `json:"source_sha256"`
	ActivityID   string `json:"activity_id"`
	InputIndex   int    `json:"input_index"`
	Port         string `json:"port"`
	Name         string `json:"name,omitempty"`
	NameStart    int    `json:"name_start_byte,omitempty"`
	NameEnd      int    `json:"name_end_byte,omitempty"`
	TypeName     string `json:"type_name"`
	TypeID       string `json:"type_id,omitempty"`
	Start        int    `json:"start_byte"`
	End          int    `json:"end_byte"`
	Expression   string `json:"expression,omitempty"`
}

func inputDeclarationSource(p preparedBody, source []byte, index int) *TypedInputDeclaration {
	r := &TypedInputDeclaration{Status: "UNMAPPED", Reason: "INPUT_DECLARATION_UNAVAILABLE",
		View: "activity_signature", SourceSHA256: digest(source), ActivityID: p.activityID, InputIndex: index}
	if index < 0 || index >= len(p.activity.Inputs) || index >= len(p.parameters) {
		return r
	}
	ref := p.activity.Inputs[index]
	r.Port, r.TypeName, r.TypeID = inputbinding.PositionName(len(p.parameters), index), ref.Name, p.entityIDs[ref.Name]
	if len(p.activity.InputNames) > 0 && p.activity.InputNames[index].Name != "" {
		name := p.activity.InputNames[index]
		start, end := name.Span.Start.Offset, name.Span.End.Offset
		if start < 0 || end <= start || end > len(source) || string(source[start:end]) != name.Name {
			return r
		}
		r.Name, r.NameStart, r.NameEnd = name.Name, start, end
	}
	start, end := ref.Span.Start.Offset, ref.Span.End.Offset
	if start < 0 || end <= start || end > len(source) || string(source[start:end]) != ref.Name {
		return r
	}
	r.Status, r.Reason, r.Start, r.End, r.Expression = "MAPPED", "", start, end, string(source[start:end])
	return r
}
