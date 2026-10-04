package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func emitCompositionProjection(steps []CompositionStep, graph compositionGraph) (string, error) {
	var source strings.Builder
	source.WriteString("package main\n")
	source.WriteString(bodycodegen.RecordDeclarations(graph.plan.Records, true))
	for i, step := range steps {
		part, err := compositionStepFunction(step.Generation, graph.nodes[i], i)
		if err != nil {
			return "", err
		}
		source.WriteString(part)
	}
	if source.Len() > 256<<10 {
		return "", fmt.Errorf("composition projection exceeds 256 KiB")
	}
	raw, err := format.Source([]byte(source.String()))
	if err != nil {
		return "", err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "composition.go", raw, 0)
	if err != nil {
		return "", err
	}
	if _, err := new(types.Config).Check("gooo.observed.composition", fset, []*ast.File{file}, nil); err != nil {
		return "", fmt.Errorf("composition projection typecheck: %w", err)
	}
	return string(raw), nil
}

func compositionFunction(source, activity string, index int) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "activity.go", source, parser.ParseComments)
	if err != nil || len(file.Decls) != 1 || len(file.Imports) != 0 {
		return "", fmt.Errorf("composition requires one pure function per body projection")
	}
	function, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok || function.Name.Name != activity || function.Recv != nil {
		return "", fmt.Errorf("composition function differs from declared activity")
	}
	start, end := fset.Position(function.Name.Pos()).Offset, fset.Position(function.Name.End()).Offset
	body := source[:start] + fmt.Sprintf("GoooComposedActivity%d", index) + source[end:]
	return "\n" + body[fset.Position(file.Name.End()).Offset:], nil
}

func compositionDriver(graph compositionGraph) (string, error) {
	var source strings.Builder
	source.WriteString("package main\nimport(\"encoding/json\";\"os\")\nfunc main(){\n")
	roots := 0
	for _, node := range graph.nodes[:graph.count] {
		for _, input := range node.inputSlots() {
			if input.From < 0 {
				roots++
			}
		}
	}
	fmt.Fprintf(&source, "var in [][%d]json.RawMessage; if json.NewDecoder(os.Stdin).Decode(&in)!=nil||len(in)>128{os.Exit(2)}\n", roots)
	fmt.Fprintf(&source, "out:=make([][%d]json.RawMessage,len(in));for c,row:=range in{\n", graph.count)
	root := 0
	for i, node := range graph.nodes[:graph.count] {
		arguments := make([]string, 0, len(node.inputSlots()))
		for _, slot := range node.inputSlots() {
			input := fmt.Sprintf("v%d", slot.From)
			if slot.From < 0 {
				input = fmt.Sprintf("input%d", root)
				fmt.Fprintf(&source, "var %s %s;if json.Unmarshal(row[%d],&%s)!=nil{os.Exit(2)}\n", input, graph.valueGoType(slot.Type), root, input)
				root++
			}
			arguments = append(arguments, input)
		}
		fmt.Fprintf(&source, "v%d:=GoooComposedActivity%d(%s);out[c][%d],_=json.Marshal(v%d)\n", i, i, strings.Join(arguments, ","), i, i)
	}
	source.WriteString("};if json.NewEncoder(os.Stdout).Encode(out)!=nil{os.Exit(3)}}\n")
	raw, err := format.Source([]byte(source.String()))
	return string(raw), err
}
