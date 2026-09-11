package policycompilation

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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
	input, err := json.Marshal(map[string]string{"id": "strict-eof"})
	if err != nil {
		t.Fatal(err)
	}
	run := func(suffix string) error {
		command := exec.Command("go", "run", judgePath)
		command.Stdin = bytes.NewBuffer(append(append(append([]byte{}, input...), '\n'), []byte(suffix)...))
		command.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
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
