package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicorchestration"
)

func TestUpstreamLabelsCannotAuthorizePartialReuse(t *testing.T) {
	digest := cache.HashBytes([]byte("test-only-header-not-an-authorization")).String()
	bounded := upstreamReport{
		Schema: publicorchestration.ReportSchema, Operation: publicorchestration.Operation,
		Decision: "CLOSED", CaseID: publicorchestration.CaseAuthorizedOrchestration,
		PolicySourceDigest: digest, PolicySemanticDigest: digest, PolicyEvaluatorDigest: digest,
		HandoffDigest: digest, AuthorizationDigest: digest, CertificateDigest: digest, ReceiptDigest: digest,
		StatePath: []string{"AUTHORIZE", "EVIDENCE"},
	}
	labels := upstreamReport{Schema: bounded.Schema, Operation: bounded.Operation, Decision: "CLOSED"}
	writes, contradictory, fixedPoint := bounded, bounded, bounded
	writes.RepositoryWrites = 1
	contradictory.Unknown = &publicorchestration.UnknownState{Reason: "NOT_CLOSED"}
	fixedPoint.Decision = "FIXED_POINT"
	for _, item := range []struct {
		name     string
		report   upstreamReport
		decision string
		reason   string
	}{
		{"labels_only", labels, "UNKNOWN", "UPSTREAM_BINDING_MISSING"},
		{"self_asserted_bound_header", bounded, "UNKNOWN", "UPSTREAM_EVIDENCE_REQUIRED"},
		{"repository_write_contradiction", writes, "REFUTED", "UPSTREAM_REPORT_CONTRADICTS_AUTHORIZED_BOUNDARY"},
		{"closed_with_unknown", contradictory, "REFUTED", "UPSTREAM_REPORT_CONTRADICTS_AUTHORIZED_BOUNDARY"},
		{"fixed_point_is_not_authorization", fixedPoint, "REFUTED", "UPSTREAM_REPORT_CONTRADICTS_AUTHORIZED_BOUNDARY"},
	} {
		t.Run(item.name, func(t *testing.T) {
			root := t.TempDir()
			filename := filepath.Join(root, "report.json")
			data, err := json.Marshal(item.report)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filename, data, 0o444); err != nil {
				t.Fatal(err)
			}
			_, _, err = verifyUpstream(runInput{OrchestrationReport: filename, Out: filepath.Join(root, "output")})
			var failure *upstreamFailure
			if !errors.As(err, &failure) || failure.Decision != item.decision || failure.Reason != item.reason {
				t.Fatalf("admission error=%v, want %s/%s", err, item.decision, item.reason)
			}
			if item.decision == "UNKNOWN" && (failure.Unknown == nil || failure.Unknown.Stage == "" || failure.Unknown.Step == "" || failure.Unknown.Reason == "" || failure.Unknown.UnknownClass == "" || failure.Unknown.NextOperation == "" || failure.Unknown.BlockedBy == nil) {
				t.Fatal("UNKNOWN admission lost its six-field cause")
			}
			if _, err := os.Stat(filepath.Join(root, "output")); !os.IsNotExist(err) {
				t.Fatal("unverified header reached the evidence or execution stage")
			}
		})
	}
}
