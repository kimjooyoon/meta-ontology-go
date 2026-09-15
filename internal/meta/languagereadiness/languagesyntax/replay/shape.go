package replay

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func astShape(file *syntax.File) (string, error) {
	if file == nil || file.Package == nil || file.Namespace == nil {
		return "", fmt.Errorf("syntax headers are absent")
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	var shape strings.Builder
	fmt.Fprintf(&shape, "package=%q\nnamespace=%q\n", file.Package.Name, file.Namespace.Name)
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			fmt.Fprintf(&shape, "entity=%q,%q,fields=%t\n", value.Name, value.ID, value.FieldsPresent)
			for _, field := range value.Fields {
				fmt.Fprintf(&shape, "field=%q,%q,%q,%q,%q\n", field.ID, field.Name,
					field.TypeRef.Spelling, field.Presence, field.Cardinality)
			}
		case *syntax.ActivityDecl:
			inputs := value.Inputs
			if inputs == nil {
				inputs = value.Parameters
			}
			fmt.Fprintf(&shape, "activity=%q", value.Name)
			for _, input := range inputs {
				fmt.Fprintf(&shape, ",input=%q", input.Name)
			}
			fmt.Fprintf(&shape, ",output=%q\n", value.Output)
		case *syntax.PolicyDecl:
			fmt.Fprintf(&shape, "policy=%q,%q\n", value.Name, value.ID)
			for _, state := range value.States {
				fmt.Fprintf(&shape, "policy-state=%q\n", state.Name)
			}
			for _, transition := range value.Transitions {
				fmt.Fprintf(&shape, "policy-transition=%q,%q\n", transition.From, transition.To)
			}
			for _, current := range value.Cases {
				fmt.Fprintf(&shape, "policy-case=%q\n", current.Name)
				for _, evidence := range current.Evidence {
					fmt.Fprintf(&shape, "policy-evidence=%q,%q\n", evidence.Name, evidence.Value)
				}
				if current.Resolution == nil {
					return "", fmt.Errorf("policy case %q has no resolution", current.Name)
				}
				resolution := current.Resolution
				fmt.Fprintf(&shape, "policy-resolution=%q,%q,%d,%q,%q,%d,%q,%q,%q,%q,%q,%q,%q,%d\n",
					resolution.Decision, resolution.Stage, resolution.Step, resolution.Reason,
					resolution.DecisionStage, resolution.DecisionStep, resolution.DecisionReason,
					resolution.UnknownClass, resolution.NextOperation, resolution.Role,
					resolution.MetaOperation, resolution.ProofChoice, resolution.Claim,
					len(resolution.BlockedBy))
				for _, blocked := range resolution.BlockedBy {
					fmt.Fprintf(&shape, "policy-blocked-by=%q\n", blocked)
				}
			}
		default:
			return "", fmt.Errorf("unsupported declaration %T", declaration)
		}
	}
	return shape.String(), nil
}
