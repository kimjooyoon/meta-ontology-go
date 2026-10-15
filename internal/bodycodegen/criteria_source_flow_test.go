package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func exportCriteriaFlow(t *testing.T, source, activity string) RecordAssemblyContextExport {
	t.Helper()
	r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "criteria.gooo", []byte(source), activity, false,
		jointdecision.RecordGraphSharedFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	if r.CriteriaFlow == nil || r.CriteriaFlow.Status != "RESOLVED" || len(r.CriteriaFlow.Predicates) != 1 {
		raw, _ := json.Marshal(r.CriteriaFlow)
		t.Fatalf("source requirements unresolved: %s", raw)
	}
	if r.ModelPredictions != 0 || r.CandidateTests != 0 || r.ExpandedPlan != nil {
		t.Fatal("source export acquired outcomes", r)
	}
	return r
}

func TestCriteriaSourceFlowKeepsScalarChoicesAndTypedObligations(t *testing.T) {
	r := exportCriteriaFlow(t, scalarCriteriaSource, "Fee")
	f := r.CriteriaFlow
	p := f.Predicates[0]
	if f.Schema != "gooo/criteria-source-flow/v1" || f.CandidatePlan == nil || len(f.CandidatePlan.Decisions) != 1 ||
		f.CandidatePlan.Decisions[0].Kind != "operand_order" || len(f.CandidatePlan.ConditionCases) != 0 {
		t.Fatal("scalar source alternatives missing", f)
	}
	prepared, err := pathplan.Prepare(*f.CandidatePlan)
	if err != nil || prepared.PlanSHA256() != f.CandidatePlanSHA256 {
		t.Fatal("plan digest mismatch", err)
	}
	if p.ID != "difference" || p.ActivityID != "feecriteria://activity/difference" || p.GeneratedSHA256 == "" ||
		p.ValueFlow.Schema != "gooo/fixed-predicate-value-flow/v1" || len(p.Inputs) != 2 ||
		p.Inputs[0] != (CriterionSourceInput{Input: 1, Role: "original_input", Type: "int64"}) ||
		p.Inputs[1] != (CriterionSourceInput{Input: 2, Role: "candidate_result", Type: "int64"}) {
		t.Fatal("predicate provenance/roles missing", p)
	}
	raw, _ := json.Marshal(p.Nodes)
	if digest(raw) != p.GraphSHA256 {
		t.Fatal("graph digest mismatch")
	}
	inputs := map[uint16]bool{}
	operators := map[string]bool{}
	literal := false
	returns := 0
	for i, n := range p.Nodes {
		if n.Kind == "input" && n.InputType == "int64" {
			inputs[n.Input] = true
		}
		if n.Kind == "return" {
			returns++
		}
		operators[n.Operator] = true
		literal = literal || n.Kind == "literal" && n.Literal == "11"
		for _, parent := range []uint16{n.Parents[0], n.Parents[1], n.Guard, n.Condition} {
			if int(parent) > i {
				t.Fatalf("forward or self reference at %d: %+v", i, n)
			}
		}
	}
	if !inputs[1] || !inputs[2] || !operators["=="] || !operators["-"] || !literal || returns != 1 {
		t.Fatal("typed relationship missing", p.Nodes)
	}
	if r.Context.Status != "DECLINED_TO_DETERMINISTIC" ||
		r.Context.Reason != "SCALAR_CRITERIA_FEATURE_CONTRACT_UNAVAILABLE" || r.ValueFlow != nil {
		t.Fatal("old model contract was reinterpreted", r.Context)
	}
}

func TestCriteriaSourceFlowSeparatesPredicateEditsFromFiniteInputs(t *testing.T) {
	original := exportCriteriaFlow(t, scalarCriteriaSource, "Fee")
	changed := exportCriteriaFlow(t, strings.Replace(scalarCriteriaSource, "input1 == input0 - 11", "input1 == 11 - input0", 1), "Fee")
	a, b := original.CriteriaFlow, changed.CriteriaFlow
	if a.Predicates[0].GraphSHA256 == b.Predicates[0].GraphSHA256 ||
		a.Predicates[0].GeneratedSHA256 == b.Predicates[0].GeneratedSHA256 || a.CandidatePlanSHA256 != b.CandidatePlanSHA256 ||
		original.OriginalSourceSHA256 == changed.OriginalSourceSHA256 {
		t.Fatal("requirement change lost its source binding")
	}
	finite := strings.NewReplacer(`input "[11]"`, `input "[876543219]"`,
		`holdout_input "[20]"`, `holdout_input "[987654321]"`,
		`criteria_input "[17]"`, `criteria_input "[765432198]"`).Replace(scalarCriteriaSource)
	c := exportCriteriaFlow(t, finite, "Fee")
	ca, _ := json.Marshal(a)
	cc, _ := json.Marshal(c.CriteriaFlow)
	if string(ca) != string(cc) || original.ContractSHA256 == c.ContractSHA256 {
		t.Fatal("finite inputs changed source graph")
	}
	for _, secret := range []string{"876543219", "987654321", "765432198"} {
		if strings.Contains(string(cc), secret) {
			t.Fatal("finite input leaked", secret)
		}
	}
}

