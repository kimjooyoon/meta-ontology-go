package extractor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go/build"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"time"
)

// One decomposition owns this session. Project packages, AST nodes and type
// evidence are never cached here. A changed import scope starts a fresh reader.
type standardImportSession struct {
	mu       sync.Mutex
	scope    string
	fallback types.Importer
	current  *moduleImporter
}

func newStandardImportSession() *standardImportSession {
	return &standardImportSession{}
}

func newModuleImporterWithStandard(root string, standard *standardImportSession) *moduleImporter {
	if standard == nil {
		standard = newStandardImportSession()
	}
	return &moduleImporter{root: root, module: readModulePath(root), standard: standard,
		packages: map[string]*types.Package{}, loading: map[string]bool{},
		exportPaths: map[string]string{}, listStarted: time.Now()}
}

func (session *standardImportSession) Import(stage *moduleImporter, path string) (*types.Package, error) {
	scope, err := standardImportScope(stage.root)
	if err != nil {
		return nil, errModuleDependency
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.fallback == nil || scope != session.scope {
		session.scope = scope
		session.fallback = importer.ForCompiler(token.NewFileSet(), runtime.Compiler, session.exportData)
	}
	session.current = stage
	defer func() { session.current = nil }()
	return session.fallback.Import(path)
}

func (session *standardImportSession) exportData(path string) (io.ReadCloser, error) {
	if session.current == nil {
		return nil, errModuleDependency
	}
	return session.current.exportData(path)
}

func standardImportScope(root string) (string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	module, err := os.ReadFile(filepath.Join(absolute, "go.mod"))
	if err != nil {
		return "", err
	}
	sum, err := os.ReadFile(filepath.Join(absolute, "go.sum"))
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	environment := moduleListEnvironment(os.Environ())
	slices.Sort(environment)
	environmentBytes, err := json.Marshal(environment)
	if err != nil {
		return "", err
	}
	identity := struct {
		Root, GOROOT, Version, Compiler, OS, Arch string
		Cgo                                       bool
		BuildTags, ToolTags, ReleaseTags          []string
		Module, Sum, Environment                  [32]byte
	}{absolute, runtime.GOROOT(), runtime.Version(), build.Default.Compiler, build.Default.GOOS,
		build.Default.GOARCH, build.Default.CgoEnabled, build.Default.BuildTags, build.Default.ToolTags,
		build.Default.ReleaseTags, sha256.Sum256(module), sha256.Sum256(sum), sha256.Sum256(environmentBytes)}
	raw, err := json.Marshal(identity)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}
