package artifact

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
)

func ValidateImprovement(
	beforeRaw []byte, before, after Receipt, receipt ImprovementArtifact,
) error {
	if err := ValidateFoundationBaseline(beforeRaw, before, receipt.Baseline); err != nil {
		return err
	}
	if err := Validate(after); err != nil {
		return fmt.Errorf("FAIL_CLOSED: current readiness invalid: %w", err)
	}
	switch {
	case receipt.Schema != ImprovementArtifactSchema:
		return fmt.Errorf("FAIL_CLOSED: improvement artifact schema mismatch")
	case receipt.Producer != improvementProducer ||
		receipt.Consumer != improvementConsumer ||
		receipt.MetaOperation != improvementOperation:
		return fmt.Errorf("FAIL_CLOSED: improvement meta binding mismatch")
	case !reflect.DeepEqual(receipt.Baseline, FoundationBaseline(receipt.Baseline)):
		return fmt.Errorf("FAIL_CLOSED: improvement baseline reference mismatch")
	case !validHeadSHA(receipt.HeadSHA) || receipt.HeadSHA != after.HeadSHA:
		return fmt.Errorf("FAIL_CLOSED: improvement head mismatch")
	case receipt.HeadSHA == before.HeadSHA:
		return fmt.Errorf("FAIL_CLOSED: improvement has no head transition")
	case receipt.BeforeArtifactDigest != before.ArtifactDigest ||
		receipt.AfterArtifactDigest != after.ArtifactDigest:
		return fmt.Errorf("FAIL_CLOSED: improvement artifact binding mismatch")
	case receipt.BeforeSnapshotDigest != before.Snapshot.Digest ||
		receipt.AfterSnapshotDigest != after.Snapshot.Digest:
		return fmt.Errorf("FAIL_CLOSED: improvement snapshot binding mismatch")
	case receipt.RepositoryWrites != 0:
		return fmt.Errorf("FAIL_CLOSED: improvement observer wrote repository")
	}
	expected := improvement.Evaluate(
		before.TransitionInput, after.TransitionInput,
	)
	if !reflect.DeepEqual(receipt.Transition, expected) {
		return fmt.Errorf("FAIL_CLOSED: improvement transition mismatch")
	}
	if err := requireAcceptedTransition(receipt.Transition); err != nil {
		return err
	}
	if receipt.ArtifactDigest != sealImprovement(receipt).ArtifactDigest {
		return fmt.Errorf("FAIL_CLOSED: improvement artifact digest mismatch")
	}
	return nil
}
