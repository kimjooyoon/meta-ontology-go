package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	moduleListTimeout      = 30 * time.Second
	moduleListTotalTimeout = 120 * time.Second
	moduleListMax          = 8 << 20
	moduleListTotalMax     = 16 << 20
	moduleListCallsMax     = 64
	moduleListPackagesMax  = 512
	moduleListWaitDelay    = 2 * time.Second
	moduleSourceMax        = 8 << 20
	moduleFileMax          = 1 << 20
	moduleExportMax        = 64 << 20
	moduleFilesMax         = 256
	modulePackageMax       = 64
)

var errModuleDependency = errors.New("module dependency could not be resolved safely")

type moduleDependencyCauseCode string

const (
	moduleCauseCallBudget   moduleDependencyCauseCode = "GO_LIST_CALL_BUDGET"
	moduleCauseTimeBudget   moduleDependencyCauseCode = "GO_LIST_TIME_BUDGET"
	moduleCauseByteBudget   moduleDependencyCauseCode = "GO_LIST_BYTE_BUDGET"
	moduleCauseTimeout      moduleDependencyCauseCode = "GO_LIST_TIMEOUT"
	moduleCauseOutputCap    moduleDependencyCauseCode = "GO_LIST_OUTPUT_CAP"
	moduleCauseSubprocess   moduleDependencyCauseCode = "GO_LIST_SUBPROCESS_FAILURE"
	moduleCausePackageLimit moduleDependencyCauseCode = "MODULE_PACKAGE_LIMIT"
	moduleCauseSourceLimit  moduleDependencyCauseCode = "MODULE_SOURCE_LIMIT"
)

type moduleDependencyFailure struct {
	code moduleDependencyCauseCode
}

func (failure *moduleDependencyFailure) Error() string {
	return "module dependency could not be resolved safely (" + string(failure.code) + ")"
}

func (failure *moduleDependencyFailure) Unwrap() error { return errModuleDependency }

func moduleFailure(code moduleDependencyCauseCode) error {
	return &moduleDependencyFailure{code: code}
}

func preserveModuleFailure(err error) error {
	if failure, ok := errors.AsType[*moduleDependencyFailure](err); ok {
		return failure
	}
	return errModuleDependency
}

type moduleImporter struct {
	root                string
	module              string
	fallback            types.Importer
	packages            map[string]*types.Package
	loading             map[string]bool
	exportPaths         map[string]string
	listStarted         time.Time
	listCalls           int
	listBytes           int64
	externalPackages    int
	externalSourceBytes int64
	exportPackages      int
}

func newModuleImporter(root string) *moduleImporter {
	imports := &moduleImporter{root: root, module: readModulePath(root),
		packages: map[string]*types.Package{}, loading: map[string]bool{},
		exportPaths: map[string]string{}, listStarted: time.Now()}
	imports.fallback = importer.ForCompiler(token.NewFileSet(), runtime.Compiler, imports.exportData)
	return imports
}

