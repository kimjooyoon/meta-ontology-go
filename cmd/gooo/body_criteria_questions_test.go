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

func TestCriteriaQuestionsCLIAndHumanReport(t *testing.T) {
	source := strings.Replace(cliCriteriaSource, `input "[7]"`, `input "[0]"`, 1)
	source = strings.Replace(source, `attempts "2"`, `observe { criteria_input "[7]" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"`, 1)
	path := filepath.Join(t.TempDir(), "main.gooo")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Select", path}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Report.RecordAssembly.CriteriaQuestions.EffectiveInputs != 2 || result.Report.RecordAssembly.Criteria.Passed != 2 {
		t.Fatal("CLI omitted the acquired obligation", out.String())
	}
	c := bodyexecution.JointConstruction{Initial: bodyexecution.Composition{Steps: []bodyexecution.CompositionStep{{Generation: result}}}}
	rows := constructionCriteriaQuestionRows(c)
	if len(rows) != 4 || !strings.Contains(rows[1][1], "1 initial + 1 acquired = 2 effective") ||
		!strings.Contains(rows[3][1], "1 surviving; 1 rejected; 0 unavailable; 0 unobserved") {
		t.Fatal(rows)
	}
}

func TestCriteriaQuestionsCLIUnavailableCandidatesRetainHistory(t *testing.T) {
	source := strings.ReplaceAll(cliCriteriaSource, "value:input", "value:input/(input-input)")
	source = strings.Replace(source, `alternative "0-input"`, `alternative "input/(input-input)"`, 1)
	source = strings.Replace(source, `attempts "2"`, `observe { criteria_input "[11]" max_candidates "2" max_rounds "1" }
 attempts "2"`, 1)
	path := filepath.Join(t.TempDir(), "main.gooo")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	code := run([]string{"body-codegen", "--json", "--activity", "Select", path}, &out, &diagnostics)
	if code == exitOK || !strings.Contains(out.String(), `"criteria_questions"`) || !strings.Contains(out.String(), `"EVALUATION_UNAVAILABLE"`) {
		t.Fatal("CLI discarded failed question evidence", code, out.String(), diagnostics.String())
	}
}
