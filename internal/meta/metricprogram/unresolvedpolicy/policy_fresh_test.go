package unresolvedpolicy

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func TestUnresolvedClassificationFresh1721(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		registered, recorded     bool
		resolution, reason, next string
	}{
		{"registered-origin", true, true, "exact_operation", "REGISTERED_SOURCE_OPERATION", "CONTINUE_SOURCE_OPERATION"},
		{"unregistered-origin", false, true, "invariant_only", "OPERATION_NOT_IN_SOURCE_REGISTRY", "DECLARE_SOURCE_OPERATION"},
		{"registered-missing-origin", true, false, "invariant_only", "OPERATION_ORIGIN_REQUIRED", "CAPTURE_OPERATION_ORIGIN"},
		{"unregistered-missing-origin", false, false, "invariant_only", "OPERATION_ORIGIN_REQUIRED", "CAPTURE_OPERATION_ORIGIN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if OperationResolution(tc.registered, tc.recorded) != tc.resolution ||
				OperationReason(tc.registered, tc.recorded) != tc.reason ||
				OperationNextStep(tc.registered, tc.recorded) != tc.next {
				t.Fatal("source policy lost a context distinction")
			}
		})
	}
}

func TestUnresolvedOriginTransportFresh1725(t *testing.T) {
	for _, recorded := range []bool{true, false} {
		t.Run(fmt.Sprint(recorded), func(t *testing.T) {
			origin := OperationOrigin{Repository: "example/compiler", SubjectSHA: "original-head",
				IndicatorID: "upstream://indicator/new", RequestedOperation: "upstream://operation/new", Registry: Origin()}
			origin.Inputs = InputDigests{StrategyPayloadDigest: projectionbundle.Digest([]byte("strategy")),
				VerificationPayloadDigest: projectionbundle.Digest([]byte("verification"))}
			if !recorded {
				origin.Registry.SourceDigest = ""
			}
			r, err := NewOperationRefusal(origin)
			if err != nil {
				t.Fatal(err)
			}
			var captured *OperationRefusal
			if !errors.As(fmt.Errorf("compile source: %w", r), &captured) || captured.Origin != origin {
				t.Fatal("wrapped refusal lost the original operation request")
			}
			raw, err := json.Marshal(captured)
			if err != nil {
				t.Fatal(err)
			}
			var decoded OperationRefusal
			if err := json.Unmarshal(raw, &decoded); err != nil || decoded != *captured ||
				decoded.OriginRecorded != recorded || decoded.Resolution != "invariant_only" ||
				decoded.ReadinessCredit || decoded.RepositoryWorkspaceWrites || decoded.PromotionAuthorized {
				t.Fatalf("refusal transport changed origin or authority: %s, %v", raw, err)
			}
		})
	}
}
