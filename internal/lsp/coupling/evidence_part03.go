package coupling

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateSourceMapBinding(id, digest string) error {
	if err := validateIdentity(id, "source-map"); err != nil {
		return fmt.Errorf("source-map ID: %w", err)
	}
	if err := validateDigest(digest); err != nil {
		return fmt.Errorf("source-map digest: %w", err)
	}
	return nil
}
func validateIdentity(raw, kind string) error {
	parsed, err := semantic.ParseIdentity(raw)
	if err != nil {
		return fmt.Errorf("%s ID: %w", kind, err)
	}
	if parsed.String() != raw {
		return fmt.Errorf("%s ID is not canonical", kind)
	}
	return nil
}
