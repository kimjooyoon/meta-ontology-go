package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

type couplingFixture struct {
	input            Input
	authorityContext AuthorityContext
	owner            semantic.ID
	code             semantic.ID
	surface          semantic.ID
	authority        semantic.ID
	projection       semantic.ID
	verification     semantic.ID
}
type fixtureContext struct {
	owner, code, surface          semantic.ID
	registry                      Registry
	config                        Config
	manifest                      ChangeManifest
	beforeBlob, afterBlob         string
	beforeSemantic, afterSemantic string
	beforeSource, afterSource     string
}
type fixtureProof struct {
	path                                semantic.InferencePathV1
	receipt                             CouplingReceipt
	authority, projection, verification semantic.ID
}

func fixtureID(name string) semantic.ID { return semantic.MustIdentity("gooo://coupling/" + name) }
func fixtureDigest(name string) string  { return semantic.StableHashString("coupling-fixture/" + name) }
func newFixture(t *testing.T, claim ChangeClaim) couplingFixture {
	t.Helper()
	context := fixtureContextFor(t, claim)
	proof := fixtureProofFor(context, claim)
	external := externalReceiptFor(context.config)
	authorityRegistry := context.registry
	authorityRegistry.Surfaces = append([]Surface(nil), context.registry.Surfaces...)
	authority := AuthorityContext{
		Schema: AuthorityContextSchemaV1, Registry: authorityRegistry,
		ToolchainDigest: context.config.ToolchainDigest, ProfileDigest: context.config.ProfileDigest,
		SnapshotDigest:         context.config.SnapshotDigest,
		ExpectedProviderDigest: context.config.ExpectedProviderDigest, ExpectedObserverDigest: context.config.ExpectedObserverDigest,
		Baseline: context.config.Baseline, ExternalReceiptRequired: true,
	}
	return couplingFixture{
		input: Input{Schema: InputSchemaV1, Config: context.config, Registry: context.registry, Manifest: context.manifest,
			Receipts: []CouplingReceipt{proof.receipt}, InferencePath: proof.path, ExternalReceipt: external, WorkspaceRoot: "/workspace"},
		authorityContext: authority,
		owner:            context.owner, code: context.code, surface: context.surface,
		authority: proof.authority, projection: proof.projection, verification: proof.verification,
	}
}
