package metricprogram_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
)

func TestRefusalConsumerFresh1737(t *testing.T) {
	for _, mode := range []string{"retained", "missing-ID", "repository", "subject", "requested-operation", "indicator",
		"strategy-bytes", "verification-bytes", "registry", "policy", "policy-source", "policy-member", "registry-source",
		"classification", "next-step", "origin-recorded", "readiness", "writes", "promotion", "schema", "registered"} {
		t.Run(mode, func(t *testing.T) {
			strategy, verification := fixturePayloads(t)
			var plan metricprogram.StrategyPlan
			if err := json.Unmarshal(strategy, &plan); err != nil {
				t.Fatal(err)
			}
			plan.Bindings[0].MetaOperation = "gooo://user/미등록/consumer-1737"
			if mode == "missing-ID" {
				plan.Bindings[0].IndicatorID = ""
			}
			strategy = fixtureJSON(t, plan)
			_, _, err := metricprogram.Compile(strategy, verification)
			var refusal *unresolvedpolicy.OperationRefusal
			if !errors.As(err, &refusal) {
				t.Fatalf("new consumer input lost diagnostic: %v", err)
			}
			repository := mutateRefusal1737(mode, refusal)
			if mode == "strategy-bytes" {
				strategy = append(strategy, ' ')
			}
			if mode == "verification-bytes" {
				verification = append(verification, '\n')
			}
			if mode == "registered" {
				plan.Bindings[0].MetaOperation = "project-algebraic-root-state"
				strategy = fixtureJSON(t, plan)
				refusal.Origin.Inputs.StrategyPayloadDigest = fixtureBytesDigest1737(strategy)
				refusal.Origin.RequestedOperation = plan.Bindings[0].MetaOperation
			}
			report, err := programverify.VerifyOperationRefusal(repository, plan.Repository, plan.SubjectSHA,
				strategy, verification, fixtureJSON(t, refusal))
			valid := mode == "retained" || mode == "missing-ID"
			if (err == nil) != valid {
				t.Fatalf("independent source reconstruction accepted=%v, want=%v: %v", err == nil, valid, err)
			}
			if err == nil && (report.Status != "VERIFIED_SOURCE_ORIGIN_REFUSAL" || report.Origin != refusal.Origin ||
				report.UpstreamVerification != "UNASSESSED" || report.ReadinessCredit ||
				report.RepositoryWorkspaceWrites || report.PromotionAuthorized) {
				t.Fatalf("diagnostic verification changed authority or origin: %+v", report)
			}
		})
	}
}

func mutateRefusal1737(mode string, refusal *unresolvedpolicy.OperationRefusal) fs.FS {
	switch mode {
	case "repository":
		refusal.Origin.Repository = "other/repository"
	case "subject":
		refusal.Origin.SubjectSHA = "0000000000000000000000000000000000000000"
	case "requested-operation":
		refusal.Origin.RequestedOperation += "changed"
	case "indicator":
		refusal.Origin.IndicatorID += "changed"
	case "registry":
		refusal.Origin.Registry.SourceDigest = fixtureDigest("other registry")
	case "policy":
		refusal.Policy.SourceDigest = fixtureDigest("other policy")
	case "classification":
		refusal.Reason = "COMPATIBILITY_KERNEL_UNSUPPORTED"
	case "next-step":
		refusal.NextOperation = "CONTINUE_SOURCE_OPERATION"
	case "origin-recorded":
		refusal.OriginRecorded = false
	case "readiness":
		refusal.ReadinessCredit = true
	case "writes":
		refusal.RepositoryWorkspaceWrites = true
	case "promotion":
		refusal.PromotionAuthorized = true
	case "schema":
		refusal.Schema = "gooo/metric-meta-program/v1"
	}
	return alteredRefusalFS1737(mode)
}

type refusalFS1737 struct {
	fs.FS
	changed fstest.MapFS
}

func (repository refusalFS1737) Open(name string) (fs.File, error) {
	if _, ok := repository.changed[name]; ok {
		return repository.changed.Open(name)
	}
	return repository.FS.Open(name)
}

func alteredRefusalFS1737(mode string) fs.FS {
	changed := fstest.MapFS{}
	member := map[string]string{
		"policy-source":   "internal/meta/metricprogram/unresolvedpolicy/policy.gooo",
		"policy-member":   "internal/meta/metricprogram/unresolvedpolicy/policy.generated.go",
		"registry-source": "internal/meta/metricprogram/registry.gooo",
	}[mode]
	if member != "" {
		changed[member] = &fstest.MapFile{Data: []byte("changed source snapshot")}
	}
	return refusalFS1737{FS: os.DirFS("../../.."), changed: changed}
}
