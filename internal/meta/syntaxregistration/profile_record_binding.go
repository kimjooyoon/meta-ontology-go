package syntaxregistration

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func bindProfileObservationLayout(ir semantic.IR) error {
	rows, found := ir.Graph.NodeByName(ir.Namespace, "ProfileCompatibility")
	if !found || rows.Kind != semantic.Entity || rows.ID != "gooo://syntax-registration/profile-compatibility" {
		return fmt.Errorf("registration profile observation row identity mismatch")
	}
	expected := map[semantic.ID]semantic.ID{
		"gooo://syntax-registration/profile-compatibility/profile":  semantic.BuiltinStringTypeID,
		"gooo://syntax-registration/profile-compatibility/passed":   semantic.BuiltinBooleanTypeID,
		"gooo://syntax-registration/profile-compatibility/eligible": semantic.BuiltinBooleanTypeID,
	}
	if len(rows.Fields) != len(expected) {
		return fmt.Errorf("registration profile observation row fields mismatch")
	}
	for _, field := range rows.Fields {
		if expected[field.ID] == "" || field.TypeRef.ID != expected[field.ID] ||
			string(field.Presence) != "required" || string(field.Cardinality) != "one" {
			return fmt.Errorf("registration profile observation field ABI mismatch: %s", field.ID)
		}
	}
	input, found := ir.Graph.NodeByName(ir.Namespace, "ProfileObservation")
	if !found || len(input.Fields) != 6 {
		return fmt.Errorf("registration profile observation input shape mismatch")
	}
	for _, field := range input.Fields {
		if field.ID == "gooo://syntax-registration/profile-observation/profiles" &&
			field.TypeRef.ID == rows.ID && string(field.Presence) == "required" && string(field.Cardinality) == "many" {
			return nil
		}
	}
	return fmt.Errorf("registration profile observation list ABI mismatch")
}
