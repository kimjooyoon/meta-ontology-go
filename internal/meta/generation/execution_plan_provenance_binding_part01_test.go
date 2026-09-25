package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestBindExecutionPlanProvenancePart01KeepsEvidenceNonAuthorizing(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	chain := provenance.BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	binding := BindExecutionPlanProvenancePart01("compile-gooo", digest, "model/gooo-planner-v1", provenance.GatewayPolicy{AllowedHosts: []string{"registry.example.com"}}, provenance.ExecutionPlanLifecycleResumed, chain)
	if binding.Status != provenance.ExecutionPlanBindingBound || binding.AdoptionAuthorized || !binding.NonAuthorizing { t.Fatalf("binding = %#v", binding) }
	if err := binding.Validate(); err != nil { t.Fatalf("binding validation failed: %v", err) }
}