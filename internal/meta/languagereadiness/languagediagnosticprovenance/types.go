package languagediagnosticprovenance

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"sort"
)

func typeObservation(fixture string) (Observation, bool) {
	source, code := typeFixture(fixture)
	if source == "" {
		return Observation{}, false
	}
	files := token.NewFileSet()
	file, parseError := parser.ParseFile(files, "generated.go", source, parser.AllErrors)
	if parseError != nil {
		return Observation{}, false
	}
	errors := []types.Error{}
	config := types.Config{GoVersion: "go1.27", Error: func(failure error) {
		switch typed := failure.(type) {
		case types.Error:
			errors = append(errors, typed)
		case *types.Error:
			errors = append(errors, *typed)
		}
	}}
	_, _ = config.Check("fixture", files, []*ast.File{file}, nil)
	if len(errors) == 0 {
		return Observation{}, false
	}
	sort.Slice(errors, func(left, right int) bool {
		leftPosition := files.PositionFor(errors[left].Pos, false)
		rightPosition := files.PositionFor(errors[right].Pos, false)
		if leftPosition.Offset != rightPosition.Offset {
			return leftPosition.Offset < rightPosition.Offset
		}
		return errors[left].Msg < errors[right].Msg
	})
	first := errors[0]
	hardness := "HARD"
	if first.Soft {
		hardness = "SOFT"
	}
	physical := tokenPosition(files.PositionFor(first.Pos, false))
	logical := tokenPosition(files.PositionFor(first.Pos, true))
	return Observation{
		Origin: "GO", Stage: "TYPE", Code: code, Message: first.Msg,
		Hardness: hardness, Physical: oneByteSpan(physical),
		Logical: oneByteSpan(logical),
	}, true
}

func typeFixture(fixture string) (string, string) {
	switch fixture {
	case "undefined-identifier":
		return "package p\nvar _ = missing\n", "GO_TYPES_UNDEFINED"
	case "assignment-mismatch":
		return "package p\nvar _ int = \"x\"\n", "GO_TYPES_ASSIGNMENT"
	case "constraint-violation":
		return "package p\nfunc F[T ~int](v T) T { return v }\nvar _ = F(\"x\")\n",
			"GO_TYPES_CONSTRAINT"
	default:
		return "", ""
	}
}
