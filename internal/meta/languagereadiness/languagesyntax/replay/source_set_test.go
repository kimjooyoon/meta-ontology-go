package replay

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcepartition"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func sourceSetFixture(t *testing.T) (fstest.MapFS, []string, string) {
	t.Helper()
	source := "// 기원\r\npackage p\r\nnamespace p\r\nentity Value id \"urn:value\"\r\n" +
		"activity Read(Value) -> Value id \"urn:read\" computes \"return input\"\r\n" +
		"activity Deliver(Value) -> Value id \"urn:deliver\" computes \"return input\"\r\n" +
		"bind Read.result -> Deliver.input\r\n// keep last comment"
	files, err := sourcepartition.SplitAt("first.gooo", source, "second.gooo", strings.Index(source, "activity Deliver"))
	if err != nil {
		t.Fatal(err)
	}
	repository := fstest.MapFS{}
	for _, file := range files {
		repository[file.Filename] = &fstest.MapFile{Data: []byte(file.Source)}
	}
	return repository, []string{"first.gooo", "second.gooo"}, source
}

func TestSourceSetPreservesCrossMemberMeaningAndPhysicalOrigins(t *testing.T) {
	files, names, original := sourceSetFixture(t)
	whole := Execute(fstest.MapFS{"whole.gooo": &fstest.MapFile{Data: []byte(original)}}, "whole.gooo", "VALID", "")
	result := ExecuteSourceSet(files, names, syntax.CurrentEntityFieldsSupport(), false)
	if result.ObservedDecision != DecisionPass || whole.ObservedDecision != DecisionPass ||
		result.SemanticDigest != whole.SemanticDigest || result.SourceDigestKind != "gooo/source-bundle/v1" ||
		len(result.SourceMembers) != 2 || !result.GetPut || !result.PutGet {
		t.Fatal("source family failed joint replay", result)
	}
	if fragment := Execute(files, "second.gooo", "VALID", ""); fragment.ObservedDecision == DecisionPass {
		t.Fatal("the dependent fragment must not pass as a standalone program")
	}
	for i, name := range names {
		if result.SourceMembers[i] != ObserveSource(name, files[name].Data) {
			t.Fatal("physical member attribution lost", result.SourceMembers)
		}
	}
	if replayed := ExecuteSourceSet(files, names, syntax.CurrentEntityFieldsSupport(), false); !reflect.DeepEqual(result, replayed) {
		t.Fatal("source set replay is not deterministic")
	}
	files["second.gooo"].Data = append(files["second.gooo"].Data, []byte("\n// current origin")...)
	changed := ExecuteSourceSet(files, names, syntax.CurrentEntityFieldsSupport(), false)
	if changed.ObservedDecision != DecisionPass || result.SourceDigest == changed.SourceDigest ||
		result.SemanticDigest != changed.SemanticDigest || result.SourceMembers[0] != changed.SourceMembers[0] ||
		result.SourceMembers[1] == changed.SourceMembers[1] {
		t.Fatal("second-member source change was not bound independently", changed)
	}
}

func TestSourceSetRefusesIncompleteInvalidAndConflictingMembers(t *testing.T) {
	for _, mode := range []string{"missing", "duplicate", "case-alias", "path", "syntax", "header", "identity", "limit"} {
		files, names, _ := sourceSetFixture(t)
		want := DecisionUnknown
		switch mode {
		case "missing":
			delete(files, "second.gooo")
		case "duplicate":
			names[1] = names[0]
		case "case-alias":
			names[1] = "FIRST.gooo"
		case "path":
			names[1] = "../second.gooo"
		case "limit":
			names = make([]string, 65)
		default:
			want = DecisionClosed
			source := string(files["second.gooo"].Data)
			switch mode {
			case "syntax":
				source += "\nbroken"
			case "header":
				source = strings.Replace(source, "namespace p", "namespace other", 1)
			case "identity":
				source = strings.Replace(source, "urn:deliver", "urn:read", 1)
			}
			files["second.gooo"].Data = []byte(source)
		}
		result := ExecuteSourceSet(files, names, syntax.CurrentEntityFieldsSupport(), false)
		if result.ObservedDecision != want || (want == DecisionUnknown && (len(result.SourceMembers) != 0 || result.SourceDigest != "")) {
			t.Fatal("invalid source family was accepted or leaked an input prefix", mode, result)
		}
	}
}