func (imports *moduleImporter) Import(path string) (*types.Package, error) {
	if !validModuleImportPath(path) {
		return nil, errModuleDependency
	}
	if package_, ok := imports.packages[path]; ok {
		return package_, nil
	}
	var files []string
	var err error
	if !imports.local(path) {
		if imports.standardLibrary(path) {
			package_, importErr := imports.fallback.Import(path)
			if importErr != nil || package_ == nil {
				return nil, preserveModuleFailure(importErr)
			}
			imports.packages[path] = package_
			return package_, nil
		}
		if !imports.allowExternalPackage() {
			return nil, moduleFailure(moduleCausePackageLimit)
		}
		files, err = imports.moduleFiles(path)
		if err != nil {
			return nil, err
		}
	} else {
		files, err = imports.files(path)
	}
	if err != nil {
		return nil, err
	}
	if imports.loading[path] {
		return nil, os.ErrInvalid
	}
	imports.loading[path] = true
	defer delete(imports.loading, path)
	fset := token.NewFileSet()
	parsed := make([]*ast.File, 0, len(files))
	for _, name := range files {
		var data []byte
		var readErr error
		if imports.local(path) {
			data, readErr = os.ReadFile(name)
		} else {
			data, readErr = readBoundedModuleSource(name)
			if int64(len(data)) > moduleSourceMax-imports.externalSourceBytes {
				return nil, moduleFailure(moduleCauseSourceLimit)
			}
			imports.externalSourceBytes += int64(len(data))
		}
		if readErr != nil {
			if !imports.local(path) {
				return nil, preserveModuleFailure(readErr)
			}
			return nil, readErr
		}
		file, parseErr := parser.ParseFile(fset, name, data, parser.ParseComments)
		if parseErr != nil {
			if !imports.local(path) {
				return nil, errModuleDependency
			}
			return nil, parseErr
		}
		parsed = append(parsed, file)
	}
	checked, checkErr := (&types.Config{Importer: imports, Error: func(error) {}}).Check(path, fset, parsed, nil)
	if checkErr != nil {
		if !imports.local(path) {
			return nil, errModuleDependency
		}
		return nil, checkErr
	}
	if checked == nil {
		return nil, os.ErrInvalid
	}
	imports.packages[path] = checked
	return checked, nil
}

func (imports *moduleImporter) allowExternalPackage() bool {
	if imports.externalPackages >= modulePackageMax {
		return false
	}
	imports.externalPackages++
	return true
}

func (imports *moduleImporter) standardLibrary(path string) bool {
	package_, err := build.Default.Import(path, imports.root, build.FindOnly)
	return err == nil && package_.Goroot
}

func (imports *moduleImporter) moduleFiles(path string) ([]string, error) {
	if !validModuleImportPath(path) || imports.root == "" {
		return nil, errModuleDependency
	}
	listed, err := imports.listModulePackage(path, false)
	if err != nil {
		return nil, preserveModuleFailure(err)
	}
	if len(listed.GoFiles) > moduleFilesMax {
		return nil, moduleFailure(moduleCausePackageLimit)
	}
	if listed.Goroot || listed.ImportPath != path || listed.Incomplete || len(listed.CgoFiles) != 0 ||
		len(listed.GoFiles) == 0 || !filepath.IsAbs(listed.Dir) {
		return nil, errModuleDependency
	}
	files := make([]string, 0, len(listed.GoFiles))
	total := int64(0)
	for _, name := range listed.GoFiles {
		if name == "" || filepath.Base(name) != name {
			return nil, errModuleDependency
		}
		path := filepath.Join(listed.Dir, name)
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > moduleFileMax {
			if err == nil && info.Size() > moduleFileMax {
				return nil, moduleFailure(moduleCauseSourceLimit)
			}
			return nil, errModuleDependency
		}
		total += info.Size()
		if total > moduleSourceMax {
			return nil, moduleFailure(moduleCauseSourceLimit)
		}
		files = append(files, path)
	}
	return files, nil
}

func (imports *moduleImporter) exportData(path string) (io.ReadCloser, error) {
	if !validModuleImportPath(path) || imports.root == "" {
		return nil, errModuleDependency
	}
	exportPath, ok := imports.exportPaths[path]
	if !ok {
		if err := imports.loadStandardExports(path); err != nil {
			return nil, preserveModuleFailure(err)
		}
		exportPath, ok = imports.exportPaths[path]
		if !ok {
			return nil, errModuleDependency
		}
	}
	file, err := os.Open(exportPath)
	if err != nil {
		return nil, errModuleDependency
	}
	return file, nil
}

