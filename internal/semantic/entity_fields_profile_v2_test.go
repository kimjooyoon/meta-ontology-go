package semantic

import (
	"errors"
	"testing"
)

func TestEntityFieldsV2SemanticProfileIsSeparatelyBound(t *testing.T) {
	profile := EntityFieldsProfile{
		ID: EntityFieldsV2ProfileID, Version: EntityFieldsV2ProfileVersion, Digest: EntityFieldsV2ProfileDigest,
	}
	if err := profile.Validate(); err != nil {
		t.Fatalf("V2 profile rejected: %v", err)
	}
	profile.Digest = "tampered"
	if err := profile.Validate(); !errors.Is(err, ErrEntityFieldsProfileDigestError) {
		t.Fatalf("tampered V2 digest error = %v", err)
	}
}
