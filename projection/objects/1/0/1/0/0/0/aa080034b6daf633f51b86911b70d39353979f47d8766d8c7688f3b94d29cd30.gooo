package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

type surfaceFixture struct {
	Owner  string
	Suffix string
}

func fixtureID(name string) semantic.ID { return semantic.MustIdentity("urn:gooo:coupling:" + name) }
func testInput(t *testing.T, beforeSources, headSources []selectiveci.SourceInput, surfaces []surfaceFixture) Input {
	t.Helper()
	registrySurfaces := make([]detector.Surface, 0, len(surfaces))
	for _, fixture := range surfaces {
		surface := detector.Surface{
			SurfaceID: fixtureID("surface-" + fixture.Suffix), CodeSymbolID: fixtureID("symbol-" + fixture.Suffix),
			SemanticOwnerID: semantic.MustIdentity(fixture.Owner), Binding: detector.SourceMapBinding{SourceMapID: fixtureID("map-" + fixture.Suffix)},
		}
		surface.Binding.BindingDigest = bindingDigestForTest(surface)
		registrySurfaces = append(registrySurfaces, surface)
	}
	registry := detector.Registry{Schema: detector.RegistrySchemaV1, Surfaces: registrySurfaces}
	registry.Digest = registryDigestForTest(registry)
	before := testSnapshot(t, registry.Digest, beforeSources...)
	head := testSnapshot(t, registry.Digest, headSources...)
	return Input{
		Before: &before, Head: &head,
		Authority: detector.AuthorityContext{
			Schema: detector.AuthorityContextSchemaV1, Registry: registry,
			ToolchainDigest: rawTestDigest(testDigest("toolchain")), ProfileDigest: rawTestDigest(testDigest("profile")), SnapshotDigest: rawTestDigest(head.Digest),
			ExpectedProviderDigest: rawTestDigest(testDigest("provider")), ExpectedObserverDigest: rawTestDigest(testDigest("observer")),
			Baseline: testBaseline(), ExternalReceiptRequired: true,
		},
		SourceMap: SourceMapContext{
			Digest: rawTestDigest(head.SourceMapDigest), Before: testObservations(before, registrySurfaces),
			Head: testObservations(head, registrySurfaces),
		},
	}
}
func testSnapshot(t *testing.T, registryDigest string, sources ...selectiveci.SourceInput) selectiveci.Snapshot {
	t.Helper()
	ids := make([]string, 0)
	for _, source := range sources {
		for _, binding := range source.Bindings {
			ids = append(ids, binding.ID)
		}
	}
	result, err := selectiveci.Build(selectiveci.SnapshotInput{
		Sources: append([]selectiveci.SourceInput{}, sources...), SourceMapDigest: testDigest("source-map"),
		RegistryDigest: "sha256:" + registryDigest, RegisteredIDs: ids,
	})
	if err != nil {
		t.Fatalf("selectiveci.Build: %v", err)
	}
	return result
}
func testSource(t *testing.T, filename, name, id string) selectiveci.SourceInput {
	t.Helper()
	source := []byte("package fixture\n\n//gooo:bind id=\"" + id + "\" role=\"HANDWRITTEN_IMPL\"\nfunc " + name + "() {}\n")
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{Filename: filename, PackagePath: "fixture", Source: source}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semanticbinding.Extract = %#v, err=%v", result, err)
	}
	return selectiveci.SourceInput{Path: filename, BlobDigest: testDigest(string(source)), Bindings: result.Bindings}
}
