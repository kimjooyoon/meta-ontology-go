package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"
)

func recordUpdatesFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/record-field-updates.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestRecordUpdatesAssembleSequentialDependenciesAndReplay(t *testing.T) {
	source := recordUpdatesFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "updates.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.FieldsPassed != 15 || r.FieldsTotal != 15 || r.SelectedMask != 7 || len(r.Attempts) != 8 || r.ModelCalls != 0 {
		t.Fatalf("sequential finite selection: %+v", r)
	}
	for _, c := range r.Choices {
		if c.Kind != "field_update" || !strings.HasPrefix(c.FieldID, "fieldupdates://candidate/") {
			t.Fatal(c)
		}
	}
	if _, err = RealizeSourceAssembly(context.Background(), "updates.gooo", source, result); err != nil {
		t.Fatal(err)
	}
	next, err := g.GenerateSourceAssembly(context.Background(), "updates.gooo", []byte(result.GoooSource), "Select")
	if err != nil || next.Source != result.Source || next.GoooSource != result.GoooSource {
		t.Fatal("checkpoint", err)
	}
	partial := []byte(strings.Replace(string(source), "attempts \"8\"", "attempts \"2\"", 1))
	result, err = g.GenerateSourceAssembly(context.Background(), "partial.gooo", partial, "Select")
	if err != nil || result.Report.RecordAssembly.FieldsPassed != 9 || result.Report.RecordAssembly.Status != "PARTIAL_FINITE" {
		t.Fatal("partial", err)
	}
}

func TestRecordUpdateRejectsInvalidTargetsAndAlternatives(t *testing.T) {
	source := string(recordUpdatesFixture(t))
	for _, target := range []string{"input0.title", "missing.title", "input1.title", "copy.missing", "copy.title.missing", "(Candidate{title: \"a\", state: \"b\", reason: \"c\"}).title"} {
		bad := []byte(strings.Replace(source, "copy.title", target, 1))
		if _, err := Generate("invalid.gooo", bad, "Select"); err == nil {
			t.Fatalf("accepted target %s", target)
		}
	}
	for _, alternative := range []string{"input1", "missing.title", "copy.missing"} {
		bad := []byte(strings.Replace(source, "alternative \"input0.title\"", "alternative \""+alternative+"\"", 1))
		if err := ValidateSourceAssembly(context.Background(), "invalid.gooo", bad, "Select"); err == nil {
			t.Fatal(alternative)
		}
	}
}

func TestRecordConstructorAndUpdateOrdinalsRemainSeparate(t *testing.T) {
	source := string(recordAssemblyFixture(t))
	source = strings.Replace(source, "copy = Candidate{title: \"draft\", state: \"wait\", reason: \"deferred\"}", "copy = Candidate{title: \"draft\", state: \"wait\", reason: \"deferred\"}\ncopy.state = \"wait\"", 1)
	source = strings.Replace(source, "choice \"state\" field_value at \"1\"", "choice \"state\" field_update at \"0\"", 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "mixed.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.FieldsPassed != 15 || r.SelectedMask != 7 || r.Choices[0].Kind != "" || r.Choices[1].Kind != "field_update" || r.Choices[2].Field != "reason" {
		t.Fatal(r)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "mixed.gooo", []byte(source), result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordUpdateKeepsCopiedRecordsAndScalarSnapshots(t *testing.T) {
	source := string(recordUpdatesFixture(t))
	body := "let copy = input0\nlet saved = copy\nlet old = copy.state\ncopy.title = \"changed\"\n(copy).state = \"ready\"\ncopy.reason = saved.title + \":\" + old\nreturn copy"
	start, end := strings.Index(source, "computes `")+len("computes `"), strings.Index(source, "` assembling {")
	source = source[:start] + body + source[end:]
	result, err := Generate("snapshots.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := prepareRecordAssembly(context.Background(), "updates.gooo", recordUpdatesFixture(t), "Select")
	if err != nil {
		t.Fatal(err)
	}
	observed, err := evaluateRecordAssembly(context.Background(), []byte(result.Source), "Select", plan.body.records, plan.spec.ValueCases[:1])
	if err != nil || string(observed[0].Actual) != `{"reason":"한글:queued","state":"ready","title":"changed"}` {
		t.Fatal("value copy", err, observed)
	}
}
