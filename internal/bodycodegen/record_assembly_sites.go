package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"sort"
	"strings"
)

type recordSiteCollector struct {
	body  preparedBody
	fset  *token.FileSet
	info  types.Info
	base  int
	sites []recordValueSite
}

func recordAssemblySites(body preparedBody) ([]recordValueSite, error) {
	calls, _, _, err := body.resolvePureCalls(body.body)
	if err != nil {
		return nil, err
	}
	records := body.records
	var suffix strings.Builder
	suffix.WriteString("\n}")
	if len(calls) > 0 {
		records = body.allRecords
	}
	for _, call := range calls {
		suffix.WriteString("\n" + call.declaration())
	}
	prefix := "package selection\n" + RecordDeclarations(records, false) +
		"func " + body.activity.Name + "(" + parameterDeclaration(body.parameters) + ") " + body.outputType + "{\n"
	c := recordSiteCollector{body: body, base: len(prefix), fset: token.NewFileSet(),
		info: types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}}
	file, err := parser.ParseFile(c.fset, "record-body", prefix+body.body+suffix.String(), parser.AllErrors)
	if err != nil {
		return nil, err
	}
	normalizeIntegerLocalInitializers("selection", file, c.fset)
	if _, err = new(types.Config).Check("selection", c.fset, []*ast.File{file}, &c.info); err != nil {
		return nil, fmt.Errorf("field site types: %w", err)
	}
	function, _ := findFunction(file, body.activity.Name)
	ast.Inspect(function.Body, c.visit)
	sort.Slice(c.sites, func(i, j int) bool { return c.sites[i].start < c.sites[j].start })
	return c.sites, nil
}

func (c *recordSiteCollector) visit(node ast.Node) bool {
	switch value := node.(type) {
	case *ast.CompositeLit:
		name, ok := value.Type.(*ast.Ident)
		if !ok {
			return true
		}
		for _, element := range value.Elts {
			pair := element.(*ast.KeyValueExpr)
			c.add(name.Name, pair.Key.(*ast.Ident).Name, pair.Value, "field_value")
		}
	case *ast.AssignStmt:
		if len(value.Lhs) != 1 || len(value.Rhs) != 1 {
			return true
		}
		selector, ok := value.Lhs[0].(*ast.SelectorExpr)
		if !ok {
			return true
		}
		named, ok := c.info.TypeOf(selector.X).(*types.Named)
		if ok {
			c.add(named.Obj().Name(), selector.Sel.Name, value.Rhs[0], "field_update")
		}
	}
	return true
}

func (c *recordSiteCollector) add(recordName, fieldName string, expression ast.Expr, kind string) {
	record := recordTypeByName(c.body.records, recordName)
	if record == nil {
		return
	}
	for _, field := range record.Fields {
		if field.Name != fieldName {
			continue
		}
		start, end := c.fset.Position(expression.Pos()).Offset-c.base, c.fset.Position(expression.End()).Offset-c.base
		choice := RecordValueChoice{RecordID: record.ID, FieldID: field.ID, Field: fieldName, First: c.body.body[start:end]}
		if field.Presence == "optional" {
			choice.TypeID, choice.Presence = field.TypeID, field.Presence
		}
		if kind == "field_update" {
			choice.Kind = kind
		}
		c.sites = append(c.sites, recordValueSite{start: start, end: end, kind: kind, choice: choice})
	}
}
