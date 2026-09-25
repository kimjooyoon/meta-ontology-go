package semantic

import (
	"fmt"
	"strings"
)

func (ir IR) Normalized() (IR, error) {
	version := strings.TrimSpace(ir.Version)
	if version == "" {
		version = CurrentIRVersion
	}
	if err := validateIRVersion(version); err != nil {
		return IR{}, err
	}
	packageName := strings.TrimSpace(ir.Package)
	if err := validatePackageName(packageName); err != nil {
		return IR{}, err
	}
	namespace := ir.Namespace
	if namespace != "" {
		parsed, err := ParseNamespace(namespace.String())
		if err != nil {
			return IR{}, err
		}
		namespace = parsed
	}
	graph, err := ir.Graph.Normalized()
	if err != nil {
		return IR{}, err
	}
	runtimeBindings, err := normalizeRuntimeBindings(ir.RuntimeBindings, graph)
	if err != nil {
		return IR{}, err
	}
	out := IR{Version: version, Package: packageName, Namespace: namespace, Graph: graph, RuntimeBindings: runtimeBindings, evidence: make(map[ID]Evidence)}
	if len(ir.Policies) > 0 {
		out.Policies = make([]Policy, len(ir.Policies))
		for index, policy := range ir.Policies {
			normalizedPolicy, err := policy.Normalized()
			if err != nil {
				return IR{}, err
			}
			out.Policies[index] = normalizedPolicy
		}
	}
	for _, evidence := range ir.Evidence() {
		if err := out.AddEvidence(evidence); err != nil {
			return IR{}, err
		}
	}
	if err := out.validateEvidence(); err != nil {
		return IR{}, err
	}
	return out, nil
}
func validateIRVersion(raw string) error {
	version := strings.TrimSpace(raw)
	if version == "" {
		return fmt.Errorf("%w: IR version is empty", ErrGraphInvalid)
	}
	if version != CurrentIRVersion {
		return fmt.Errorf("%w: unsupported IR version %q (want %q)", ErrGraphInvalid, version, CurrentIRVersion)
	}
	return nil
}
func (ir *IR) Normalize() error {
	normalized, err := ir.Normalized()
	if err != nil {
		return err
	}
	*ir = normalized
	return nil
}
func (ir IR) Canonical() string {
	version, packageName, namespace := canonicalIRMetadata(ir)
	var b strings.Builder
	b.WriteString("ir\t")
	b.WriteString(version)
	b.WriteByte('\t')
	b.WriteString(packageName)
	b.WriteByte('\t')
	b.WriteString(namespace)
	b.WriteByte('\n')
	b.WriteString(ir.Graph.Canonical())
	for _, binding := range ir.RuntimeBindings {
		b.WriteString(binding.Canonical())
		b.WriteByte('\n')
	}
	for _, policy := range ir.Policies {
		b.WriteString(policy.Canonical())
	}
	b.WriteString(ir.EvidenceCanonical())
	return b.String()
}
