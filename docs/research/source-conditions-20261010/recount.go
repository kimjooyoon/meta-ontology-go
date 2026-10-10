// Recount the paired native observations without a model or code generation.
// Usage: go run recount.go <directory containing the original files or .gz>
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func read(path string) []byte {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		raw, err = os.ReadFile(path + ".gz")
		must(err)
		z, err := gzip.NewReader(bytes.NewReader(raw))
		must(err)
		defer z.Close()
		raw, err = io.ReadAll(z)
	}
	must(err)
	return raw
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func require(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func object(v any) map[string]any { return v.(map[string]any) }
func rows(v any) []any            { return v.([]any) }
func number(v any) int64 {
	n, err := v.(json.Number).Int64()
	must(err)
	return n
}
func count(v any) int64 {
	if v == nil {
		return 0
	}
	return number(v)
}
func at(v any, keys ...string) any {
	for _, key := range keys {
		v = object(v)[key]
	}
	return v
}
func decode(path string) map[string]any {
	var result map[string]any
	d := json.NewDecoder(bytes.NewReader(read(path)))
	d.UseNumber()
	must(d.Decode(&result))
	return result
}
func digest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }

func nativeRows(record map[string]any) []any {
	runtime := object(at(record, "evaluation", "runtime"))
	var out []any
	for _, trace := range rows(runtime["traces"]) {
		for _, delivery := range rows(at(trace, "deliveries")) {
			d := object(delivery)
			actual, expected := number(d["actual"]), number(d["expected"])
			require(actual == expected && d["passed"] == true, "native output differs")
			out = append(out, map[string]any{"input": number(d["input"]), "actual": actual, "expected": expected})
		}
	}
	require(len(out) == 11 && count(runtime["finite_passed"]) == 11 && count(runtime["finite_total"]) == 11, "native finite count differs")
	require(at(out[10], "actual") == int64(18014398509481990), "large integer was rounded")
	return out
}

func integer(e ast.Expr, input int64) int64 {
	switch v := e.(type) {
	case *ast.Ident:
		if v.Name == "input" {
			return input
		}
	case *ast.BasicLit:
		if v.Kind == token.INT {
			n, err := strconv.ParseInt(v.Value, 0, 64)
			must(err)
			return n
		}
	case *ast.ParenExpr:
		return integer(v.X, input)
	}
	panic("audit requires a literal or input operand")
}

func predicate(raw []byte) map[string]any {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", raw, 0)
	must(err)
	var found []*ast.IfStmt
	ast.Inspect(file, func(n ast.Node) bool {
		if s, ok := n.(*ast.IfStmt); ok {
			found = append(found, s)
		}
		return true
	})
	require(len(found) == 1, "audit is scoped to one emitted if")
	e, ok := found[0].Cond.(*ast.BinaryExpr)
	require(ok && e.Op == token.LSS, "audit requires less-than")
	var rendered bytes.Buffer
	must(format.Node(&rendered, fset, e))
	passed := 0
	var probes []any
	for _, input := range []int64{-1, 0, 1} {
		actual, expected := integer(e.X, input) < integer(e.Y, input), input < 0
		if actual == expected {
			passed++
		}
		probes = append(probes, map[string]any{"input": input, "actual": actual, "expected": expected})
	}
	return map[string]any{"expression": rendered.String(), "passed": passed, "total": 3, "probes": probes, "generated_sha256": digest(raw)}
}

func processTime(raw []byte) map[string]any {
	lines := strings.Split(string(raw), "\n")
	var wall, user, system float64
	_, err := fmt.Sscanf(lines[0], "%f real %f user %f sys", &wall, &user, &system)
	must(err)
	var rss int64
	_, err = fmt.Sscanf(lines[1], "%d", &rss)
	must(err)
	return map[string]any{"wall_seconds": wall, "user_seconds": user, "system_seconds": system,
		"maximum_resident_bytes": rss, "aggregate_cpu_seconds_over_wall_percent": 100 * (user + system) / wall,
		"scope": "one CLI invocation including native Go build/run subprocesses; not host utilization or isolated inference"}
}

