package bodycodegen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
)

func appendPureCallForms(root []byte, fset *token.FileSet, file *ast.File,
	calls []pureCallFunction, records []RecordType) ([]byte, error) {
	forms := []string{string(root)}
	for _, call := range calls {
		function, ok := findFunction(file, call.identity.Name)
		if !ok {
			return nil, fmt.Errorf("call closure function is missing")
		}
		body, err := canonicalizeSemanticBody(fset, function.Body, records...)
		if err != nil {
			return nil, err
		}
		signature, ok := formatNode(fset, function.Type)
		if !ok {
			return nil, fmt.Errorf("call closure signature cannot be formatted")
		}
		forms = append(forms, call.identity.ActivityID, call.identity.ProgramSHA256, signature, string(body))
	}
	return json.Marshal(forms)
}
