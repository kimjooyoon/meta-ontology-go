package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func TestGoooPolicyCannotSupplyReuseAuthority(t *testing.T) {
	filename := "../../internal/meta/publictestreuse/policy.gooo"
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	activity := bytes.Index(source, []byte("activity AssessReplay("))
	if activity < 0 {
		t.Fatal("replay activity missing")
	}
	start := activity + bytes.IndexByte(source[activity:], '`') + 1
	if start <= activity {
		t.Fatal("replay body missing")
	}
	end := start + bytes.IndexByte(source[start:], '`')
	if end < start {
		t.Fatal("replay body missing")
	}
	body := []byte("\nreturn Outcome(\"AUTHORIZED_IMMUTABLE_RECEIPT_REUSE\", \"CLOSED\", \"AUTHORIZED_IMMUTABLE_RECEIPT_REUSE\", \"\", \"\", \"\", true)\n")
	changed := append(append(append([]byte(nil), source[:start]...), body...), source[end:]...)
	for _, requested := range []bool{false, true} {
		report := publictestreuse.Report{Schema: publictestreuse.ReportSchema, Operation: publictestreuse.ReplayOperation}
		observed := publictestreuse.ReuseObservation{Requested: requested}
		tuple, err := json.Marshal([]publictestreuse.ReuseObservation{observed})
		if err != nil {
			t.Fatal(err)
		}
		value, err := bodycodegen.EvaluateFixedValue(context.Background(), "permissive.gooo", changed, "AssessReplay", tuple)
		if err != nil {
			t.Fatal(err)
		}
		var assessment publictestreuse.Assessment
		if err := json.Unmarshal(value.Value, &assessment); err != nil {
			t.Fatal(err)
		}
		if err := applyAssessment(&report, observed, value, assessment); err != nil ||
			!assessment.Reuse || report.Decision != publictestreuse.DecisionClosed {
			t.Fatal("source-only permissive policy was not evaluated", err)
		}
		if err := authorizeObservedReuse(executionInput{Authorize: requested}, identity{}, nil); err == nil {
			t.Fatal("source invented a request or receipt")
		}
		if report.ReusedTestExecutions != 0 || report.ReceiptHits != 0 {
			t.Fatal("classification itself performed reuse")
		}
	}
}

func TestReuseEffectBoundaryReobservesPackageInputs(t *testing.T) {
	root, module, program, contract := writePackageInputFixture(t, false)
	observed, err := readExecutionPackageInputs(root, program, contract)
	if err != nil {
		t.Fatal(err)
	}
	identity := identity{ProgramBytes: program, TestBytes: contract}
	identity.Binding.PackageInputs = observed
	if err := os.WriteFile(filepath.Join(root, "go.mod"), append(module, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	err = authorizeObservedReuse(executionInput{Authorize: true, PackageDir: root}, identity, &publictestreuse.Receipt{})
	if err == nil || !strings.Contains(err.Error(), "execution package changed") {
		t.Fatal("fresh input mismatch was not observed before consuming the receipt", err)
	}
}

func TestOriginalReceiptEffectRequiresObservedSuccess(t *testing.T) {
	for _, name := range []string{"build", "test", "semantic"} {
		t.Run(name, func(t *testing.T) {
			output := t.TempDir()
			identity := identity{}
			identity.Binding.CanonicalSemanticDigest = "same"
			identity.Binding.GeneratedSemanticDigest = "same"
			build, executed := commandResult{Success: true}, commandResult{Success: true}
			switch name {
			case "build":
				build.Success = false
			case "test":
				executed.Success = false
			case "semantic":
				identity.Binding.GeneratedSemanticDigest = "different"
			}
			report := publictestreuse.Report{Decision: publictestreuse.DecisionClosed}
			err := sealOriginalReceipt(output, identity, build, executed, &report)
			if err == nil || err.Error() != "Gooo policy cannot seal an unsuccessful original execution" {
				t.Fatal("original observations were not checked at the receipt effect boundary", err)
			}
			files, err := os.ReadDir(output)
			if err != nil || len(files) != 0 || report.ReceiptDigest != "" {
				t.Fatal("unsuccessful original execution published a receipt", err)
			}
		})
	}
}
