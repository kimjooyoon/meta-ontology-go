package main

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func rejects(t *testing.T, check func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("altered evidence was accepted")
		}
	}()
	check()
}

func TestRecordedContractInputsAndProgress(t *testing.T) {
	root := filepath.Join("..", "result", names[0])
	pre := decode[bodycodegen.TypedPathContextExport](filepath.Join(root, "preflight.json.gz"))
	doc := decode[pathplan.Document](filepath.Join(root, "original-document.json.gz"))
	result := decode[bodycodegen.Result](filepath.Join(root, "initial.json.gz"))
	checkInputs(pre, result.Report.BodyPaths, doc)
	for _, field := range []string{"int64_actual", "int64_both", "case_feature", "ranking_link", "attempt_output", "condition_outcome", "model_schema"} {
		t.Run(field, func(t *testing.T) { rejectMutation(t, root, field) })
	}
}

func rejectMutation(t *testing.T, root, field string) {
	t.Helper()
	pre := decode[bodycodegen.TypedPathContextExport](filepath.Join(root, "preflight.json.gz"))
	doc := decode[pathplan.Document](filepath.Join(root, "original-document.json.gz"))
	r := decode[bodycodegen.Result](filepath.Join(root, "initial.json.gz"))
	p := r.Report.BodyPaths
	switch field {
	case "int64_actual":
		p.NativeCases[0].Actual++
	case "int64_both":
		p.NativeCases[0].Actual++
		p.NativeCases[0].Expected++
	case "case_feature":
		pre.ContractCases.Features[7][0]++
	case "ranking_link":
		p.ContractProgress[1].RankingSHA = "altered"
	case "attempt_output":
		p.Search.Attempts[0].Results[7].Expected++
	case "condition_outcome":
		p.Search.Attempts[0].Conditions[0].Passed = !p.Search.Attempts[0].Conditions[0].Passed
	case "model_schema":
		pre.ModelCompatibility.Model.ModelSchema = "wrong-schema"
	}
	rejects(t, func() { checkInputs(pre, p, doc) })
}
