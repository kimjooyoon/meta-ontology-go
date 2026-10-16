package unresolvedpolicy

import (
	"embed"

	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

//go:generate go run ../../../../scripts/metric-unresolved-policy-codegen policy.gooo .

//go:embed policy.gooo
var source []byte

//go:embed policy.projection.json
var manifest []byte

//go:embed policy*.generated.go
var members embed.FS

type ProjectionOrigin struct {
	SourcePath       string `json:"source_path"`
	SourceDigest     string `json:"source_digest"`
	ProjectionPath   string `json:"projection_path"`
	ProjectionDigest string `json:"projection_digest"`
}

func Origin() ProjectionOrigin {
	return ProjectionOrigin{
		SourcePath:       "internal/meta/metricprogram/unresolvedpolicy/policy.gooo",
		SourceDigest:     projectionbundle.Digest(source),
		ProjectionPath:   "internal/meta/metricprogram/unresolvedpolicy/policy.projection.json",
		ProjectionDigest: projectionbundle.Digest(manifest),
	}
}

func VerifyProjection() error {
	return projectionbundle.Verify(source, manifest, members, members)
}
