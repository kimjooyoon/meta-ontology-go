package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
)

func validateBodyLocal(value *ast.DeclStmt, readonly, locals map[string]bool, records []RecordType) (string, error) {
	declaration, ok := value.Decl.(*ast.GenDecl)
	if !ok || declaration.Tok != token.VAR || len(declaration.Specs) != 1 {
		return "", fmt.Errorf("only one local let declaration is supported")
	}
	spec, ok := declaration.Specs[0].(*ast.ValueSpec)
	if !ok || len(spec.Names) != 1 {
		return "", fmt.Errorf("let requires one local name")
	}
	name := spec.Names[0].Name
	if spec.Type != nil {
		typeName, supported := spec.Type.(*ast.Ident)
		if name != "_goooResult" || !supported || !supportedBodyType(typeName.Name, records) || len(spec.Values) != 0 {
			return "", fmt.Errorf("explicit local types are reserved for compiler-generated result joins")
		}
	} else if len(spec.Values) != 1 {
		return "", fmt.Errorf("let requires one inferred local value")
	}
	if readonly[name] || locals[name] {
		return "", fmt.Errorf("let name %q is already bound", name)
	}
	if len(spec.Values) == 1 {
		if err := validateExpression(spec.Values[0]); err != nil {
			return "", err
		}
	}
	return name, nil
}
