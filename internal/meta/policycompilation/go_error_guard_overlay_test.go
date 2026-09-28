package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

type goGuardPackageFiles struct {
	Dir, Name                                    string
	GoFiles, CgoFiles, TestGoFiles, XTestGoFiles []string
}

type goGuardNativeView struct {
	directory, oracleDigest string
	production              []string
	files, runtime          map[string][]byte
	backing                 map[string]string
	active                  goGuardPackageFiles
}

func goGuardActivePackage(t *testing.T, root string) goGuardPackageFiles {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "list", "-json", "-mod=readonly", "./cmd/gooo")
	command.Dir = root
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("observe active Go package: %v stderr=%s", err, stderr.String())
	}
	var active goGuardPackageFiles
	if err := json.Unmarshal(stdout.Bytes(), &active); err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(active.Dir) != filepath.Join(root, "cmd", "gooo") || active.Name != "main" ||
		len(active.GoFiles)+len(active.CgoFiles) == 0 || len(active.TestGoFiles)+len(active.XTestGoFiles) == 0 {
		t.Fatal("active package or test source set is not the declared CLI package")
	}
	return active
}

func freezeGoGuardNativeView(t *testing.T, root, temp string) goGuardNativeView {
	t.Helper()
	active := goGuardActivePackage(t, root)
	view := goGuardNativeView{
		directory: active.Dir, active: active,
		production: append(append([]string{}, active.GoFiles...), active.CgoFiles...),
		files:      map[string][]byte{}, runtime: map[string][]byte{}, backing: map[string]string{},
	}
	snapshot := filepath.Join(temp, "snapshot")
	if err := os.Mkdir(snapshot, 0700); err != nil {
		t.Fatal(err)
	}
	names := append(append(append([]string{}, view.production...), active.TestGoFiles...), active.XTestGoFiles...)
	sourcePins, oraclePins := map[string]string{}, map[string]string{}
	for _, name := range names {
		if name == "" || filepath.Base(name) != name || view.files[name] != nil {
			t.Fatal("active package has an unsafe or duplicate file name")
		}
		raw := readGoGuardNativeFile(t, filepath.Join(active.Dir, name))
		view.files[name], sourcePins[name] = raw, DigestBytes(raw)
		backing := filepath.Join(snapshot, name)
		writeGoGuardNativeFile(t, backing, raw)
		view.backing[filepath.Join(active.Dir, name)] = backing
	}
	for _, name := range append(append([]string{}, active.TestGoFiles...), active.XTestGoFiles...) {
		oraclePins[name] = sourcePins[name]
	}
	oracle, err := json.Marshal(oraclePins)
	if err != nil {
		t.Fatal(err)
	}
	view.oracleDigest = DigestBytes(oracle)
	fixtures, err := filepath.Glob(filepath.Join(root, "examples", "billing-package", "*.gooo"))
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("runtime Gooo fixture input is unavailable: %v", err)
	}
	runtimePins := map[string]string{}
	for _, path := range fixtures {
		raw := readGoGuardNativeFile(t, path)
		view.runtime[path] = raw
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		runtimePins[filepath.ToSlash(relative)] = DigestBytes(raw)
	}
	manifest, err := json.Marshal(map[string]map[string]string{"package_go": sourcePins, "oracle": oraclePins, "runtime_gooo": runtimePins})
	if err != nil {
		t.Fatal(err)
	}
	writeGoGuardNativeFile(t, filepath.Join(temp, "input-snapshot.json"), manifest)
	t.Logf("guard input snapshot: digest=%s manifest=%s", DigestBytes(manifest), manifest)
	return view
}

