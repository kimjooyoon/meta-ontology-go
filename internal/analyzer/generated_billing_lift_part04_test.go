package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func generatedBillingSource(t *testing.T) SourceFile {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root = filepath.Clean(filepath.Join(root, "..", ".."))
	out := t.TempDir()
	command := exec.Command("go", "run", "./cmd/gooo", "generate", "examples/billing/main.gooo", "--out", out)
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generate billing projection: %v\n%s", err, output)
	}
	projection, err := os.ReadFile(filepath.Join(out, "semantic.gooo.go"))
	if err != nil {
		t.Fatalf("read generated billing projection: %v", err)
	}
	return SourceFile{Filename: "semantic.gooo.go", PackagePath: "billing", Source: projection}
}
func generatedBillingRegistry(t *testing.T) *Registry {
	t.Helper()
	return generatedBillingRegistryWithOrder(t, "billing://entity/order")
}
func generatedBillingRegistryWithOrder(t *testing.T, orderID string) *Registry {
	t.Helper()
	registry := NewRegistry()
	for _, entry := range []Registration{
		{Ref: billingRef("PayOrder"), Kind: KindActivity, Identity: NewIdentity("billing", "billing://activity/pay-order")},
		{Ref: billingRef("Order"), Kind: KindEntity, Identity: NewIdentity("billing", orderID)},
		{Ref: billingRef("PaymentMethod"), Kind: KindEntity,
			Identity: NewIdentity("billing", "billing://entity/payment-method")},
		{Ref: billingRef("Payment"), Kind: KindEntity, Identity: NewIdentity("billing", "billing://entity/payment")},
	} {
		if err := registry.Register(entry); err != nil {
			t.Fatal(err)
		}
	}
	return registry
}
func generatedBillingPolicy(t *testing.T) MappingPolicy {
	t.Helper()
	policy, err := NewMappingPolicy(CurrentSemanticAdapterPolicy)
	if err != nil {
		t.Fatal(err)
	}
	for _, mapping := range []RelationMapping{
		{Source: RelationUses, Predicate: semantic.Used,
			SourceSubjectKind: semantic.Activity, SourceObjectKind: semantic.Entity,
			AllowedOrigins: []ObservationOrigin{OriginSignature}},
		{Source: RelationGenerates, Predicate: semantic.WasGeneratedBy,
			SourceSubjectKind: semantic.Activity, SourceObjectKind: semantic.Entity,
			Reverse: true, AllowedOrigins: []ObservationOrigin{OriginSignature}},
	} {
		if err := policy.Register(mapping); err != nil {
			t.Fatal(err)
		}
	}
	return policy
}
