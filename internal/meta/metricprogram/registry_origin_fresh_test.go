package metricprogram

import (
	"testing"
	"testing/fstest"
)

func TestRegistryOriginFresh1635(t *testing.T) {
	origin := RegistrySourceOrigin()
	for _, item := range []struct {
		name, source, projection string
		allowed                  bool
	}{
		{"exact-members", string(registrySource), string(registryProjection), true},
		{"changed-source", string(registrySource) + "\n", string(registryProjection), false},
		{"changed-projection", string(registrySource), string(registryProjection) + "\n", false},
		{"missing-source", "", string(registryProjection), false},
	} {
		t.Run(item.name, func(t *testing.T) {
			members := registryFilesFixture(t)
			members[origin.ProjectionPath] = &fstest.MapFile{Data: []byte(item.projection)}
			if item.source != "" {
				members[origin.SourcePath] = &fstest.MapFile{Data: []byte(item.source)}
			} else {
				delete(members, origin.SourcePath)
			}
			if err := VerifyRegistrySource(members, origin); (err == nil) != item.allowed {
				t.Fatalf("origin verification: %v", err)
			}
		})
	}
	t.Run("resealed-origin", func(t *testing.T) {
		origin.SourceDigest = bytesDigest([]byte("different declaration"))
		if err := VerifyRegistrySource(fstest.MapFS{}, origin); err == nil {
			t.Fatal("self-described origin accepted")
		}
	})
}
