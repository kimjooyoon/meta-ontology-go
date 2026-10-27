package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const originProbeSource = `package origins
namespace origins
entity Value id "urn:origin:value" fields {
 field number id "urn:origin:number" type integer required one
}
activity Divide(Value) -> Value id "urn:origin:divide" computes "return Value{number:10/input.number}"
// 한국어 <>& "quoted"
`

func TestPackageGenerateEmbeddedOriginSurvivesNativeImport(t *testing.T) {
	fixture, consumer := t.TempDir(), t.TempDir()
	writeWorkspaceSource(t, fixture, "main.gooo", originProbeSource)
	manifest := writeWorkspaceManifest(t, fixture, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origins","activity":"Divide"},"packages":[{"path":"origins","sources":["main.gooo"]}]}`)
	var out, diagnostics bytes.Buffer
	t.Run("generate-without-tools", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if code := run([]string{"package", "generate", "--go-package", "origins", "--embed-origin", "--export-stage", "Divide=urn:origin:divide", manifest}, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String())
		}
	})
	origin := readLibraryOriginConstant(t, out.Bytes())
	var value struct {
		Schema, EntryID, Lowered, Plan string
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(origin), &fields); err != nil {
		t.Fatal(err)
	}
	for key, target := range map[string]*string{"schema": &value.Schema, "entry_activity_id": &value.EntryID, "lowered_source_sha256": &value.Lowered, "typed_plan_sha256": &value.Plan} {
		if err := json.Unmarshal(fields[key], target); err != nil {
			t.Fatal(key, err)
		}
	}
	if value.Schema != "gooo/library-generation-origin/v1" || value.EntryID != "urn:origin:divide" || !strings.Contains(out.String(), "// Lowered Gooo source: "+value.Lowered+"\n") || !strings.Contains(out.String(), "// Typed plan: "+value.Plan+"\n") {
		t.Fatal("compiled identity and generated view disagree", value)
	}
	checkEmbeddedOriginalSource(t, fields["context"], manifest)
	var stages []struct {
		Name string `json:"name"`
		ID   string `json:"activity_id"`
	}
	if err := json.Unmarshal(fields["stage_exports"], &stages); err != nil || len(stages) != 1 || stages[0].Name != "Divide" || stages[0].ID != "urn:origin:divide" {
		t.Fatal("export lost its source identity", stages, err)
	}
	writeWorkspaceSource(t, consumer, "go.mod", "module imported-origin\n\ngo 1.27.3\n")
	writeWorkspaceSource(t, consumer, "generated.go", out.String())
	writeWorkspaceSource(t, consumer, "origin_test.go", originNativeConsumer)
	writeWorkspaceSource(t, consumer, "expected.json", origin)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", ".")
	command.Dir, command.Env = consumer, append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
}

func checkEmbeddedOriginalSource(t *testing.T, raw json.RawMessage, manifest string) {
	t.Helper()
	var context struct {
		Manifest       string `json:"workspace_manifest_sha256"`
		SourceStatus   string `json:"compiler_source_status"`
		CompilerSource string `json:"compiler_source_sha"`
		Sources        []struct {
			Package  string `json:"package_path"`
			Filename string `json:"filename"`
			SHA      string `json:"source_json_string_sha256"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(raw, &context); err != nil {
		t.Fatal(err)
	}
	manifestBytes, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(originProbeSource)
	if context.Manifest != fmt.Sprintf("sha256:%x", sha256.Sum256(manifestBytes)) || len(context.Sources) != 1 {
		t.Fatal(context)
	}
	source := context.Sources[0]
	if source.Package != "origins" || source.Filename != "main.gooo" || source.SHA != fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)) {
		t.Fatal(source)
	}
	if context.SourceStatus != "CLEAN_VCS" && context.CompilerSource != "UNBOUND_LOCAL_SOURCE" {
		t.Fatal("invented clean compiler source", context)
	}
}

func readLibraryOriginConstant(t *testing.T, source []byte) string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, spec := range group.Specs {
			value := spec.(*ast.ValueSpec)
			if len(value.Names) == 1 && value.Names[0].Name == "GoooGenerationOrigin" {
				literal, ok := value.Values[0].(*ast.BasicLit)
				if !ok {
					t.Fatal("origin is not a constant literal")
				}
				decoded, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				return decoded
			}
		}
	}
	t.Fatal("generation origin constant absent")
	return ""
}
