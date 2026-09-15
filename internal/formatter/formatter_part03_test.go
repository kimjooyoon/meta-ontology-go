package formatter

import (
	"strconv"
	"strings"
)

func parseFixture(source string) (*fixtureAST, Diagnostics) {
	ast := &fixtureAST{}
	var diagnostics Diagnostics
	for lineNumber, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "package" {
			ast.packageName = fields[1]
			continue
		}
		if len(fields) == 2 && fields[0] == "namespace" {
			ast.namespace = fields[1]
			continue
		}
		if strings.HasPrefix(line, "entity ") {
			declaration, ok := parseFixtureEntity(fields)
			if !ok {
				diagnostics = appendFixtureError(diagnostics, lineNumber)
				continue
			}
			ast.declarations = append(ast.declarations, declaration)
			continue
		}
		if strings.HasPrefix(line, "activity ") {
			declaration, ok := parseFixtureActivity(line)
			if !ok {
				diagnostics = appendFixtureError(diagnostics, lineNumber)
				continue
			}
			ast.declarations = append(ast.declarations, declaration)
			continue
		}
		diagnostics = appendFixtureError(diagnostics, lineNumber)
	}
	return ast, diagnostics
}
func parseFixtureEntity(fields []string) (fixtureDeclaration, bool) {
	if len(fields) != 4 || fields[0] != "entity" || fields[2] != "id" {
		return fixtureDeclaration{}, false
	}
	decoded, err := strconv.Unquote(fields[3])
	if err != nil {
		return fixtureDeclaration{}, false
	}
	return fixtureDeclaration{kind: EntityDeclaration, name: fields[1], id: decoded}, decoded != ""
}
