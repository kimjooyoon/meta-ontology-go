package metricprogram_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
)

func TestCompileUnregisteredOriginFresh1727(t *testing.T) {
	for _, mode := range []string{"original-context", "missing-indicator", "missing-operation", "unicode-operation"} {
		t.Run(mode, func(t *testing.T) {
			payload, verification := fixturePayloads(t)
			var plan metricprogram.StrategyPlan
			if err := json.Unmarshal(payload, &plan); err != nil {
				t.Fatal(err)
			}
			plan.Bindings[0].MetaOperation = "unregistered://operation/new"
			if mode == "missing-indicator" {
				plan.Bindings[0].IndicatorID = ""
			}
			if mode == "missing-operation" {
				plan.Bindings[0].MetaOperation = ""
			}
			if mode == "unicode-operation" {
				plan.Bindings[0].MetaOperation = "gooo://user/연산/원본-유지"
			}
			program, source, err := metricprogram.Compile(fixtureJSON(t, plan), verification)
			var refusal *unresolvedpolicy.OperationRefusal
			if !errors.As(err, &refusal) {
				t.Fatalf("compiler discarded structured operation origin: %v", err)
			}
			original := plan.Bindings[0]
			registry := metricprogram.RegistrySourceOrigin()
			if program.Schema != "" || source != nil || refusal.Origin.Repository != plan.Repository ||
				refusal.Origin.SubjectSHA != plan.SubjectSHA || refusal.Origin.IndicatorID != original.IndicatorID ||
				refusal.Origin.RequestedOperation != original.MetaOperation ||
				refusal.Origin.Registry.SourcePath != registry.SourcePath ||
				refusal.Origin.Registry.SourceDigest != registry.SourceDigest ||
				refusal.Origin.Registry.ProjectionDigest != registry.ProjectionDigest ||
				refusal.ReadinessCredit || refusal.RepositoryWorkspaceWrites || refusal.PromotionAuthorized {
				t.Fatalf("refusal changed request or execution authority: %+v", refusal)
			}
			recorded := mode == "original-context" || mode == "unicode-operation"
			if refusal.OriginRecorded != recorded || refusal.Resolution != "invariant_only" ||
				refusal.Reason != unresolvedpolicy.OperationReason(false, recorded) ||
				refusal.NextOperation != unresolvedpolicy.OperationNextStep(false, recorded) {
				t.Fatalf("compiler bypassed source-owned classification: %+v", refusal)
			}
		})
	}
}