func (imports *moduleImporter) loadStandardExports(path string) error {
	args := append([]string{"list"}, imports.moduleListCommonArgs()...)
	args = append(args, "-deps", "-export", "-json", path)
	output, err := imports.runGoList(args)
	if err != nil {
		return preserveModuleFailure(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	loaded := make(map[string]string)
	found := false
	for count := 0; ; count++ {
		var listed moduleListPackage
		err := decoder.Decode(&listed)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil || count >= moduleListPackagesMax || listed.ImportPath == "" || listed.Incomplete || !listed.Goroot {
			if count >= moduleListPackagesMax {
				return moduleFailure(moduleCausePackageLimit)
			}
			return errModuleDependency
		}
		if listed.ImportPath == "unsafe" {
			if listed.ImportPath == path {
				found = true
			}
			continue
		}
		if !filepath.IsAbs(listed.Export) {
			return errModuleDependency
		}
		info, statErr := os.Stat(listed.Export)
		if statErr != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > moduleExportMax {
			return errModuleDependency
		}
		loaded[listed.ImportPath] = listed.Export
		if listed.ImportPath == path {
			found = true
		}
	}
	if !found {
		return errModuleDependency
	}
	newPackages := 0
	for packagePath := range loaded {
		if _, ok := imports.exportPaths[packagePath]; !ok {
			newPackages++
		}
	}
	if newPackages > moduleListPackagesMax-imports.exportPackages {
		return moduleFailure(moduleCausePackageLimit)
	}
	maps.Copy(imports.exportPaths, loaded)
	imports.exportPackages += newPackages
	return nil
}

func (imports *moduleImporter) listModulePackage(path string, export bool) (moduleListPackage, error) {
	if !validModuleImportPath(path) || imports.root == "" {
		return moduleListPackage{}, errModuleDependency
	}
	args := append([]string{"list"}, imports.moduleListCommonArgs()...)
	args = append(args, "-json")
	if export {
		args = append(args, "-export")
	}
	args = append(args, path)
	output, err := imports.runGoList(args)
	if err != nil {
		return moduleListPackage{}, err
	}
	var listed moduleListPackage
	decoder := json.NewDecoder(bytes.NewReader(output))
	if err := decoder.Decode(&listed); err != nil {
		return moduleListPackage{}, errModuleDependency
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return moduleListPackage{}, errModuleDependency
	}
	return listed, nil
}

func (imports *moduleImporter) moduleListCommonArgs() []string {
	args := []string{"-mod=readonly", "-modfile=" + filepath.Join(imports.absoluteRoot(), "go.mod"), "-compiler=" + build.Default.Compiler}
	if len(build.Default.BuildTags) != 0 {
		args = append(args, "-tags="+strings.Join(build.Default.BuildTags, ","))
	}
	return args
}

func (imports *moduleImporter) absoluteRoot() string {
	root, err := filepath.Abs(imports.root)
	if err != nil {
		return ""
	}
	return root
}

func (imports *moduleImporter) runGoList(args []string) ([]byte, error) {
	if imports.listCalls >= moduleListCallsMax {
		return nil, moduleFailure(moduleCauseCallBudget)
	}
	if imports.listBytes >= moduleListTotalMax {
		return nil, moduleFailure(moduleCauseByteBudget)
	}
	if time.Since(imports.listStarted) >= moduleListTotalTimeout {
		return nil, moduleFailure(moduleCauseTimeBudget)
	}
	root := imports.absoluteRoot()
	if root == "" {
		return nil, errModuleDependency
	}
	goPath := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goPath += ".exe"
	}
	remaining := moduleListTotalTimeout - time.Since(imports.listStarted)
	callTimeout := min(remaining, moduleListTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, goPath, args...)
	command.Dir = root
	command.Env = moduleListEnvironment(os.Environ())
	configureModuleListProcess(command)
	remainingBytes := int(moduleListTotalMax - imports.listBytes)
	stderrLimit := min(32<<10, remainingBytes)
	stdoutLimit := min(moduleListMax, max(0, remainingBytes-stderrLimit))
	stdout := &boundedBuffer{limit: stdoutLimit}
	stderr := &boundedBuffer{limit: stderrLimit}
	command.Stdout = stdout
	command.Stderr = stderr
	imports.listCalls++
	runErr := command.Run()
	written := int64(stdout.Len() + stderr.Len())
	imports.listBytes += written
	if written > moduleListTotalMax || imports.listBytes > moduleListTotalMax {
		return nil, moduleFailure(moduleCauseOutputCap)
	}
	if failure := moduleListCommandFailure(runErr, ctx.Err(), stdout.exceeded, stderr.exceeded); failure != nil {
		return nil, failure
	}
	return stdout.Bytes(), nil
}

func moduleListCommandFailure(runErr, contextErr error, stdoutExceeded, stderrExceeded bool) error {
	if errors.Is(contextErr, context.DeadlineExceeded) {
		return moduleFailure(moduleCauseTimeout)
	}
	if stdoutExceeded || stderrExceeded {
		return moduleFailure(moduleCauseOutputCap)
	}
	if runErr != nil || contextErr != nil {
		return moduleFailure(moduleCauseSubprocess)
	}
	return nil
}

func moduleListEnvironment(parent []string) []string {
	controlled := map[string]string{
		"GOWORK": "off", "GOTOOLCHAIN": "local", "GOROOT": runtime.GOROOT(),
		"GOOS": build.Default.GOOS, "GOARCH": build.Default.GOARCH,
		"GOFLAGS": "", "GOEXPERIMENT": "",
		"GOARM": "", "GOARM64": "", "GOAMD64": "", "GO386": "",
		"GOMIPS": "", "GOMIPS64": "", "GOPPC64": "", "GORISCV64": "", "GOWASM": "",
	}
	maps.Copy(controlled, moduleListArchitectureSettings())
	if build.Default.CgoEnabled {
		controlled["CGO_ENABLED"] = "1"
	} else {
		controlled["CGO_ENABLED"] = "0"
	}
	result := make([]string, 0, len(parent)+len(controlled))
	for _, entry := range parent {
		name, _, ok := strings.Cut(entry, "=")
		if _, replace := controlled[name]; !ok || replace {
			continue
		}
		result = append(result, entry)
	}
	for name, value := range controlled {
		result = append(result, name+"="+value)
	}
	return result
}

func moduleListArchitectureSettings() map[string]string {
	settings := map[string]string{
		"GOARM": "", "GOARM64": "", "GOAMD64": "", "GO386": "",
		"GOMIPS": "", "GOMIPS64": "", "GOPPC64": "", "GORISCV64": "", "GOWASM": "",
	}
	for _, tag := range build.Default.ToolTags {
		switch build.Default.GOARCH {
		case "386":
			if after, ok := strings.CutPrefix(tag, "386."); ok {
				settings["GO386"] = after
			}
		case "amd64":
			if value, ok := strings.CutPrefix(tag, "amd64.v"); ok {
				if level, err := strconv.Atoi(value); err == nil {
					current, _ := strconv.Atoi(strings.TrimPrefix(settings["GOAMD64"], "v"))
					if level > current {
						settings["GOAMD64"] = "v" + value
					}
				}
			}
		case "arm":
			if value, ok := strings.CutPrefix(tag, "arm."); ok {
				if level, err := strconv.Atoi(value); err == nil {
					current, _ := strconv.Atoi(settings["GOARM"])
					if level > current {
						settings["GOARM"] = value
					}
				}
			}
		case "arm64":
			if value := strings.TrimPrefix(tag, "arm64.v"); value != tag && newerArchitectureVersion(value, strings.TrimPrefix(settings["GOARM64"], "v")) {
				settings["GOARM64"] = "v" + value
			}
		case "mips", "mipsle":
			if after, ok := strings.CutPrefix(tag, build.Default.GOARCH+"."); ok {
				settings["GOMIPS"] = after
			}
		case "mips64", "mips64le":
			if after, ok := strings.CutPrefix(tag, build.Default.GOARCH+"."); ok {
				settings["GOMIPS64"] = after
			}
		case "ppc64", "ppc64le":
			if value, ok := strings.CutPrefix(tag, "ppc64.power"); ok {
				if level, err := strconv.Atoi(value); err == nil {
					current, _ := strconv.Atoi(strings.TrimPrefix(settings["GOPPC64"], "power"))
					if level > current {
						settings["GOPPC64"] = "power" + value
					}
				}
			}
		case "riscv64":
			if value, ok := strings.CutPrefix(tag, "riscv64."); ok {
				settings["GORISCV64"] = value
			}
		case "wasm":
			if after, ok := strings.CutPrefix(tag, "wasm."); ok {
				value := after
				if settings["GOWASM"] == "" {
					settings["GOWASM"] = value
				} else {
					settings["GOWASM"] += "," + value
				}
			}
		}
	}
	return settings
}

func newerArchitectureVersion(candidate, current string) bool {
	majorMinor := func(value string) (int, int) {
		majorText, minorText, ok := strings.Cut(value, ".")
		if !ok {
			return 0, 0
		}
		major, majorErr := strconv.Atoi(majorText)
		minor, minorErr := strconv.Atoi(minorText)
		if majorErr != nil || minorErr != nil {
			return 0, 0
		}
		return major, minor
	}
	candidateMajor, candidateMinor := majorMinor(candidate)
	currentMajor, currentMinor := majorMinor(current)
	return candidateMajor > currentMajor || (candidateMajor == currentMajor && candidateMinor > currentMinor)
}

type moduleListPackage struct {
	Dir        string   `json:"Dir"`
	ImportPath string   `json:"ImportPath"`
	Export     string   `json:"Export"`
	GoFiles    []string `json:"GoFiles"`
	CgoFiles   []string `json:"CgoFiles"`
	Incomplete bool     `json:"Incomplete"`
	Goroot     bool     `json:"Goroot"`
}

type boundedBuffer struct {
	bytes.Buffer
	limit    int
	exceeded bool
}

func (buffer *boundedBuffer) Write(value []byte) (int, error) {
	remaining := buffer.limit - buffer.Len()
	if remaining <= 0 {
		buffer.exceeded = true
		return len(value), nil
	}
	if len(value) > remaining {
		buffer.exceeded = true
		_, _ = buffer.Buffer.Write(value[:remaining])
		return len(value), nil
	}
	_, _ = buffer.Buffer.Write(value)
	return len(value), nil
}

func validModuleImportPath(path string) bool {
	if path == "" || strings.HasPrefix(path, "-") || strings.ContainsAny(path, "\\\"'<>?*|[]{}:;@\n\r\t ") {
		return false
	}
	for segment := range strings.SplitSeq(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func readBoundedModuleSource(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errModuleDependency
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, moduleFileMax+1))
	if err != nil {
		return nil, errModuleDependency
	}
	if len(data) > moduleFileMax {
		return nil, moduleFailure(moduleCauseSourceLimit)
	}
	return data, nil
}

func (imports *moduleImporter) local(path string) bool {
	return imports.module != "" && (path == imports.module || strings.HasPrefix(path, imports.module+"/"))
}

func (imports *moduleImporter) files(path string) ([]string, error) {
	if !validModuleImportPath(path) || !imports.local(path) {
		return nil, errModuleDependency
	}
	root, err := filepath.Abs(imports.root)
	if err != nil {
		return nil, errModuleDependency
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, errModuleDependency
	}
	relative := strings.TrimPrefix(path, imports.module)
	directory := filepath.Join(imports.root, filepath.FromSlash(strings.TrimPrefix(relative, "/")))
	directory, ok := canonicalContainedPath(root, directory)
	if !ok {
		return nil, errModuleDependency
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, errModuleDependency
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		name, ok := canonicalContainedPath(root, filepath.Join(directory, entry.Name()))
		if !ok {
			return nil, errModuleDependency
		}
		info, err := os.Stat(name)
		if err != nil || !info.Mode().IsRegular() || !moduleBuildFile(directory, entry.Name()) {
			if err != nil {
				return nil, errModuleDependency
			}
			continue
		}
		result = append(result, name)
	}
	return result, nil
}

func canonicalContainedPath(root, target string) (string, bool) {
	target, err := filepath.Abs(target)
	if err != nil {
		return "", false
	}
	canonical, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", false
	}
	relative, err := filepath.Rel(root, canonical)
	if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}
	return canonical, true
}

func moduleBuildFile(directory, name string) bool {
	matched, err := build.Default.MatchFile(directory, name)
	return err == nil && matched
}

func readModulePath(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	return ""
}
