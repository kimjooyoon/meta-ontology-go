package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sourceObservationFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func sourceObservationInput(t *testing.T, source []byte) sourceObservationRequest {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.gooo")
	if err := os.WriteFile(path, source, 0400); err != nil {
		t.Fatal(err)
	}
	return sourceObservationRequest{
		PolicyPath: path, ExpectedPackage: "metapolicycompilation",
		ExpectedNamespace: "metapolicycompilation", Flags: []string{"observe-source", "policy"},
	}
}

func requireSourceObservation(t *testing.T, source []byte, namespace string) policySourceObservation {
	t.Helper()
	observed, err := reconstructSourceObservation("policy.gooo", source, "metapolicycompilation", namespace)
	if err != nil {
		t.Fatal(err)
	}
	return observed
}

func TestSourceObservationBindsRawBytesAndSourceRules(t *testing.T) {
	source := sourceObservationFixture(t)
	base := requireSourceObservation(t, source, "metapolicycompilation")
	t.Run("raw-source-and-rule-bindings", func(t *testing.T) {
		if base.Schema != sourceObservationSchema || base.PolicyID == "" ||
			base.SourceDigest != digestBytes(source) || !strings.HasPrefix(base.SemanticDigest, "sha256:") ||
			base.Denominator != fixedDenom || len(base.RuleBindings) != fixedDenom || len(base.DecisionRules) != fixedDenom ||
			!base.SourceParseObserved || !base.LoweringObserved {
			t.Fatal("source observation is detached from its raw input or source rules")
		}
		for _, binding := range base.RuleBindings {
			if binding.ActivityID == "" || binding.MetaOperation == "" || binding.ProofChoice == "" {
				t.Fatal("source rule lost its meta-operation binding")
			}
		}
	})
	t.Run("comment-is-not-a-new-semantic-identity", func(t *testing.T) {
		commented := append(bytes.Clone(source), []byte("\n// source observation binding\n")...)
		observed := requireSourceObservation(t, commented, "metapolicycompilation")
		if observed.SourceDigest == base.SourceDigest || observed.SemanticDigest != base.SemanticDigest {
			t.Fatal("raw source and semantic identity were conflated")
		}
	})
	t.Run("namespace-is-not-a-cached-semantic-identity", func(t *testing.T) {
		old := []byte("namespace metapolicycompilation")
		if !bytes.Contains(source, old) {
			t.Fatal("fixture namespace anchor is missing")
		}
		changed := bytes.Replace(source, old, []byte("namespace policyobservation"), 1)
		observed := requireSourceObservation(t, changed, "policyobservation")
		if observed.Namespace != "policyobservation" || observed.SourceDigest == base.SourceDigest ||
			observed.SemanticDigest == base.SemanticDigest {
			t.Fatal("source observation reused a different source identity")
		}
	})
}

func TestSourceObservationRejectsInvalidSourceAndIdentity(t *testing.T) {
	source := sourceObservationFixture(t)
	for _, test := range []struct {
		name      string
		source    []byte
		pkg       string
		namespace string
	}{
		{"empty-source", nil, "metapolicycompilation", "metapolicycompilation"},
		{"invalid-syntax", []byte("not a Gooo policy"), "metapolicycompilation", "metapolicycompilation"},
		{"wrong-package", source, "other", "metapolicycompilation"},
		{"wrong-namespace", source, "metapolicycompilation", "other"},
	} {
		t.Run(test.name, func(t *testing.T) {
			observed, err := reconstructSourceObservation("policy.gooo", test.source, test.pkg, test.namespace)
			if err == nil || observed.Schema != "" || observed.SourceDigest != "" || len(observed.RuleBindings) != 0 {
				t.Fatal("unobserved source produced an observation")
			}
		})
	}
}

func TestSourceObservationRejectsMixedModesBeforeFileObservation(t *testing.T) {
	for _, name := range []string{"cases", "artifact", "manifest", "output", "unexpected-option",
		"positional", "missing-policy", "missing-package", "missing-namespace"} {
		t.Run(name, func(t *testing.T) {
			request := sourceObservationRequest{
				PolicyPath:      filepath.Join(t.TempDir(), "must-not-be-read.gooo"),
				ExpectedPackage: "metapolicycompilation", ExpectedNamespace: "metapolicycompilation",
				Flags: []string{"observe-source", "policy"},
			}
			want := "does not accept flag"
			switch name {
			case "positional":
				request.Arguments, want = []string{"ignored.gooo"}, "does not accept positional"
			case "missing-policy":
				request.PolicyPath, want = "", "requires -policy"
			case "missing-package":
				request.ExpectedPackage, want = "", "requires an expected"
			case "missing-namespace":
				request.ExpectedNamespace, want = "", "requires an expected"
			default:
				request.Flags = append(request.Flags, name)
			}
			var output bytes.Buffer
			err := runSourceObservation(request, &output)
			if err == nil || !strings.Contains(err.Error(), want) || output.Len() != 0 {
				t.Fatalf("mixed mode reached source observation: %v", err)
			}
		})
	}
}

func TestSourceObservationEmitsOnlyReadOnlyUnknownConformance(t *testing.T) {
	source := sourceObservationFixture(t)
	request := sourceObservationInput(t, source)
	var output bytes.Buffer
	if err := runSourceObservation(request, &output); err != nil {
		t.Fatal(err)
	}
	var observed policySourceObservation
	if err := decodeRequiredDocument(output.Bytes(), &observed, []string{
		"schema", "policy_id", "package", "namespace", "source_digest", "semantic_digest",
		"fixed_denominator", "rule_bindings", "decision_rules", "source_parse_observed", "lowering_observed",
		"policy_execution_observed", "producer_artifacts_observed", "current_conformance", "conformance_unknown",
		"repository_writes", "mutation_authority", "promotion_authority",
	}); err != nil {
		t.Fatal(err)
	}
	unknown := observed.ConformanceUnknown
	if observed.PolicyExecutionObserved || observed.ProducerArtifactsObserved ||
		observed.RepositoryWrites != 0 || observed.MutationAuthority != 0 || observed.PromotionAuthority != 0 ||
		observed.CurrentConformance != decisionUnknown || unknown.State != decisionUnknown ||
		unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" ||
		unknown.UnknownClass != "DIRECT_MISSING" || unknown.NextOperation == "" ||
		unknown.BlockedBy == nil || len(unknown.BlockedBy) != 0 {
		t.Fatal("raw reconstruction was promoted into execution, conformance, or authority")
	}
	after, err := os.ReadFile(request.PolicyPath)
	if err != nil || !bytes.Equal(after, source) {
		t.Fatalf("source observation changed its read-only input: %v", err)
	}
}

type sourceObservationClosedWriter struct{}

func (sourceObservationClosedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestSourceObservationPreservesReadAndOutputFailures(t *testing.T) {
	t.Run("missing-input", func(t *testing.T) {
		request := sourceObservationRequest{
			PolicyPath:      filepath.Join(t.TempDir(), "missing.gooo"),
			ExpectedPackage: "metapolicycompilation", ExpectedNamespace: "metapolicycompilation",
		}
		var output bytes.Buffer
		if err := runSourceObservation(request, &output); err == nil || output.Len() != 0 {
			t.Fatal("missing input produced a successful observation")
		}
	})
	t.Run("output-unavailable", func(t *testing.T) {
		request := sourceObservationInput(t, sourceObservationFixture(t))
		if err := runSourceObservation(request, sourceObservationClosedWriter{}); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("source observation hid its output failure: %v", err)
		}
	})
}
