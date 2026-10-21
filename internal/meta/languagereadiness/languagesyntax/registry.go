package languagesyntax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func expectedRegistry() Registry {
	return Registry{Schema: RegistrySchema, ProjectionUnits: expectedProjectionUnits(), Cases: expectedSyntaxCases(), PackageUnits: expectedSyntaxPackageUnits(), MetaSources: []string{"internal/meta/syntaxregistration/contract.gooo", "internal/meta/entityfields/entity-fields-meta.gooo", "examples/public-trust-surface/main.gooo", "internal/meta/generation/operation-input-contract.gooo", "internal/meta/generation/callback-preview-contract.gooo", "internal/meta/generation/callback-extraction-contract.gooo", "internal/meta/generation/self_improvement_outcome_contract.gooo", "internal/meta/policycompilation/revision-operation.gooo", "examples/relay-game-contract/main.gooo", "internal/meta/generation/experience_memory_receipt_contract.gooo", "internal/meta/syntaxregistration/profile_migration.gooo", "examples/source-owned-cohort-observations/main.gooo", "internal/meta/publictestreuse/policy.gooo", "examples/measurement-metadata/main.gooo", "internal/meta/languagereadiness/toolchainrelease/payload.gooo", "examples/agent-tool-inventory/main.gooo", "examples/recorded-use-origins/retention.gooo", "examples/toolchain-version-impact/impact.gooo", "examples/toolchain-version-impact/plan.gooo", "examples/direct-module-impact/plan.gooo", "examples/ci-strategy/strategy.gooo"}}
}

func decodeRegistry(raw []byte) (Registry, error) {
	registry := Registry{}
	if err := validateJSONKeyUniqueness(raw); err != nil {
		return registry, fmt.Errorf("decode language syntax registry: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return registry, fmt.Errorf("decode language syntax registry: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return registry, fmt.Errorf("decode language syntax registry: trailing content")
	}
	if err := validateCaseScopes(registry); err != nil {
		return registry, err
	}
	if !reflect.DeepEqual(registry, expectedRegistry()) {
		return registry, fmt.Errorf("language syntax registry mismatch")
	}
	return registry, nil
}

func validateJSONKeyUniqueness(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := walkJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func walkJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate JSON object key %q", key)
			}
			seen[key] = true
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}

func CapabilityCaseTotal() int {
	return FixedCapabilityTotal
}

func validateCaseScopes(registry Registry) error {
	if len(registry.Cases) != FixedTotal || FixedCapabilityTotal+FixedGovernanceTotal != FixedTotal {
		return fmt.Errorf("language syntax scope denominator mismatch")
	}
	capability, governance := 0, 0
	governanceIDs, governancePaths := []string{}, []string{}
	for _, definition := range registry.Cases {
		switch definition.Scope {
		case ScopeLanguageCapability:
			capability++
		case ScopeGovernanceObservation:
			governance++
			governanceIDs = append(governanceIDs, definition.ID)
			governancePaths = append(governancePaths, definition.Path)
		default:
			return fmt.Errorf("language syntax case %q has missing or unknown scope", definition.ID)
		}
	}
	if capability != FixedCapabilityTotal || governance != FixedGovernanceTotal ||
		!reflect.DeepEqual(governanceIDs, []string{"self-improvement-ci-continuation", "domain-completeness-profile"}) ||
		!reflect.DeepEqual(governancePaths, []string{"examples/self-improvement-ci-continuation/continuation.gooo", "scripts/domain-completeness/profile.gooo"}) {
		return fmt.Errorf("language syntax scope partition mismatch")
	}
	return nil
}

func unresolvedCases(source Source) []CaseResult {
	results := make([]CaseResult, 0, totalCases)
	for _, definition := range expectedRegistry().Cases {
		item := CaseResult{Definition: definition, Evidence: replay.Result{ObservedDecision: replay.DecisionUnknown}, Status: "UNRESOLVED"}
		item.EvidenceDigest = caseDigest(item, source)
		results = append(results, item)
	}
	return results
}
