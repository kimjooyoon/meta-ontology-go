package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const bundleHeader = "package bundle\nnamespace bundle\n"

func bodyBundleFixture() []sourcebundle.Source {
	return []sourcebundle.Source{
		{Filename: "types.gooo", Content: bundleHeader + "entity Integer id \"urn:gooo:type:integer\"\n"},
		{Filename: "helpers.gooo", Content: bundleHeader + "# 계산의 기원\n" +
			"activity Twice(Integer) -> Integer id \"urn:bundle:twice\" computes `return input * 2`\n"},
		{Filename: "root.gooo", Content: bundleHeader +
			"activity Score(Integer) -> Integer id \"urn:bundle:score\" computes `if input < 0 { return Twice(-input) }; return Twice(input) + 3`\n"},
	}
}

func TestSourceBundleKeepsCrossFileCallsAndPhysicalOrigins(t *testing.T) {
	sources := bodyBundleFixture()
	sources[1].Content = strings.ReplaceAll(sources[1].Content, "\n", "\r\n")
	result, err := GenerateSourceBundle(sources, "Score")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Provenance.Members) != 3 || len(result.Provenance.Declarations) != 3 ||
		result.Report.CallClosure == nil || len(result.Report.CallClosure.Activities) != 1 {
		t.Fatal("source set or fixed callee lost", result)
	}
	for i, wantID := range []string{"urn:gooo:type:integer", "urn:bundle:twice", "urn:bundle:score"} {
		origin, member := result.Provenance.Declarations[i], result.Provenance.Members[i]
		content := sources[i].Content
		wantStart := strings.Index(content, "entity ")
		if i != 0 {
			wantStart = strings.Index(content, "activity ")
		}
		if origin.ID != wantID || origin.Span.File != sources[i].Filename || origin.Span.Start != wantStart ||
			origin.Span.End > len(content) || origin.Span.End <= wantStart ||
			!strings.Contains(content[wantStart:origin.Span.End], wantID) ||
			member.Filename != origin.Span.File || member.ByteLength != len(content) ||
			"sha256:"+member.SHA256 != digest([]byte(content)) {
			t.Fatal("origin not bound to the physical member", origin, member)
		}
	}
	cases := []IRBodyFillTestCase{{Input: -5, Expected: 10}, {Input: 0, Expected: 3},
		{Input: 7, Expected: 17}, {Input: 9007199254740993, Expected: 18014398509481989}}
	_, passed, err := evaluateIntegerCasesContext(context.Background(), []byte(result.Source), "Score", cases)
	if err != nil || passed != len(cases) {
		t.Fatal("separately authored call outcomes differ", passed, err)
	}
	changed := bodyBundleFixture()
	changed[1].Content = strings.Replace(changed[1].Content, "input * 2", "input * 3", 1)
	next, err := GenerateSourceBundle(changed, "Score")
	if err != nil || next.Report.SourceDigest == result.Report.SourceDigest ||
		next.Report.RouteEquivalence.SourceSemanticDigest == result.Report.RouteEquivalence.SourceSemanticDigest {
		t.Fatal("helper edit was omitted from the selected body's dependencies", err)
	}
}

func TestSourceBundleRelocationChangesOriginNotProjection(t *testing.T) {
	sources := bodyBundleFixture()
	var joined strings.Builder
	joined.WriteString(bundleHeader)
	for _, source := range sources {
		joined.WriteString(strings.TrimPrefix(source.Content, bundleHeader))
	}
	single, err := Generate("old.gooo", []byte(joined.String()), "Score")
	if err != nil {
		t.Fatal(err)
	}
	var previousDigest string
	for step := range 4 {
		result, err := GenerateSourceBundle(sources, "Score")
		if err != nil || result.Source != single.Source || result.Report.ActivityID != single.Report.ActivityID ||
			result.Report.SourceDigest == single.Report.SourceDigest || result.Report.SourceDigest == previousDigest ||
			!result.Report.DeterministicReplay || result.Report.RepositoryWrites != 0 {
			t.Fatal("source placement changed code or failed to change provenance", step, err, result)
		}
		previousDigest = result.Report.SourceDigest
		switch step {
		case 0:
			sources[1].Filename = "moved/helper.gooo"
		case 1:
			sources[1].Content += "# observation only\n"
		case 2:
			sources[0], sources[2] = sources[2], sources[0]
		}
	}
}

