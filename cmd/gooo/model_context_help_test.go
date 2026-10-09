package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRecordModelHelpConnectsInspectionAssemblyAndReplay(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "http://invalid.example/never")
	for _, args := range [][]string{{"help", "body-context"}, {"body-context", "--help"},
		{"help", "body-compose"}, {"body-compose", "--help"}, {"help", "models"}} {
		var out, stderr bytes.Buffer
		if run(args, &out, &stderr) != exitOK || stderr.Len() != 0 ||
			!strings.Contains(out.String(), "examples/scalar-identity/model/model.json") {
			t.Fatal("implemented model route has no executable guide", args, out.String(), stderr.String())
		}
	}
	var out, stderr bytes.Buffer
	if run([]string{"help", "body-context"}, &out, &stderr) != exitOK ||
		!strings.Contains(out.String(), bodyContextUsage) || !strings.Contains(out.String(), "READY_FOR_RANKING") ||
		!strings.Contains(out.String(), "THREE_FIELD_CHOICES_REQUIRED") || !strings.Contains(out.String(), "zero predictions") {
		t.Fatal("input contract and decline boundary missing", out.String())
	}
	out.Reset()
	if run([]string{"help", "body-compose"}, &out, &stderr) != exitOK || !strings.Contains(out.String(), bodyComposeUsage) ||
		!strings.Contains(out.String(), "--composition out/scalar-model/composition.json") {
		t.Fatal("saved replay route missing", out.String())
	}
}

func TestRootHelpListsRecordInspectionAndExecution(t *testing.T) {
	var out, stderr bytes.Buffer
	if run([]string{"help"}, &out, &stderr) != exitOK || !strings.Contains(out.String(), "body-context") ||
		!strings.Contains(out.String(), "body-compose") || !strings.Contains(out.String(), "local model") {
		t.Fatal("root help omits current own-model workflow", out.String(), stderr.String())
	}
}
