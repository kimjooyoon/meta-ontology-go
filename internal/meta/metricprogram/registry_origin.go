package metricprogram

import (
	"embed"
	"fmt"
	"io/fs"
	"path"

	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

//go:generate go run ../../../scripts/metric-registry-codegen registry.gooo registry.generated.go

//go:embed registry.gooo
var registrySource []byte

//go:embed registry.projection.json
var registryProjection []byte

//go:embed registry*.generated.go
var registryParts embed.FS

// RegistryOrigin binds the executable declaration to the projection used by this binary.
type RegistryOrigin struct {
	SourcePath       string `json:"source_path"`
	SourceDigest     string `json:"source_digest"`
	ProjectionPath   string `json:"projection_path"`
	ProjectionDigest string `json:"projection_digest"`
	OperationCount   int64  `json:"operation_count"`
	EntityCount      int64  `json:"entity_count"`
}

func RegistrySourceOrigin() RegistryOrigin {
	return RegistryOrigin{
		SourcePath: "internal/meta/metricprogram/registry.gooo", SourceDigest: bytesDigest(registrySource),
		ProjectionPath:   "internal/meta/metricprogram/registry.projection.json",
		ProjectionDigest: bytesDigest(registryProjection), OperationCount: OperationCount(0), EntityCount: EntityCount(0),
	}
}

func VerifyCompiledRegistry() error {
	return projectionbundle.Verify(registrySource, registryProjection, registryParts, registryParts)
}

// VerifyRegistrySource also reads the subject repository; receipt self-hashes alone cannot bind origin.
func VerifyRegistrySource(repository fs.FS, origin RegistryOrigin) error {
	if repository == nil || origin != RegistrySourceOrigin() {
		return fmt.Errorf("metric registry origin differs from compiled declaration")
	}
	for _, member := range []struct{ path, digest string }{
		{origin.SourcePath, origin.SourceDigest}, {origin.ProjectionPath, origin.ProjectionDigest},
	} {
		raw, err := fs.ReadFile(repository, member.path)
		if err != nil {
			return fmt.Errorf("read metric registry member %s: %w", member.path, err)
		}
		if bytesDigest(raw) != member.digest {
			return fmt.Errorf("metric registry member %s differs from compiled origin", member.path)
		}
	}
	actual, err := fs.Sub(repository, path.Dir(origin.SourcePath))
	if err != nil {
		return err
	}
	if err := projectionbundle.Verify(registrySource, registryProjection, registryParts, actual); err != nil {
		return fmt.Errorf("metric registry member bundle: %w", err)
	}
	return nil
}