func (view goGuardNativeView) requireUnchanged(t *testing.T, root string) {
	t.Helper()
	before, err := json.Marshal(view.active)
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(goGuardActivePackage(t, root))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("native trials changed active package membership")
	}
	for name, raw := range view.files {
		if !bytes.Equal(raw, readGoGuardNativeFile(t, filepath.Join(view.directory, name))) {
			t.Fatalf("native trials changed package input %s", name)
		}
	}
	fixtures, err := filepath.Glob(filepath.Join(root, "examples", "billing-package", "*.gooo"))
	if err != nil || len(fixtures) != len(view.runtime) {
		t.Fatal("native trials changed runtime Gooo input membership")
	}
	for _, path := range fixtures {
		if raw, exists := view.runtime[path]; !exists || !bytes.Equal(raw, readGoGuardNativeFile(t, path)) {
			t.Fatalf("native trials changed runtime Gooo input %s", path)
		}
	}
}

func readGoGuardNativeFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

type goGuardDeclaration struct {
	file     *ast.File
	function *ast.FuncDecl
	set      *token.FileSet
}

func findGoGuardDeclaration(raw []byte, name string) (goGuardDeclaration, error) {
	declaration := goGuardDeclaration{set: token.NewFileSet()}
	file, err := parser.ParseFile(declaration.set, "subject.go", raw, parser.ParseComments)
	if err != nil {
		return declaration, err
	}
	declaration.file = file
	for _, node := range file.Decls {
		function, ok := node.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name.Name != name {
			continue
		}
		if declaration.function != nil {
			return declaration, fmt.Errorf("duplicate function declaration %s", name)
		}
		declaration.function = function
	}
	return declaration, nil
}

func bindGoGuardDeclaration(files map[string][]byte, production []string, name string, variant []byte) (string, []byte, error) {
	replacement, err := findGoGuardDeclaration(variant, name)
	if err != nil || replacement.function == nil {
		return "", nil, fmt.Errorf("variant declaration %s is unavailable: %v", name, err)
	}
	var owner string
	var current goGuardDeclaration
	for _, path := range production {
		declaration, err := findGoGuardDeclaration(files[path], name)
		if err != nil {
			return "", nil, fmt.Errorf("current file %s: %w", path, err)
		}
		if declaration.function == nil {
			continue
		}
		if owner != "" {
			return "", nil, fmt.Errorf("ambiguous current declaration %s", name)
		}
		owner, current = path, declaration
	}
	if owner == "" {
		return "", nil, fmt.Errorf("current declaration %s is missing", name)
	}
	if err := matchGoGuardDeclarationContext(current, replacement); err != nil {
		return "", nil, err
	}
	start, end := current.set.Position(current.function.Pos()).Offset, current.set.Position(current.function.End()).Offset
	first, last := replacement.set.Position(replacement.function.Pos()).Offset, replacement.set.Position(replacement.function.End()).Offset
	rebound := append([]byte{}, files[owner][:start]...)
	rebound = append(rebound, variant[first:last]...)
	rebound = append(rebound, files[owner][end:]...)
	return owner, rebound, nil
}

func matchGoGuardDeclarationContext(current, replacement goGuardDeclaration) error {
	if current.file.Name.Name != replacement.file.Name.Name {
		return fmt.Errorf("native declaration package differs")
	}
	var before, after bytes.Buffer
	if err := format.Node(&before, current.set, current.function.Type); err != nil {
		return err
	}
	if err := format.Node(&after, replacement.set, replacement.function.Type); err != nil {
		return err
	}
	if !bytes.Equal(before.Bytes(), after.Bytes()) {
		return fmt.Errorf("native declaration signature differs")
	}
	currentImports, variantImports := goGuardDeclarationImports(current.file), goGuardDeclarationImports(replacement.file)
	var mismatch error
	ast.Inspect(replacement.function, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		identifier, ok := selector.X.(*ast.Ident)
		if !ok || identifier.Obj != nil && identifier.Obj.Kind != ast.Pkg {
			return true
		}
		if path, imported := variantImports[identifier.Name]; imported && currentImports[identifier.Name] != path {
			mismatch = fmt.Errorf("native import binding %s differs", identifier.Name)
		}
		return true
	})
	return mismatch
}

func goGuardDeclarationImports(file *ast.File) map[string]string {
	imports := map[string]string{}
	for _, imported := range file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			continue
		}
		name := filepath.Base(path)
		if imported.Name != nil {
			name = imported.Name.Name
		}
		imports[name] = path
	}
	return imports
}
