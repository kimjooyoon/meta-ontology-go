package policycompilation

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestFirstClassPolicyOwnsASTIRAndEvaluation(t *testing.T) {
	source, err := os.ReadFile("../../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "computes") || strings.Contains(string(source), "policy-compilation:v3") {
		t.Fatal("canonical policy still contains an opaque marker program")
	}
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"policy": 1, "state": 11, "transition": 8, "case": 8, "evidence": 8, "resolution": 8}
	if policy.Structure.GrammarNodeKinds != 6 || policy.Structure.TransitionBindings != 8 || policy.Structure.EvidenceBindings != 8 || policy.Structure.ResolutionBindings != 8 {
		t.Fatalf("unexpected first-class bindings: %#v", policy.Structure)
	}
	for kind, count := range want {
		if policy.Structure.ASTNodeCounts[kind] != count || policy.Structure.IRNodeCounts[kind] != count {
			t.Fatalf("node count %s: AST=%d IR=%d want=%d", kind, policy.Structure.ASTNodeCounts[kind], policy.Structure.IRNodeCounts[kind], count)
		}
	}
	file, diagnostics := syntax.ParseFile("policy.gooo", string(source))
	if diagnostics.HasErrors() || len(file.Declarations) != 1 {
		t.Fatalf("first-class policy did not parse as one declaration: %v", diagnostics)
	}
	ir, err := bidir.Lower(file)
	if err != nil || len(ir.Policies) != 1 || len(ir.Policies[0].Cases) != 8 {
		t.Fatalf("first-class policy did not lower to semantic IR: policies=%d err=%v", len(ir.Policies), err)
	}
	judge := DigestBytes(GenerateJudge(policy))
	unknown := EvaluateSourcePolicy(policy, Case{ID: "unknown", ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: "sha256:not-a-valid-content-digest", ObservedArtifactSourceDigest: policy.SourceDigest, ObservedGeneratedJudgeDigest: judge, ObservedIndependentDigest: policy.SemanticDigest})
	if unknown.Decision != DecisionUnknown || unknown.Stage == "" || unknown.Step == 0 || unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" || len(unknown.BlockedBy) == 0 {
		t.Fatalf("UNKNOWN resolution lost its six fields: %#v", unknown)
	}
}
