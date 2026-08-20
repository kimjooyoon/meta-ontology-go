package duplicates

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func Analyze(root string) ([]sourcepolicy.Observation, error) {
	if root == "" {
		return nil, fmt.Errorf("duplicate metric root must not be empty")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	groups := map[string][]string{}
	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		relative, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse duplicate metric source %s: %w", relative, err)
		}
		packageKey := filepath.ToSlash(filepath.Dir(relative)) + ":" + file.Name.Name + ":" + sourceDomain(relative, file)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			member, ok, err := fingerprintFunction(fset, function, packageKey, relative)
			if err != nil {
				return err
			}
			if ok {
				groups[member.group] = append(groups[member.group], member.subject)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return duplicateObservations(groups), nil
}
