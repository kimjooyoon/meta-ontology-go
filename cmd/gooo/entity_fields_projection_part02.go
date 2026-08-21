package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"strings"
)

func sameCLIField(projected generator.Field, semanticField generator.Field) bool {
	return projected.ID == semanticField.ID && projected.Parent == semanticField.Parent &&
		projected.Name == semanticField.Name && projected.TypeRefID == semanticField.TypeRefID &&
		projected.Presence == semanticField.Presence && projected.Cardinality == semanticField.Cardinality &&
		projected.Source == semanticField.Source
}
func classifyCLIEntityFieldsModelError(err error) error {
	message := err.Error()
	if strings.Contains(message, "GOOO-EF-V1-") {
		return err
	}
	switch {
	case strings.Contains(message, "duplicate field ID"), strings.Contains(message, "collides with"):
		return fmt.Errorf("GOOO-EF-V1-ID-COLLISION: %w", err)
	case strings.Contains(message, "parent"):
		return fmt.Errorf("GOOO-EF-V1-WRONG-PARENT: %w", err)
	case strings.Contains(message, "type ref"), strings.Contains(message, "unknown semantic type"), strings.Contains(message, "ambiguous semantic type"):
		return fmt.Errorf("GOOO-EF-V1-UNKNOWN-TYPE: %w", err)
	case strings.Contains(message, "span"):
		return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: %w", err)
	default:
		return fmt.Errorf("GOOO-EF-V1-INCOMPLETE-FIELD: %w", err)
	}
}
func validateCLIEntityFieldsSupport(support syntax.EntityFieldsSupport) error {
	if support.State != syntax.EntityFieldsDeferred && support.State != syntax.EntityFieldsSupported {
		return fmt.Errorf("GOOO-EF-V1-UNKNOWN-STATE: %q", support.State)
	}
	if support.Profile.ID == "" && support.Profile.Version == 0 && support.Profile.Digest == "" {
		return errors.New("GOOO-EF-V1-UNBOUND-PROFILE: profile is unbound")
	}
	if support.Profile.ID != syntax.EntityFieldsProfileID || support.Profile.Version != syntax.EntityFieldsProfileVersion {
		return errors.New("GOOO-EF-V1-PROFILE-MISMATCH: profile identity or version does not match")
	}
	if support.Profile.Digest != syntax.EntityFieldsProfileDigest {
		return errors.New("GOOO-EF-V1-PROFILE-DIGEST-MISMATCH: profile digest does not match")
	}
	return nil
}
