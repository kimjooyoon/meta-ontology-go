package policy

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestContinuationDependenciesFresh1619(t *testing.T) {
	for _, c := range []struct{ reason, dependency, next string }{
		{"SOURCE_RUN_METADATA_UNKNOWN", "producer_run", "observe_exact_source_run"},
		{"SOURCE_GRANT_JOB_UNAVAILABLE", "producer_run", "observe_exact_source_run"},
		{"SUBJECT_SOURCE_UNKNOWN", "subject_blob", "observe_exact_subject_blob"},
		{"GRANT_POLICY_UNAVAILABLE", "policy_source", "restore_policy_source"},
		{"SOURCE_ARTIFACT_EXPIRED", "expired_source_artifact", "produce_fresh_source_artifact"},
		{"SOURCE_MEMBER_MISSING", "source_artifact", "observe_exact_source_artifact"},
		{"DISPATCH_RESPONSE_UNKNOWN", "dispatch_outcome", "observe_dispatch_outcome"},
		{"TARGET_REFERENCE_UNKNOWN", "target_run_reference", "observe_dispatch_response"},
		{"RESERVATION_RECORD_UNBOUND", "reservation_history", "restore_retained_reservation_history"},
		{"TARGET_RUN_METADATA_UNKNOWN", "target_run", "observe_returned_target_run"},
	} {
		t.Run(c.reason, func(t *testing.T) {
			dependency := MissingDependency(c.reason)
			if dependency != c.dependency || NextObservation(dependency) != c.next {
				t.Fatal("causal continuation", c, dependency, NextObservation(dependency))
			}
		})
	}
}

func TestDependencyProjectionFresh1619(t *testing.T) {
	for _, c := range []struct{ activity, file string }{
		{"MissingDependency", "dependency.generated.go"}, {"NextObservation", "next.generated.go"},
	} {
		t.Run(c.activity, func(t *testing.T) {
			members, err := policySources()
			if err != nil {
				t.Fatal(err)
			}
			projection, err := bodycodegen.GenerateSourceBundle(members, c.activity)
			if err != nil {
				t.Fatal(err)
			}
			committed, err := os.ReadFile(c.file)
			if err != nil || string(committed) != projection.Source {
				t.Fatal("source projection differs", err)
			}
		})
	}
}
