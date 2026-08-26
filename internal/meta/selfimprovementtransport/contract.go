package selfimprovementtransport

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var expectedEntities = map[string]string{
	"TransportInput":              "gooo://self-improvement/transport/entity/input",
	"SourceIdentityEvidence":      "gooo://self-improvement/transport/evidence/source-identity",
	"CheckoutBindingEvidence":     "gooo://self-improvement/transport/evidence/checkout-binding",
	"ProducerIdentityEvidence":    "gooo://self-improvement/transport/evidence/producer-identity",
	"LogicalDigestEvidence":       "gooo://self-improvement/transport/evidence/logical-digest",
	"ImmutableLocatorEvidence":    "gooo://self-improvement/transport/evidence/immutable-locator",
	"ArchiveDigestEvidence":       "gooo://self-improvement/transport/evidence/archive-digest",
	"ConsumerReplayEvidence":      "gooo://self-improvement/transport/evidence/consumer-replay",
	"ProducerAttestationEvidence": "gooo://self-improvement/transport/evidence/producer-attestation",
	"TransportReceipt":            "gooo://self-improvement/transport/entity/receipt",
}

var expectedActivities = map[string][2]string{
	"ObserveSourceIdentity":      {"TransportInput", "SourceIdentityEvidence"},
	"ObserveCheckoutBinding":     {"TransportInput", "CheckoutBindingEvidence"},
	"ObserveProducerIdentity":    {"TransportInput", "ProducerIdentityEvidence"},
	"ObserveLogicalDigest":       {"TransportInput", "LogicalDigestEvidence"},
	"ObserveImmutableLocator":    {"TransportInput", "ImmutableLocatorEvidence"},
	"ObserveArchiveDigest":       {"TransportInput", "ArchiveDigestEvidence"},
	"ObserveConsumerReplay":      {"TransportInput", "ConsumerReplayEvidence"},
	"ObserveProducerAttestation": {"TransportInput", "ProducerAttestationEvidence"},
	"ReduceTransport":            {"TransportInput", "TransportReceipt"},
}

func CompileContract(repository fs.FS, path string) (ContractEvidence, error) {
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return ContractEvidence{}, fmt.Errorf("read transport contract: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return ContractEvidence{}, fmt.Errorf("parse transport contract")
	}
	canonical, err := syntax.Format(file)
	if err != nil || !contractKnown(file) {
		return ContractEvidence{}, fmt.Errorf("transport contract declarations mismatch")
	}
	lines := bytes.Count(raw, []byte{'\n'})
	if len(raw) != 0 && raw[len(raw)-1] != '\n' {
		lines++
	}
	return ContractEvidence{
		ContractID: ContractID, Path: path, Package: file.Package.Name,
		Namespace: file.Namespace.Name, EntityCount: len(expectedEntities),
		ActivityCount: len(expectedActivities), ObligationTotal: fixedObligationTotal,
		SourceLines: lines, SourceDigest: digestBytes(raw), CanonicalDigest: digestBytes([]byte(canonical)),
	}, nil
}

func contractKnown(file *syntax.File) bool {
	if file == nil || file.Package == nil || file.Namespace == nil ||
		file.Package.Name != "selfimprovementtransport" || file.Namespace.Name != "selfimprovementtransport" {
		return false
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	if len(declarations) != len(expectedEntities)+len(expectedActivities) {
		return false
	}
	entities, activities := map[string]string{}, map[string][2]string{}
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			if value.FieldsPresent || len(value.Fields) != 0 {
				return false
			}
			entities[value.Name] = value.ID
		case *syntax.ActivityDecl:
			inputs := value.Inputs
			if inputs == nil {
				inputs = value.Parameters
			}
			if len(inputs) != 1 {
				return false
			}
			activities[value.Name] = [2]string{inputs[0].Name, value.Output}
		default:
			return false
		}
	}
	return equalMap(entities, expectedEntities) && equalMap(activities, expectedActivities)
}

func equalMap[K comparable, V comparable](left, right map[K]V) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
