package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func compositionReadNode(n ast.Node) (expression, base ast.Expr, kind string) {
	switch n := n.(type) {
	case *ast.IndexExpr:
		return n, n.X, "index"
	case *ast.RangeStmt:
		return n.X, n.X, "range_value"
	}
	return nil, nil, ""
}

func compositionSourceReads(prior Composition) (map[conditionOriginKey]*bodycodegen.BodyReadSource, error) {
	var origins []bodycodegen.BodyReadOrigin
	var keys []conditionOriginKey
	for _, step := range prior.Steps {
		for _, o := range step.Generation.Report.BodyReads {
			origins = append(origins, o)
			keys = append(keys, conditionOriginKey{step.Generation.Report.ActivityID, o.ActivityID, o.NativeOrdinal})
		}
	}
	result := map[conditionOriginKey]*bodycodegen.BodyReadSource{}
	if len(origins) == 0 {
		return result, nil
	}
	bound, err := bodycodegen.BindBodyReadSources([]byte(prior.GoooSource), origins)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		result[key] = bound[i]
	}
	return result, nil
}

func compositionReadSites(prior Composition) ([]CompositionReadSite, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "composition.go", prior.Source, 0)
	if err != nil {
		return nil, err
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	if _, err := new(types.Config).Check("gooo.observed.reads", fset, []*ast.File{file}, info); err != nil {
		return nil, err
	}
	origins, err := compositionSourceReads(prior)
	if err != nil {
		return nil, err
	}
	bindings, sha := arithmeticFunctionIDs(prior), digest([]byte(prior.Source))
	addresses := addressedCollectionIndexes(file)
	var sites []CompositionReadSite
	for _, declaration := range file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ids, ordinal := bindings[f.Name.Name], 0
		ast.Inspect(f.Body, func(node ast.Node) bool {
			expression, base, kind := compositionReadNode(node)
			if expression == nil {
				return true
			}
			current := ordinal
			ordinal++
			if index, ok := node.(*ast.IndexExpr); ok && addresses[index] {
				return true
			}
			typed := info.TypeOf(base)
			if typed == nil {
				return true
			}
			sequence, ok := typed.Underlying().(*types.Slice)
			if !ok {
				return true
			}
			start, end := fset.Position(expression.Pos()).Offset, fset.Position(expression.End()).Offset
			site := CompositionReadSite{RootActivityID: ids[0], ActivityID: ids[1], Kind: kind, ElementType: sequence.Elem().String(), NativeOrdinal: current,
				Expression: prior.Source[start:end], ProjectionSHA256: sha, Start: start, End: end}
			site.ExpressionID = compositionDigest(site)
			site.Source = origins[conditionOriginKey{ids[0], ids[1], current}]
			if site.Source == nil {
				site.Source = bodycodegen.UnmappedReadSource(prior.SelectedSourceSHA256, "NO_RETAINED_ORIGIN")
			}
			sites = append(sites, site)
			return true
		})
	}
	if len(sites) > 4096 {
		return nil, fmt.Errorf("collection reads exceed 4096 sites")
	}
	for _, site := range sites {
		if site.ActivityID == "" || site.RootActivityID == "" || (site.ElementType != "int64" && site.ElementType != "bool" && site.ElementType != "string") {
			return nil, fmt.Errorf("collection read lacks a supported source identity or type")
		}
	}
	return sites, nil
}

func compositionReadInputs(prior Composition) []CompositionReadInput {
	var result []CompositionReadInput
	for _, node := range prior.Plan.Activities {
		for _, slot := range node.inputSlots() {
			if slot.From >= 0 {
				continue
			}
			for _, record := range prior.Plan.Records {
				if record.Name != slot.Type {
					continue
				}
				for _, field := range record.Fields {
					if field.Cardinality != "many" {
						continue
					}
					kind, _ := compositionRecordFieldScalar(field.TypeID)
					result = append(result, CompositionReadInput{Presence: field.Presence, ActivityID: node.ID, InputKey: node.inputKey(slot.Port), Port: slot.Port, EntityID: slot.EntityID, FieldID: field.ID, FieldName: field.Name, ElementType: scalarGoType(kind)})
				}
			}
		}
	}
	return result
}

// Taking an element address does not evaluate its value. Nested index operands
// still execute, so only the addressed index itself is excluded.
func addressedCollectionIndexes(file *ast.File) map[*ast.IndexExpr]bool {
	result := map[*ast.IndexExpr]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		unary, ok := node.(*ast.UnaryExpr)
		if !ok || unary.Op != token.AND {
			return true
		}
		expression := unary.X
		for {
			paren, ok := expression.(*ast.ParenExpr)
			if !ok {
				break
			}
			expression = paren.X
		}
		if index, ok := expression.(*ast.IndexExpr); ok {
			result[index] = true
		}
		return true
	})
	return result
}
