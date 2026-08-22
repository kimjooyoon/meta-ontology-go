package artifact

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
)

func Validate(receipt Receipt) error {
	summary := receipt.Snapshot.Summary
	switch {
	case receipt.Schema != Schema:
		return fmt.Errorf("FAIL_CLOSED: readiness artifact schema %q", receipt.Schema)
	case !validHeadSHA(receipt.HeadSHA):
		return fmt.Errorf("FAIL_CLOSED: readiness artifact head sha is invalid")
	case receipt.Snapshot.Schema != readiness.SnapshotSchema:
		return fmt.Errorf("FAIL_CLOSED: readiness snapshot schema %q", receipt.Snapshot.Schema)
	case receipt.Snapshot.Decision != "PASS" || receipt.Snapshot.Reason != "READINESS_EXACTLY_COUNTED":
		return fmt.Errorf("FAIL_CLOSED: readiness snapshot decision %q", receipt.Snapshot.Decision)
	case summary.Total != int(improvement.SnapshotTotal) || len(receipt.Snapshot.Obligations) != summary.Total:
		return fmt.Errorf("FAIL_CLOSED: readiness denominator %d", summary.Total)
	case summary.Completed+summary.NotSatisfied+summary.Unresolved != summary.Total:
		return fmt.Errorf("FAIL_CLOSED: readiness counts do not close")
	case summary.Unresolved != 0 || receipt.Snapshot.RepositoryWrites != 0:
		return fmt.Errorf("FAIL_CLOSED: readiness guardrail unresolved=%d writes=%d",
			summary.Unresolved, receipt.Snapshot.RepositoryWrites)
	case summary.ReadinessBPS != summary.Completed*10_000/summary.Total:
		return fmt.Errorf("FAIL_CLOSED: readiness basis points %d", summary.ReadinessBPS)
	case receipt.Snapshot.SourceArtifactDigest == "":
		return fmt.Errorf("FAIL_CLOSED: source artifact digest is missing")
	case !validPromotionDigest(receipt):
		return fmt.Errorf("FAIL_CLOSED: proposal promotion digest is inconsistent")
	case receipt.Snapshot.Digest != snapshotDigest(receipt.Snapshot):
		return fmt.Errorf("FAIL_CLOSED: readiness snapshot digest mismatch")
	}
	expectedInput := improvement.FromReadiness(receipt.Snapshot)
	if !reflect.DeepEqual(receipt.TransitionInput, expectedInput) {
		return fmt.Errorf("FAIL_CLOSED: readiness transition input mismatch")
	}
	expectedFixedPoint := improvement.Evaluate(expectedInput, expectedInput)
	if !reflect.DeepEqual(receipt.FixedPoint, expectedFixedPoint) ||
		receipt.FixedPoint.Decision != improvement.NoChange {
		return fmt.Errorf("FAIL_CLOSED: readiness fixed point mismatch")
	}
	if receipt.ArtifactDigest != seal(receipt).ArtifactDigest {
		return fmt.Errorf("FAIL_CLOSED: readiness artifact digest mismatch")
	}
	return nil
}

func validHeadSHA(value string) bool {
	if len(value) != commitHexLength || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

const commitHexLength = 40

func validPromotionDigest(receipt Receipt) bool {
	required := false
	for _, result := range receipt.Snapshot.Obligations { required = required || (result.ID == "AUTONOMY-CHANGE-PROPOSAL" && result.Status == "SATISFIED") }
	if !required {
		return receipt.ProposalPromotionDigest == ""
	}
	value := strings.TrimPrefix(receipt.ProposalPromotionDigest, "sha256:")
	_, err := hex.DecodeString(value)
	return strings.HasPrefix(receipt.ProposalPromotionDigest, "sha256:") && len(value) == 64 && err == nil
}
