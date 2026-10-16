package policyprojection

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/recordbridge"
)

func renderAdapter(report bodycodegen.Report, nativeTarget []byte) (string, error) {
	if len(report.RecordTypes) != 1 || report.RecordTypes[0].ID != RecordID ||
		len(report.InputParameters) != 2 || report.InputParameters[0].Type != "bool" ||
		report.InputParameters[1].Type != "bool" {
		return "", fmt.Errorf("policy record or input identity differs")
	}
	record := report.RecordTypes[0]
	fields, err := recordbridge.Fields("disposition.go", nativeTarget, "SourceDisposition")
	if err != nil {
		return "", err
	}
	if len(fields) != len(record.Fields) || len(fields) != 6 {
		return "", fmt.Errorf("policy native record field population differs")
	}
	var assignments []string
	seen := map[string]bool{}
	for _, field := range record.Fields {
		target, found := fields[field.ID]
		if !found || field.ID == "" || field.GoName == "" || seen[field.ID] ||
			field.Presence != "" || field.Cardinality != "" {
			return "", fmt.Errorf("policy field %q has no unique mandatory native target", field.ID)
		}
		if !(target.Kind == "string" && field.TypeID == "urn:gooo:type:string" ||
			target.Kind == "bool" && field.TypeID == "urn:gooo:type:boolean") {
			return "", fmt.Errorf("policy field %q changes native scalar type", field.ID)
		}
		seen[field.ID] = true
		assignments = append(assignments, target.Name+": row."+field.GoName)
	}
	return fmt.Sprintf("func sourceOperationDisposition(registered, originRecorded bool) SourceDisposition {\n"+
		"row := %s(registered, originRecorded)\nreturn SourceDisposition{%s}\n}\n\n",
		report.Activity, strings.Join(assignments, ",\n")), nil
}
