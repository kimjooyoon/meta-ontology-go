package toolchainconformance

type conceptArtifact struct {
	Decision       string `json:"decision"`
	CatalogDigest  string `json:"catalog_digest"`
	ArtifactDigest string `json:"artifact_digest"`
	Report         struct {
		Concepts []conceptItem `json:"concepts"`
	} `json:"report"`
}

type conceptItem struct {
	ID             string   `json:"id"`
	MetaOperation  string   `json:"meta_operation"`
	Stage          string   `json:"stage"`
	NoveltyClaim   bool     `json:"novelty_claim"`
	CodeBindings   []string `json:"code_bindings"`
	MetricBindings []string `json:"metric_bindings"`
	UseCases       []struct {
		ID string `json:"id"`
	} `json:"use_cases"`
}

type conceptCounts struct {
	ArtifactDigest  string
	CatalogDigest   string
	ConceptBindings int
	CodeBindings    int
	MetricBindings  int
	UseCaseBindings int
}

var expectedCodeBindings = []string{
	"internal/meta/languagereadiness/toolchainconformance",
	"cmd/toolchain-conformance-witness",
	"examples/toolchain-conformance",
	".github/workflows/transformation-effect.yml",
	"docs/language/toolchain-conformance.md",
}

var expectedUseCases = []string{
	"same-head-surface-closure",
	"in-memory-drift-rejection",
	"effect-authority-boundary",
}
