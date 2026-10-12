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

func TestCriteriaStrategyCLIReportsSharedCounterexample(t *testing.T) {
	source := strings.Replace(cliCriteriaSource, `input "[7]"`, `input "[0]"`, 1)
	source = strings.Replace(source, `alternative "0-input"`, `alternative "input+0"`, 1)
	source = strings.Replace(source, `attempts "2"`, `observe {
 criteria_strategy "distinguish_then_check"
 criteria_input "[7]"
 max_candidates "2"
 max_rounds "1"
} attempts "2"`, 1)
	path := filepath.Join(t.TempDir(), "main.gooo")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	code := run([]string{"body-codegen", "--json", "--activity", "Select", path}, &output, &diagnostics)
	var result bodycodegen.Result
	if code != exitOK || json.Unmarshal(output.Bytes(), &result) != nil || result.Report.RecordAssembly.Criteria.Failed != 1 {
		t.Fatal(code, output.String(), diagnostics.String())
	}
	c := bodyexecution.JointConstruction{Initial: bodyexecution.Composition{Steps: []bodyexecution.CompositionStep{{Generation: result}}}}
	rows := constructionCriteriaQuestionRows(c)
	found := false
	for _, row := range rows {
		if row[0] == "Acquired criteria input" && strings.Contains(row[1], "[7]: CHECK_CRITERIA") {
			found = true
		}
	}
	if !found {
		t.Fatal("human report lost source question reason", rows)
	}
}
