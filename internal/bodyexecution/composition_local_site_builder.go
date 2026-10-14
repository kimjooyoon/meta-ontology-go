package bodyexecution

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"go/ast"
	"go/token"
	"go/types"
)

type localSiteBuilder struct {
	source, sha string
	fset        *token.FileSet
	info        *types.Info
	origins     map[conditionOriginKey]compositionLocalOrigin
	matched     map[conditionOriginKey]bool
	objects     map[types.Object]conditionOriginKey
	variables   map[conditionOriginKey]types.Object
	sites       []CompositionLocalSite
	mismatch    error
}

func (b *localSiteBuilder) visit(node ast.Node, ids [2]string, ordinal *int) bool {

	name, expression, kind, operation := bodycodegen.LocalWriteExpression(node)
	if name == nil {
		return true
	}
	key := conditionOriginKey{ids[0], ids[1], *ordinal}
	origin, found := b.origins[key]
	current := *ordinal
	*ordinal++
	if !found {
		return true
	}
	b.matched[key] = true
	if origin.definition.Name != name.Name || origin.definition.Kind != kind || origin.definition.Operation != operation {
		b.mismatch = fmt.Errorf("local write origin differs from emitted projection")
		return false
	}
	object := b.info.ObjectOf(name)
	if object == nil {
		b.mismatch = fmt.Errorf("local write has no native lexical binding")
		return false
	}
	variable := conditionOriginKey{ids[0], ids[1], origin.definition.VariableOrdinal}
	if previous, found := b.objects[object]; (found && previous != variable) ||
		(b.variables[variable] != nil && b.variables[variable] != object) || origin.definition.VariableOrdinal < 0 {
		b.mismatch = fmt.Errorf("local variable identity differs from native lexical binding")
		return false
	}
	b.objects[object], b.variables[variable] = variable, object
	scalar := object.Type().String()
	if scalar != "int64" && scalar != "bool" && scalar != "string" {
		return true
	}
	start, end := b.fset.Position(expression.Pos()).Offset, b.fset.Position(expression.End()).Offset
	site := CompositionLocalSite{RootActivityID: ids[0], ActivityID: ids[1], NativeOrdinal: current,
		VariableOrdinal: origin.definition.VariableOrdinal, Name: name.Name, Kind: kind, Operation: operation,
		ScalarType: scalar, Expression: b.source[start:end], ProjectionSHA256: b.sha, Start: start, End: end,
		Source: origin.source}
	site.ExpressionID = compositionDigest(site)
	b.sites = append(b.sites, site)
	return true
}
