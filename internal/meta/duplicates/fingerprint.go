package duplicates

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/format"
	"go/token"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const minimumDuplicateStatements = 3

type member struct {
	group   string
	subject string
}

func fingerprintFunction(fset *token.FileSet, declaration *ast.FuncDecl, packageKey, path string) (member, bool, error) {
	if declaration.Body == nil || len(declaration.Body.List) < minimumDuplicateStatements {
		return member{}, false, nil
	}
	normalized := &ast.FuncDecl{Recv: declaration.Recv, Name: ast.NewIdent("_"), Type: declaration.Type, Body: declaration.Body}
	var output bytes.Buffer
	if err := format.Node(&output, fset, normalized); err != nil {
		return member{}, false, err
	}
	digest := sha256.Sum256(output.Bytes())
	identity := sourcepolicy.SourceSubject{Path: path, Line: fset.Position(declaration.Pos()).Line, Name: declaration.Name.Name}
	return member{group: packageKey + "#sha256:" + hex.EncodeToString(digest[:]), subject: identity.String()}, true, nil
}

func sourceDomain(path string, file *ast.File) string {
	kind := "production"
	if strings.HasSuffix(path, "_test.go") {
		kind = "test"
	}
	constraints := make([]string, 0)
	for _, group := range file.Comments {
		if group.Pos() > file.Package {
			continue
		}
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "//go:build ") || strings.HasPrefix(comment.Text, "// +build ") {
				constraints = append(constraints, comment.Text)
			}
		}
	}
	sort.Strings(constraints)
	return kind + ":" + strings.Join(constraints, ",")
}