func TestSourceBundleRetainsRequiredAndOptionalFieldProfiles(t *testing.T) {
	for _, presence := range []string{"required", "optional"} {
		t.Run(presence, func(t *testing.T) {
			entity := fmt.Sprintf("entity Box id \"urn:bundle:box\" fields { field value id \"urn:bundle:value\" type integer %s one }\n", presence)
			activity := "activity Echo(Box) -> Box id \"urn:bundle:echo\" computes `return input`\n"
			sources := []sourcebundle.Source{{Filename: "types.gooo", Content: bundleHeader + entity},
				{Filename: "body.gooo", Content: bundleHeader + activity}}
			parsed, err := parseBodySources(sources)
			if err != nil {
				t.Fatal(err)
			}
			want := syntax.EntityFieldsV3Support()
			if presence == "optional" {
				want = syntax.EntityFieldsV4Support()
			}
			if bodyEntityFieldsSupport(parsed.File) != want || parsed.File.Span.Filename != "" {
				t.Fatal("profile changed or virtual file invented", parsed)
			}
			single, err := Generate("one.gooo", []byte(bundleHeader+entity+activity), "Echo")
			if err != nil {
				t.Fatal(err)
			}
			bundle, err := GenerateSourceBundle(sources, "Echo")
			if err != nil || bundle.Source != single.Source || !reflect.DeepEqual(bundle.Report.RecordTypes, single.Report.RecordTypes) {
				t.Fatal("partition changed typed record projection", err)
			}
		})
	}
}

func TestSourceBundleRejectsPartialOrConflictingPrograms(t *testing.T) {
	for _, name := range []string{"missing-helper", "duplicate-id", "duplicate-name", "wrong-package", "import", "syntax", "cycle", "unknown-type", "alias-name"} {
		t.Run(name, func(t *testing.T) {
			sources := bodyBundleFixture()
			switch name {
			case "missing-helper":
				sources = append(sources[:1], sources[2])
			case "duplicate-id":
				sources[2].Content = strings.Replace(sources[2].Content, "urn:bundle:score", "urn:bundle:twice", 1)
			case "duplicate-name":
				sources[2].Content = strings.Replace(sources[2].Content, "activity Score", "activity Twice", 1)
			case "wrong-package":
				sources[1].Content = strings.Replace(sources[1].Content, "package bundle", "package elsewhere", 1)
			case "import":
				sources[1].Content = bundleHeader + "import \"other\"\n" + strings.TrimPrefix(sources[1].Content, bundleHeader)
			case "syntax":
				sources[1].Content += "?\n"
			case "cycle":
				sources[1].Content = strings.Replace(sources[1].Content, "input * 2", "Score(input)", 1)
			case "unknown-type":
				sources[1].Content = strings.Replace(sources[1].Content, "Twice(Integer)", "Twice(Missing)", 1)
			case "alias-name":
				sources[1].Filename = "TYPES.gooo"
			}
			result, err := GenerateSourceBundle(sources, "Score")
			if err == nil || !reflect.DeepEqual(result, SourceBundleResult{}) {
				t.Fatal("partial code admitted", name, err, result)
			}
			if name == "syntax" && !strings.Contains(err.Error(), "helpers.gooo") {
				t.Fatal("diagnostic lost member address", err)
			}
		})
	}
}

func TestSourceBundleRequiresFixedEntryAndCallees(t *testing.T) {
	sources := []sourcebundle.Source{{Filename: "choice.gooo", Content: sourceRecordObservationFixture}}
	result, err := GenerateSourceBundle(sources, "Decide")
	if err == nil || !strings.Contains(err.Error(), "fixed computes") || !reflect.DeepEqual(result, SourceBundleResult{}) {
		t.Fatal("unresolved assembly accepted as a fixed entry", err)
	}
	// An unrelated assembling declaration does not prevent a fixed entry from
	// using its own typed body. Selection is scoped to the entry's call closure.
	if _, err := GenerateSourceBundle(sources, "Expected"); err != nil {
		t.Fatal("unrelated draft blocked a fixed body", err)
	}
	sources = append(sources, sourcebundle.Source{Filename: "invoke.gooo", Content: "package refunds\nnamespace refunds\nactivity Invoke(Order,Boolean) -> Order computes `return Decide(input0,input1)`\n"})
	result, err = GenerateSourceBundle(sources, "Invoke")
	if err == nil || !reflect.DeepEqual(result, SourceBundleResult{}) {
		t.Fatal("unresolved assembly accepted through a callee", err)
	}
}
