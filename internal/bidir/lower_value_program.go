package bidir

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func bindSemanticValueProgram(declaration Declaration, node *semantic.Node) error {
	if len(declaration.Attributes) == 0 {
		return nil
	}
	program, known := declaration.Attributes[ActivityValueProgramAttribute]
	if !known {
		return fmt.Errorf("semantic IR does not support declaration attributes")
	}
	for key := range declaration.Attributes {
		if key != ActivityValueProgramAttribute && key != ActivityAssemblyAttribute {
			return fmt.Errorf("semantic IR does not support declaration attributes")
		}
	}
	if node.Kind != semantic.Activity {
		return fmt.Errorf("declaration %q: value program requires an Activity", declaration.Name)
	}
	if program == "" || program != strings.TrimSpace(program) {
		return fmt.Errorf("activity %q has a non-canonical value program", declaration.Name)
	}
	node.ValueProgram = program
	if raw, present := declaration.Attributes[ActivityAssemblyAttribute]; present {
		assembly, err := assemblyspec.DecodeCanonical(raw)
		if err != nil {
			return err
		}
		node.Assembly = assembly
	}
	return nil
}
