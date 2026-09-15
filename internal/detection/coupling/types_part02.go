package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type ManifestEntry struct {
	SurfaceID           semantic.ID `json:"surface_id"`
	CodeSymbolID        semantic.ID `json:"code_symbol_id"`
	SemanticOwnerID     semantic.ID `json:"semantic_owner_id"`
	BeforeBindingDigest string      `json:"before_binding_digest"`
	AfterBindingDigest  string      `json:"after_binding_digest"`
	BeforeBlobDigest    string      `json:"before_blob_digest"`
	AfterBlobDigest     string      `json:"after_blob_digest"`
	BeforeSourcePath    string      `json:"before_source_path"`
	AfterSourcePath     string      `json:"after_source_path"`
}
type ChangeManifest struct {
	Schema               string          `json:"schema"`
	Complete             bool            `json:"complete"`
	ZeroChange           bool            `json:"zero_change"`
	RegistryDigest       string          `json:"registry_digest"`
	ToolchainDigest      string          `json:"toolchain_digest"`
	ProfileDigest        string          `json:"profile_digest"`
	BeforeSnapshotDigest string          `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string          `json:"after_snapshot_digest"`
	Entries              []ManifestEntry `json:"entries"`
	Digest               string          `json:"digest"`
}
type BaselineConfig struct {
	Schema            string `json:"schema"`
	FullSuiteRequired bool   `json:"full_suite_required"`
	Digest            string `json:"digest"`
}
type Config struct {
	Schema                  string         `json:"schema"`
	RegistryDigest          string         `json:"registry_digest"`
	ToolchainDigest         string         `json:"toolchain_digest"`
	ProfileDigest           string         `json:"profile_digest"`
	SnapshotDigest          string         `json:"snapshot_digest"`
	ExpectedProviderDigest  string         `json:"expected_provider_digest"`
	ExpectedObserverDigest  string         `json:"expected_observer_digest"`
	Baseline                BaselineConfig `json:"baseline"`
	ExternalReceiptRequired bool           `json:"external_receipt_required"`
}

// ApplicabilityProof is evaluator-owned evidence that an empty registry is
// applicable to one exact immutable snapshot and policy tuple.
type ApplicabilityProof struct {
	Schema          string
	RegistryDigest  string
	ToolchainDigest string
	ProfileDigest   string
	SnapshotDigest  string
	AllowsEmpty     bool
	Digest          string
}

// AuthorityContext is supplied by the evaluator owner, never decoded from an
// Input packet. Its values define the registry, policy, snapshot, applicability
// and resource obligations against which producer claims are compared.
type AuthorityContext struct {
	Schema                  string
	Registry                Registry
	ToolchainDigest         string
	ProfileDigest           string
	SnapshotDigest          string
	ExpectedProviderDigest  string
	ExpectedObserverDigest  string
	Baseline                BaselineConfig
	Applicability           *ApplicabilityProof
	ExternalReceiptRequired bool
}
