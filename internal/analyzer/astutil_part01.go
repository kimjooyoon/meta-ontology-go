package analyzer

import (
	"go/ast"
	"strconv"
	"strings"
)

type importTable struct {
	aliases map[string]string
	dot     []string
}

func importsFor(file *ast.File) importTable {
	table := importTable{aliases: make(map[string]string)}
	for _, specification := range file.Imports {
		path, err := strconv.Unquote(specification.Path.Value)
		if err != nil || path == "" {
			continue
		}
		if specification.Name != nil {
			switch specification.Name.Name {
			case "_":
				continue
			case ".":
				table.dot = append(table.dot, path)
			default:
				table.aliases[specification.Name.Name] = path
			}
			continue
		}
		alias := path
		if slash := strings.LastIndex(alias, "/"); slash >= 0 {
			alias = alias[slash+1:]
		}
		table.aliases[alias] = path
	}
	return table
}

type typeReference struct {
	packagePath string
	packageName string
	name        string
}

func variableTypes(function *ast.FuncDecl, file parsedFile, imports importTable) map[string]typeReference {
	result := make(map[string]typeReference)
	collectVariableTypes(result, function.Recv, file, imports)
	collectVariableTypes(result, function.Type.Params, file, imports)
	if function.Body != nil {
		ast.Inspect(function.Body, func(node ast.Node) bool {
			declaration, ok := node.(*ast.ValueSpec)
			if ok && declaration.Type != nil {
				collectVariableType(result, declaration, file, imports)
			}
			return true
		})
	}
	return result
}
func collectVariableTypes(result map[string]typeReference, fields *ast.FieldList, file parsedFile, imports importTable) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		ref, ok := typeReferenceFor(field.Type, file, imports)
		if ok {
			for _, name := range field.Names {
				result[name.Name] = ref
			}
		}
	}
}
