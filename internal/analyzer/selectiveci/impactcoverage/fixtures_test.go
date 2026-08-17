package impactcoverage

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
)

func boundSource(path, blob string, ids ...string) selectiveci.SourceInput {
	bindings := make([]semanticbinding.Binding, 0, len(ids))
	for _, id := range ids {
		binding := semanticbinding.Binding{
			ID: id, Role: semanticbinding.RoleHandwrittenImpl,
			Span: semanticbinding.Span{
				Filename: path,
				Start:    semanticbinding.Position{Offset: 1, Line: 1, Column: 1},
				End:      semanticbinding.Position{Offset: 2, Line: 1, Column: 2},
			},
		}
		binding.Digest = binding.StableHash()
		binding.CanonicalDigest = binding.Digest
		bindings = append(bindings, binding)
	}
	return selectiveci.SourceInput{Path: path, BlobDigest: blobDigest(blob), Bindings: bindings}
}

func emptySource(path, blob string) selectiveci.SourceInput {
	return selectiveci.SourceInput{Path: path, BlobDigest: blobDigest(blob), Bindings: []semanticbinding.Binding{}}
}

func snap(t *testing.T, sourceMap, registry string, sources ...selectiveci.SourceInput) selectiveci.Snapshot {
	t.Helper()
	ids := []string{}
	for _, source := range sources {
		for _, binding := range source.Bindings {
			ids = append(ids, binding.ID)
		}
	}
	result, err := selectiveci.Build(selectiveci.SnapshotInput{
		Sources: sources, SourceMapDigest: digest(sourceMap),
		RegistryDigest: digest(registry), RegisteredIDs: ids,
	})
	if err != nil {
		t.Fatalf("Build snapshot: %v", err)
	}
	return result
}

func digest(value string) string { return blobDigest(value) }

func blobDigest(value string) string { return "sha256:" + hexDigest(value) }

func hexDigest(value string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(value))) }
