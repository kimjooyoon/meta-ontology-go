package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestOriginalAssemblyFiles12503(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	for _, mode := range []string{"candidate", "grammar"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			files := assemblyFiles12503(mode)
			for name, source := range files {
				if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			var output, diagnostics bytes.Buffer
			if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "w.json")}, &output, &diagnostics); code != exitOK {
				t.Fatal(mode, code, diagnostics.String(), output.String())
			}
			saveLocalFiles12421(t, "producer12503-"+mode+".json", output.Bytes())
			var receipt packageExecutionReceipt
			if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 2 {
				t.Fatal(mode, err)
			}
			rows, _, err := bodyexecution.SelectRecordedNativeOrigins(receipt.Result.Runtime, bodyexecution.NativeOriginQuery{Kind: "local-input-locations", Limit: 128})
			if err != nil || len(rows) == 0 {
				t.Fatal(err)
			}
			checkAssemblyFileRows12503(t, mode, files, rows, receipt.Result.Runtime)
			checkAssemblyFileReader12503(t, root, mode, output.Bytes(), rows)
			t.Logf("new %s producer2/2;physical source facts%d;candidate/grammar context and retained declaration/use locations distinguished;Gooo reader output%d matched;original producer/model calls during reads0", mode, len(rows), len(rows))
		})
	}
}

func assemblyFiles12503(mode string) map[string]string {
	if mode == "candidate" {
		return map[string]string{
			"main.gooo": `package assembledfiles
namespace assembled12503
import math "origin/math12503"
entity Integer id "urn:gooo:type:integer"
activity Build(Integer) -> Integer id "urn:assembled12503:build" computes "let picked=__GOOO_BODY_HOLE_value__;let copied=picked+23;return copied" assembling {
 source_fill intent "연산을 조립하고 원래 후보 파일을 가리킨다" {
  hole "value"
  candidate "helper12503" {fill "value" "math.Adjust(input)"}
  candidate "other12503" {fill "value" "input-743"}
 }
 case "751" -> "1517"
 case "-757" -> "9"
}
`,
			"ops.gooo":   "package assemblyops\nnamespace assemblyops12503\nentity Integer id \"urn:gooo:type:integer\"\nactivity Adjust(Integer) -> Integer id \"urn:assemblyops12503:adjust\" computes `let delta=input+743;return delta`\n",
			"w.json":     `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/main12503","activity":"Build"},"packages":[{"path":"origin/main12503","name":"assembledfiles","imports":["origin/math12503"],"sources":["main.gooo"]},{"path":"origin/math12503","name":"assemblyops","imports":[],"sources":["ops.gooo"]}]}`,
			"cases.json": `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/main12503:Build":761},"expected":{"origin/main12503:Build":1527}},{"inputs":{"origin/main12503:Build":-769},"expected":{"origin/main12503:Build":-3}}]}`,
		}
	}
	return map[string]string{
		"main.gooo": `package grammarfiles
namespace grammar12503
entity Integer id "urn:gooo:type:integer"
activity Build(Integer) -> Integer id "urn:grammar12503:build" computes "let seed=__GOOO_BODY_HOLE_shift__;let copied=seed;return copied" assembling {
 source_fill intent "원본 파일의 문법 선언을 기원으로 남긴다" {
  hole "shift"
  derive grammar "integer-offset-constant/v1" max_expressions "16" max_candidates "16"
 }
 case "797" -> "1584"
 case "-809" -> "-22"
}
`,
		"w.json":     `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/grammar12503","activity":"Build"},"packages":[{"path":"origin/grammar12503","name":"grammarfiles","imports":[],"sources":["main.gooo"]}]}`,
		"cases.json": `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/grammar12503:Build":811},"expected":{"origin/grammar12503:Build":1598}},{"inputs":{"origin/grammar12503:Build":-821},"expected":{"origin/grammar12503:Build":-34}}]}`,
	}
}

func checkAssemblyFileRows12503(t *testing.T, mode string, files map[string]string, rows []json.RawMessage, runtime bodyexecution.CompositionRuntime) {
	t.Helper()
	fragment, unknown, template, declaration := false, false, false, false
	for _, raw := range rows {
		var fact bodyexecution.CompositionLocalFileFact
		if err := json.Unmarshal(raw, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.Mapping == "UNMAPPED" {
			unknown = true
			continue
		}
		source, ok := files[fact.Filename]
		if !ok || fact.SourceStart < 0 || fact.SourceEnd <= fact.SourceStart || fact.SourceEnd > len(source) {
			t.Fatal("physical source range", fact)
		}
		text := source[fact.SourceStart:fact.SourceEnd]
		if mode == "candidate" && fact.Category == "declared_candidate" && strings.Contains(text, "math.Adjust(input)") {
			fragment = true
		}
		if mode == "grammar" && fact.Category == "derived_grammar" && strings.Contains(text, "integer-offset-constant/v1") {
			fragment = true
		}
		if fact.Category == "retained_template" {
			template = true
		}
		if fact.Category == "declaration" {
			declaration = true
		}
	}
	if !fragment || !unknown || !template || !declaration {
		t.Fatal("file origin categories", mode, fragment, unknown, template, declaration)
	}
	if mode == "grammar" {
		for _, binding := range runtime.LocalFileSources.Bindings {
			if c := binding.Construction; c != nil {
				if c.CandidateSource != nil {
					t.Fatal("derived candidate acquired an authored file range", c)
				}
				for _, f := range c.Fragments {
					if f.Kind == "derived_grammar" && (f.ExpressionSource != nil || f.GrammarSource == nil) {
						t.Fatal("invented expression source", f)
					}
				}
			}
		}
	}
}

func checkAssemblyFileReader12503(t *testing.T, root, mode string, producer []byte, rows []json.RawMessage) {
	t.Helper()
	path := filepath.Join(root, "saved.json")
	if err := os.WriteFile(path, producer, 0600); err != nil {
		t.Fatal(err)
	}
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "local-input-locations", "--origin-limit", "128", filepath.Join(sourceRoot, "examples/native-origin-reader/local-files/reader.workspace.json")}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), output.String())
	}
	saveLocalFiles12421(t, "consumer12503-"+mode+".json", output.Bytes())
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err)
	}
	r := receipt.Result.Runtime
	if len(r.Traces) != len(rows) || r.FiniteTotal != 0 || r.ModelCalls != 0 || receipt.OriginInput.ProducerExecutions != 0 || receipt.OriginInput.ProducerModelCalls != 0 {
		t.Fatal("reader execution scope")
	}
	for index, row := range rows {
		var fact bodyexecution.CompositionLocalFileFact
		var got string
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(r.Traces[index].Deliveries[0].Actual, &got); err != nil {
			t.Fatal(err)
		}
		want := fact.Filename + ":" + fact.EventKind + ":" + fact.Category
		if fact.Mapping != "MAPPED" {
			want = fact.Mapping + ":" + fact.Reason
		}
		if got != want {
			t.Fatal(index, got, want)
		}
	}
}