func TestCriteriaSourceFlowNormalizesLocalNamesAndRetainsGuardedHelpers(t *testing.T) {
	body := "let amount = Delta(input0); if input0 < 0 { return input1 == amount }; return input1 == input0 - 11"
	source := strings.Replace(scalarCriteriaSource, "return input1 == input0 - 11", body, 1) +
		`activity Delta(Integer) -> Integer id "fee-criteria://delta" computes "return input - 11"`
	a := exportCriteriaFlow(t, source, "Fee").CriteriaFlow.Predicates[0]
	b := exportCriteriaFlow(t, strings.ReplaceAll(source, "amount", "calculatedDifference"), "Fee").CriteriaFlow.Predicates[0]
	if a.GraphSHA256 != b.GraphSHA256 || a.ValueFlow.BodySHA256 == b.ValueFlow.BodySHA256 {
		t.Fatal("local spelling changed structural graph")
	}
	if len(a.ValueFlow.Helpers) != 1 || a.ValueFlow.Helpers[0].ActivityID != "fee-criteria://delta" {
		t.Fatal("helper identity missing")
	}
	calls, guardedReturns := 0, 0
	for _, n := range a.ValueFlow.Nodes {
		if n.Kind == "call" && n.CalleeID == "fee-criteria://delta" {
			calls++
		}
		if n.Kind == "return" && n.Guard != 0 {
			guardedReturns++
		}
	}
	if calls == 0 || guardedReturns != 2 {
		t.Fatal("call/return conditions missing", calls, guardedReturns, a.ValueFlow.Nodes)
	}
	changed := exportCriteriaFlow(t, strings.Replace(source, "if input0 < 0", "if input0 > 0", 1), "Fee").CriteriaFlow.Predicates[0]
	if changed.GraphSHA256 == a.GraphSHA256 {
		t.Fatal("branch condition missing from graph")
	}
	changed = exportCriteriaFlow(t, strings.Replace(source, `computes "return input - 11"`, `computes "return 11 - input"`, 1), "Fee").CriteriaFlow.Predicates[0]
	if changed.GraphSHA256 == a.GraphSHA256 || changed.GeneratedSHA256 == a.GeneratedSHA256 {
		t.Fatal("helper body change missing")
	}
}

func TestCriteriaSourceFlowSupportsRecordResultAndRemainsOptional(t *testing.T) {
	r := exportCriteriaFlow(t, typedCriteriaSource, "Select")
	f := r.CriteriaFlow
	p := f.Predicates[0]
	if f.CandidatePlan != nil || r.ValueFlow == nil || p.Inputs[1].Role != "candidate_result" || p.Inputs[1].Type != "Box" {
		t.Fatal("record source profile changed", p.Inputs)
	}
	field := false
	for _, n := range p.Nodes {
		field = field || n.Kind == "input" && n.Input == 2 && n.FieldID == "criteria://value"
	}
	if !field {
		t.Fatal("result field identity missing", p.Nodes)
	}
	plain, err := ExportRecordAssemblyContext(context.Background(), "c.gooo", []byte(typedCriteriaSource), "Select", false)
	if err != nil || plain.CriteriaFlow != nil {
		t.Fatal("optional source graph was enabled by default", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = ExportRecordAssemblyContextWithFlow(canceled, "c.gooo", []byte(scalarCriteriaSource), "Fee", false, "")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost", err)
	}
	f.CandidatePlan = nil
	p.Nodes[0].Kind = "mutated"
	fresh := exportCriteriaFlow(t, typedCriteriaSource, "Select")
	if fresh.CriteriaFlow.Predicates[0].Nodes[0].Kind == "mutated" {
		t.Fatal("export shares mutable graph storage")
	}
}

func TestCriteriaSourceFlowDeclinesBoundWithoutPartialCanonicalGraph(t *testing.T) {
	body := "let value = input0; " + strings.Repeat("value = value + 1; ", 180) + "return input1 == value"
	source := strings.Replace(scalarCriteriaSource, "return input1 == input0 - 11", body, 1)
	r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "bound.gooo", []byte(source), "Fee", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if r.CriteriaFlow == nil || r.CriteriaFlow.Status != "UNRESOLVED" || len(r.CriteriaFlow.Predicates) != 1 {
		t.Fatal("bound did not decline", r.CriteriaFlow)
	}
	p := r.CriteriaFlow.Predicates[0]
	if p.ValueFlow.Reason != "FLOW_NODE_BOUND" || len(p.Nodes) != 0 || p.GraphSHA256 != "" || len(p.ValueFlow.Nodes) != 0 {
		t.Fatal("bounded partial graph exposed as usable", p)
	}
	if r.ModelPredictions != 0 || r.CandidateTests != 0 {
		t.Fatal("optional graph evaluated candidates")
	}
}

func TestCriteriaSourceFlowUnsupportedLiteralDoesNotExposePartialGraph(t *testing.T) {
	source := strings.Replace(scalarCriteriaSource, "input1 == input0 - 11", "input1 == 'a'", 1)
	prepared := preparedRecordCriterion{declaration: assemblyspec.Criterion{ID: "difference", Activity: "Difference"}}
	p := criterionSourceFlow(context.Background(), "unsupported.gooo", []byte(source), prepared)
	if p.ValueFlow.Status != "UNRESOLVED" || p.ValueFlow.Reason != "source literal is outside the graph model profile" ||
		len(p.ValueFlow.Nodes) != 0 || len(p.Nodes) != 0 || p.GraphSHA256 != "" {
		t.Fatal("canonicalization decline retained a partial usable graph", p.ValueFlow, p.Nodes)
	}
}
