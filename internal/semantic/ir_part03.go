package semantic

import (
	"fmt"
	"strings"
	"unicode"
)

func (ir IR) SemanticCanonical() string {
	version, packageName, namespace := canonicalIRMetadata(ir)
	var b strings.Builder
	b.WriteString("ir\t")
	b.WriteString(version)
	b.WriteByte('\t')
	b.WriteString(packageName)
	b.WriteByte('\t')
	b.WriteString(namespace)
	b.WriteByte('\n')
	b.WriteString(ir.Graph.SemanticCanonical())
	for _, policy := range ir.Policies {
		b.WriteString(policy.SemanticCanonical())
	}
	return b.String()
}
func (ir IR) StableHash() string {
	return StableHashString(ir.SemanticCanonical())
}
func (ir IR) Hash() string {
	return ir.StableHash()
}
func canonicalIRMetadata(ir IR) (string, string, string) {
	version := strings.TrimSpace(ir.Version)
	if version == "" {
		version = CurrentIRVersion
	}
	packageName := strings.TrimSpace(ir.Package)
	namespace := strings.TrimSpace(ir.Namespace.String())
	if parsed, err := ParseNamespace(namespace); err == nil {
		namespace = parsed.String()
	}
	return version, packageName, namespace
}
func validatePackageName(packageName string) error {
	if packageName == "" {
		return nil
	}
	if strings.IndexFunc(packageName, unicode.IsSpace) >= 0 {
		return fmt.Errorf("%w: package name contains whitespace", ErrGraphInvalid)
	}
	return nil
}
