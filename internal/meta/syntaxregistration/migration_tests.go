package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func generateMigrationTests(source *goSource, version, capability int) error {
	for _, declaration := range source.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		current := function.Name.Name == "TestCurrentDenominatorPinsExistingCapabilities" ||
			function.Name.Name == "TestCurrentDenominatorRejectsLoweredTarget"
		if !current && function.Name.Name != "TestRecordMigrationPreservesPreviousBoundaryEvidence" &&
			!strings.HasPrefix(function.Name.Name, "TestGeneratedRegistrationMigrationV") {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if !current {
				if call, ok := node.(*ast.CallExpr); ok && source.text(call.Fun) == "activeDenominator" && len(call.Args) == 0 {
					source.replace(call, evidenceName(version-1))
					return false
				}
				return true
			}
			if identifier, ok := node.(*ast.Ident); ok && identifier.Name == digestName(version-1) {
				source.replace(identifier, digestName(version))
			}
			comparison, ok := node.(*ast.BinaryExpr)
			if ok && (comparison.Op == token.NEQ || comparison.Op == token.EQL) {
				value, literal := integer(comparison.Y)
				name := source.text(comparison.X)
				if literal && name == "observed.Version" && value == version-1 {
					source.replace(comparison.Y, strconv.Itoa(version))
				}
				if literal && name == "observed.Boundaries[0].Target" && value == capability {
					source.replace(comparison.Y, strconv.Itoa(capability+1))
				}
			}
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(literal.Value)
			if err != nil {
				return true
			}
			next := text
			switch text {
			case fmt.Sprintf("\"target\": %d", capability):
				next = fmt.Sprintf("\"target\": %d", capability+1)
			case fmt.Sprintf("\"target\": %d", capability-1):
				next = fmt.Sprintf("\"target\": %d", capability)
			default:
				next = strings.ReplaceAll(text, fmt.Sprintf("pinned v%d migration", version-1), fmt.Sprintf("pinned v%d migration", version))
			}
			if text != next {
				source.replace(literal, strconv.Quote(next))
			}
			return true
		})
	}
	for _, name := range []string{"TestCurrentDenominatorPinsExistingCapabilities", "TestCurrentDenominatorRejectsLoweredTarget"} {
		if _, err := source.function(name); err != nil {
			return err
		}
	}
	test := fmt.Sprintf("\nfunc TestGeneratedRegistrationMigrationV%dPreservesBaseline(t *testing.T) {\n"+
		"if digestBytes(%s) != %s { t.Fatal(\"baseline evidence was rewritten\") }\n"+
		"var previous, current denominator\n"+
		"if err := json.Unmarshal(%s, &previous); err != nil { t.Fatal(err) }\n"+
		"if err := json.Unmarshal(%s, &current); err != nil { t.Fatal(err) }\n"+
		"if previous.Version != %d || current.Version != %d || len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 { t.Fatal(\"migration inventory changed\") }\n"+
		"for index, expected := range previous.Boundaries {\n"+
		"if index == 0 { expected.Target++ }\n"+
		"if current.Boundaries[index] != expected { t.Fatalf(\"unrelated boundary changed: %%d\", index) }\n"+
		"}\n}\n", version, evidenceName(version-1), digestName(version-1),
		evidenceName(version-1), evidenceName(version), version-1, version)
	anchor, err := source.function("TestCurrentDenominatorPinsExistingCapabilities")
	if err != nil {
		return err
	}
	if err := source.requireImport(anchor, "encoding/json", "json"); err != nil {
		return err
	}
	if err := source.requireImport(anchor, "testing", "testing"); err != nil {
		return err
	}
	source.appendAt(anchor, test)
	return nil
}
