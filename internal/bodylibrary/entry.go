package bodylibrary

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func libraryEntry(plan bodyexecution.CompositionPlan, source string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "projection.go", source, 0)
	if err != nil {
		return "", err
	}
	pkg, err := new(types.Config).Check("gooo.library.entry", fset, []*ast.File{file}, nil)
	if err != nil {
		return "", err
	}
	var parameters, body, origins strings.Builder
	root, selected, outputType := 0, -1, ""
	for i, node := range plan.Activities {
		function, ok := pkg.Scope().Lookup(node.GoFunction).(*types.Func)
		if !ok {
			return "", fmt.Errorf("library activity %q has no native function", node.ID)
		}
		signature := function.Type().(*types.Signature)
		inputs := node.Inputs
		if len(inputs) == 0 {
			inputs = []bodyexecution.CompositionInput{{Port: "input", EntityID: node.InputEntityID, From: node.InputFrom}}
		}
		if signature.Params().Len() != len(inputs) || signature.Results().Len() != 1 {
			return "", fmt.Errorf("library activity %q has a different native signature", node.ID)
		}
		if node.Name == plan.EntryActivity {
			selected, outputType = i, localType(signature.Results().At(0).Type())
		}
		arguments := make([]string, 0, len(inputs))
		for port, slot := range inputs {
			argument := fmt.Sprintf("value%d", slot.From)
			if slot.From < 0 {
				argument = fmt.Sprintf("input%d", root)
				fmt.Fprintf(&parameters, "%s %s,", argument, localType(signature.Params().At(port).Type()))
				fmt.Fprintf(&origins, "// %s receives activity=%q port=%q entity=%q\n", argument, node.ID, slot.Port, slot.EntityID)
				root++
			}
			arguments = append(arguments, argument)
		}
		fmt.Fprintf(&body, "value%d := %s(%s)\n", i, node.GoFunction, strings.Join(arguments, ","))
	}
	if selected < 0 {
		return "", fmt.Errorf("library entry is absent from the typed plan")
	}
	node := plan.Activities[selected]
	fmt.Fprintf(&origins, "// Evaluate returns activity=%q entity=%q\n", node.ID, node.OutputEntityID)
	return fmt.Sprintf("\n//gooo:generated:start id=%q kind=\"library-entry\"\n%s"+
		"func Evaluate(%s) %s {\n%sreturn value%d\n}\n"+
		"//gooo:generated:end id=%q kind=\"library-entry\"\n",
		node.ID, origins.String(), parameters.String(), outputType, body.String(), selected, node.ID), nil
}

func localType(value types.Type) string {
	return types.TypeString(value, func(*types.Package) string { return "" })
}
