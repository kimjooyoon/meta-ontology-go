package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const cliCriteriaSource = `package cli
namespace cli
entity Integer id "cli://integer"
entity Boolean id "cli://boolean"
entity Box id "cli://box" fields { field value id "cli://value" type integer required one }
activity Select(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[7]" holdout_input "[9]" }
 attempts "2"
}
activity Negative(Integer,Box) -> Boolean computes "return input1.value == 0-input0"`

func TestCriteriaCLISelectionAndFailureEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.gooo")
	for _, bad := range []bool{false, true} {
		source := cliCriteriaSource
		if bad {
			source = strings.Replace(source, "value:input", "value:input/(input-input)", 1)
			source = strings.Replace(source, `alternative "0-input"`, `alternative "input/(input-input)"`, 1)
		}
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		var out, diagnostics bytes.Buffer
		code := run([]string{"body-codegen", "--json", "--activity", "Select", path}, &out, &diagnostics)
		if !bad {
			var result bodycodegen.Result
			if code != exitOK || json.Unmarshal(out.Bytes(), &result) != nil || result.Report.RecordAssembly == nil || result.Report.RecordAssembly.Criteria.Passed != 1 {
				t.Fatal(code, out.String(), diagnostics.String())
			}
		} else if code == exitOK || !strings.Contains(out.String(), `"unknown":1`) {
			t.Fatal("CLI lost unavailable candidate conditions", code, out.String(), diagnostics.String())
		}
	}
}

func TestCriteriaReportUsesSeparatePredicateDenominator(t *testing.T) {
	c := bodyexecution.JointConstruction{SelectedAttempt: 0, Attempts: []bodyexecution.JointAttempt{{CriteriaPassed: 2, CriteriaFailed: 1, CriteriaUnknown: 1, CriteriaTotal: 4}}}
	rows := constructionCriteriaRows(c)
	if len(rows) != 1 || !strings.Contains(rows[0][1], "2/4 satisfied; 1 unsatisfied; 1 unknown") || !strings.Contains(rows[0][1], "predicate/input pairs") {
		t.Fatal(rows)
	}
}