func main() {
	require(len(os.Args) == 2, "observation directory required")
	root := os.Args[1]
	var original []string
	for _, line := range strings.Split(string(read(filepath.Join(root, "constrained.gooo"))), "\n") {
		if !strings.HasPrefix(line, "    condition_case ") {
			original = append(original, line)
		}
	}
	require(bytes.Equal([]byte(strings.Join(original, "\n")), read(filepath.Join(root, "baseline.gooo"))), "paired sources differ beyond condition declarations")
	build := decode(filepath.Join(root, "compiler-build.json"))
	require(build["vcs_revision"] == "62f2252daf8f250e350db53ccea0dc6abe1d7571" && build["vcs_modified"] == "false", "producer source differs")
	var results []any
	for _, arm := range []string{"baseline", "constrained"} {
		fresh := decode(filepath.Join(root, arm+".stdout.json"))
		replay := decode(filepath.Join(root, arm+"-replay.stdout.json"))
		freshRows, replayRows := nativeRows(fresh), nativeRows(replay)
		left, _ := json.Marshal(freshRows)
		right, _ := json.Marshal(replayRows)
		require(bytes.Equal(left, right), "saved replay changed outputs")
		require(at(replay, "evaluation", "construction_replayed") == true && count(at(replay, "evaluation", "new_model_calls")) == 0, "replay called model or regenerated")
		construction := at(fresh, "construction")
		initial := at(rows(at(construction, "initial", "preparations"))[0], "generation", "report", "body_paths")
		search := object(at(initial, "search"))
		selection := object(search["selection"])
		require(selection["model_weights_sha256"] == "dcd8e44591626d421d4961bfef82ec197e947cb1d5d2cd92868d908bc7de4aed", "frozen model differs")
		var ns int64
		for _, receipt := range rows(selection["receipts"]) {
			ns += number(at(receipt, "predict_ns"))
		}
		require(count(selection["local_model_predictions"]) == 3, "model prediction count differs")
		attempts := rows(at(construction, "attempts"))
		final := at(rows(at(attempts[number(at(construction, "selected_attempt"))], "path_candidates"))[0], "conditions")
		if arm == "constrained" {
			require(count(at(final, "passed")) == 3 && count(at(final, "declared")) == 3, "caller lost source conditions")
			for _, row := range rows(at(final, "results")) {
				require(at(row, "observation", "reached") == true && at(row, "observation", "value") == at(row, "case", "expected") && at(row, "passed") == true, "condition row differs")
			}
		}
		results = append(results, map[string]any{"arm": arm, "initial_choices": selection["choices"],
			"initial_proposals": search["initial_proposals"], "initial_evaluated_candidates": search["evaluated_candidates"],
			"initial_condition_rejected_candidates": count(search["condition_rejected_candidates"]),
			"model_predictions":                     selection["local_model_predictions"], "prediction_total_ns": ns,
			"model_weights_sha256": selection["model_weights_sha256"], "construction_attempts": len(attempts),
			"final_condition_receipt": final, "final_predicate_audit": predicate(read(filepath.Join(root, arm, "generated.go"))),
			"native_passed": len(freshRows), "native_total": len(freshRows), "exact_native_rows": freshRows,
			"saved_replay_equal": true, "saved_replay_new_model_calls": 0,
			"process": processTime(read(filepath.Join(root, arm+".time.log")))})
	}
	encoded, err := json.MarshalIndent(map[string]any{"schema": "gooo/source-condition-paired-study/v1", "results": results,
		"scope": "same compiler, frozen model and cases; only three source condition_case declarations added; one timing observation per arm; no training or model-quality improvement claim"}, "", "  ")
	must(err)
	fmt.Println(string(encoded))
}
