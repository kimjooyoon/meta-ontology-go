package policycompilation

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeStrictJSONRejectsDuplicateObjectKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
		path  string
	}{
		{
			name:  "same value",
			input: `{"value":"same","value":"same"}`,
			path:  `$["value"]`,
		},
		{
			name:  "conflicting values",
			input: `{"value":"first","value":"second"}`,
			path:  `$["value"]`,
		},
		{
			name:  "escaped key alias",
			input: `{"value":"first","va\u006cue":"second"}`,
			path:  `$["value"]`,
		},
		{
			name:  "nested object",
			input: `{"outer":{"value":"first","value":"second"}}`,
			path:  `$["outer"]["value"]`,
		},
		{
			name:  "nested array object",
			input: `{"items":[{"value":"first","value":"second"}]}`,
			path:  `$["items"][0]["value"]`,
		},
	}
	type nestedValue struct {
		Value string `json:"value"`
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var value struct {
				Value string        `json:"value"`
				Outer nestedValue   `json:"outer"`
				Items []nestedValue `json:"items"`
			}
			err := decodeStrictJSON([]byte(test.input), &value)
			if err == nil {
				t.Fatalf("duplicate object key input %q was accepted", test.input)
			}
			if !strings.Contains(err.Error(), `duplicate JSON object key "value"`) {
				t.Fatalf("error = %q, want duplicate-key diagnostic", err)
			}
			if !strings.Contains(err.Error(), test.path) {
				t.Fatalf("error = %q, want location %q", err, test.path)
			}
		})
	}
}

func TestDecodeStrictJSONAllowsSeparateObjectsAndExplicitZeroValues(t *testing.T) {
	var values []struct {
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`[{"value":"first"},{"value":"second"}]`), &values); err != nil {
		t.Fatalf("same key in separate objects = %v, want valid document", err)
	}
	if len(values) != 2 || values[0].Value != "first" || values[1].Value != "second" {
		t.Fatalf("separate object values = %#v, want both values preserved", values)
	}

	var explicit struct {
		Flag  bool   `json:"flag"`
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`{"flag":false,"value":""}`), &explicit); err != nil {
		t.Fatalf("explicit false/empty values = %v, want valid document", err)
	}
	if explicit.Flag || explicit.Value != "" {
		t.Fatalf("explicit zero values = %#v, want false and empty", explicit)
	}
	var missing struct {
		Flag  bool   `json:"flag"`
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`{}`), &missing); err != nil {
		t.Fatalf("missing values = %v, want valid document", err)
	}
	if missing.Flag || missing.Value != "" {
		t.Fatalf("missing values = %#v, want zero values", missing)
	}
}

func TestDecodeStrictJSONRetainsUnknownFieldRejection(t *testing.T) {
	var value struct {
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte(`{"value":"ok","unknown":true}`), &value); err == nil {
		t.Fatal("unknown field was accepted")
	}
}

func TestDecodeStrictJSONAcceptsWhitespaceAndRejectsTrailingDocuments(t *testing.T) {
	var value struct {
		Value string `json:"value"`
	}
	if err := decodeStrictJSON([]byte("{\"value\":\"ok\"}\n \t"), &value); err != nil || value.Value != "ok" {
		t.Fatalf("whitespace suffix = %#v, want valid document", err)
	}
	for _, input := range []string{"{\"value\":\"ok\"}{}", "{\"value\":\"ok\"} trailing"} {
		var got struct {
			Value string `json:"value"`
		}
		if err := decodeStrictJSON([]byte(input), &got); err == nil {
			t.Fatalf("trailing input %q was accepted", input)
		}
	}
}

func TestGeneratedJudgeRejectsTrailingDocuments(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	judgePath := filepath.Join(t.TempDir(), "judge.go")
	if err := os.WriteFile(judgePath, GenerateJudge(policy), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(filepath.Dir(judgePath), "judge")
	build := exec.Command("go", "build", "-o", binaryPath, judgePath)
	build.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generated judge: %v: %s", err, output)
	}
	input, err := json.Marshal(map[string]string{"id": "strict-eof"})
	if err != nil {
		t.Fatal(err)
	}
	run := func(suffix string) error {
		command := exec.Command(binaryPath)
		command.Stdin = bytes.NewBuffer(append(append(append([]byte{}, input...), '\n'), []byte(suffix)...))
		return command.Run()
	}
	if err := run(" \t\n"); err != nil {
		t.Fatalf("whitespace suffix rejected: %v", err)
	}
	if err := run("{}\n"); err == nil {
		t.Fatal("second JSON document was accepted")
	}
	if err := run("trailing"); err == nil {
		t.Fatal("malformed trailing bytes were accepted")
	}
}

func TestGeneratedJudgeRejectsDuplicateObjectKeys(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	tempDir := t.TempDir()
	judgePath := filepath.Join(tempDir, "judge.go")
	if err := os.WriteFile(judgePath, GenerateJudge(policy), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(tempDir, "judge")
	build := exec.Command("go", "build", "-o", binaryPath, judgePath)
	build.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generated judge: %v: %s", err, output)
	}
	run := func(input string) (string, error) {
		command := exec.Command(binaryPath)
		command.Stdin = strings.NewReader(input)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		err := command.Run()
		return stderr.String(), err
	}
	for _, test := range []struct {
		name  string
		input string
		key   string
		path  string
	}{
		{name: "same value", input: `{"id":"same","id":"same"}`, key: "id", path: `$["id"]`},
		{name: "escaped key alias", input: `{"id":"first","i\u0064":"second"}`, key: "id", path: `$["id"]`},
		{name: "repeated boolean", input: `{"producer_available":false,"producer_available":true}`, key: "producer_available", path: `$["producer_available"]`},
	} {
		stderr, err := run(test.input)
		t.Run(test.name, func(t *testing.T) {
			if err == nil {
				t.Fatalf("duplicate object key input %q was accepted", test.input)
			}
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) || exitError.ExitCode() != 2 {
				t.Fatalf("error = %v, want process exit status 2", err)
			}
			if !strings.Contains(stderr, `duplicate JSON object key "`+test.key+`"`) {
				t.Fatalf("stderr = %q, want duplicate-key diagnostic", stderr)
			}
			if !strings.Contains(stderr, test.path) || !strings.Contains(stderr, "byte offset") {
				t.Fatalf("stderr = %q, want path %q and byte offsets", stderr, test.path)
			}
		})
	}
	if stderr, err := run(`{"id":"","producer_available":false,"consumer_available":false}`); err != nil {
		t.Fatalf("explicit false/empty generated input = %v, want accepted", err)
	} else if stderr != "" {
		t.Fatalf("explicit false/empty generated input stderr = %q, want empty", stderr)
	}
}
