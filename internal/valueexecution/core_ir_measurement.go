package valueexecution

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

type coreIRMeasurement struct {
	fingerprint                string
	programPreserved           bool
	fingerprintSensitive       bool
	unknownAttributeFailClosed bool
}

func measureCoreIR(program Program) coreIRMeasurement {
	core, err := bidir.LowerDocument(program.document)
	if err != nil {
		return coreIRMeasurement{}
	}
	measured := coreIRMeasurement{fingerprint: core.StableHash()}
	for _, node := range core.Graph.Nodes() {
		if node.Name == program.Activity {
			measured.programPreserved = node.ValueProgram == program.Text
		}
	}
	changedDocument := withActivityAttributes(program.document, program.Activity, map[string]string{
		bidir.ActivityValueProgramAttribute: "int.add:2",
	})
	changed, changedErr := bidir.LowerDocument(changedDocument)
	measured.fingerprintSensitive = changedErr == nil && changed.StableHash() != measured.fingerprint
	unknownDocument := withActivityAttributes(program.document, program.Activity, map[string]string{
		bidir.ActivityValueProgramAttribute: "int.add:1",
		"gooo:activity:unknown":             "must-fail-closed",
	})
	_, unknownErr := bidir.LowerDocument(unknownDocument)
	measured.unknownAttributeFailClosed = unknownErr != nil &&
		strings.Contains(unknownErr.Error(), "semantic IR does not support declaration attributes")
	return measured
}

func withActivityAttributes(document bidir.Document, activity string, attributes map[string]string) bidir.Document {
	document.Declarations = append([]bidir.Declaration(nil), document.Declarations...)
	for index, declaration := range document.Declarations {
		if declaration.Name != activity {
			continue
		}
		declaration.Attributes = attributes
		document.Declarations[index] = declaration
		break
	}
	return document
}
