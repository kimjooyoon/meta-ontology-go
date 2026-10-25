package sourcebundle

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestLowerRetainsPublishedPolicyOriginsAcrossMembers(t *testing.T) {
	raw, err := os.ReadFile("../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	const header = "package metapolicycompilation\nnamespace metapolicycompilation\n"
	const policyID = "gooo://meta-policy-compilation/policy/v4"
	sources := []Source{{"entity.gooo", header + "entity Input id \"urn:example:input\"\n"},
		{" rules.gooo", string(raw)}}
	before, err := Lower(sources)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.IR.Policies) != 1 || len(before.Provenance.Declarations) != 2 {
		t.Fatalf("policy admitted without its physical origin: policies=%d origins=%d",
			len(before.IR.Policies), len(before.Provenance.Declarations))
	}
	assertPolicyOrigin(t, before, policyID, sources[1])
	sources[1].Filename = "moved.gooo"
	after, err := Lower(sources)
	if err != nil || after.SourceDigest == before.SourceDigest ||
		after.IR.SemanticCanonical() != before.IR.SemanticCanonical() {
		t.Fatalf("moving the policy changed its meaning or retained source identity: %v", err)
	}
	assertPolicyOrigin(t, after, policyID, sources[1])
	sources = append(sources, Source{"duplicate.gooo", string(raw)})
	if got, err := Lower(sources); err == nil || !reflect.DeepEqual(got, Bundle{}) {
		t.Fatal("duplicate policy returned a partial source bundle")
	}
}

func assertPolicyOrigin(t *testing.T, bundle Bundle, id string, source Source) {
	t.Helper()
	found := 0
	for _, origin := range bundle.Provenance.Declarations {
		if origin.ID != id {
			continue
		}
		found++
		span := origin.Span
		if span.File != source.Filename || !span.Valid() || span.Start < 0 ||
			span.End > len(source.Content) || span.Start >= span.End ||
			!strings.HasPrefix(source.Content[span.Start:span.End], "policy DecisionReduction ") ||
			span.StartLine != strings.Count(source.Content[:span.Start], "\n")+1 {
			t.Fatalf("policy origin no longer locates its physical declaration: %+v", origin)
		}
	}
	if found != 1 {
		t.Fatalf("want one origin for policy %q, got %d", id, found)
	}
}
