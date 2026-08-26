package coupling

// SourceManifest is an explicit completeness claim. ZeroChange is only
// authoritative when Complete is true and all snapshot bindings agree.
type SourceManifest struct {
	Complete             bool   `json:"complete"`
	ZeroChange           bool   `json:"zero_change"`
	BeforeSnapshotDigest string `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string `json:"after_snapshot_digest"`
	ToolchainDigest      string `json:"toolchain_digest"`
	ProfileDigest        string `json:"profile_digest"`
	RegistryDigest       string `json:"registry_digest"`
}

// SemanticNode keeps labels available for the rename partition. Canonical
// semantic identity deliberately excludes Name and Aliases.
type SemanticNode struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"`
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Aliases   []string `json:"aliases,omitempty"`
}
type SemanticRelation struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}
type SemanticIR struct {
	Nodes     []SemanticNode     `json:"nodes"`
	Relations []SemanticRelation `json:"relations"`
}
type CodeBinding struct {
	RegisteredSurfaceID string `json:"registered_surface_id"`
	CodeSymbolID        string `json:"code_symbol_id"`
	SemanticOwnerID     string `json:"semantic_owner_id"`
	SourceMapID         string `json:"source_map_id"`
	BindingDigest       string `json:"binding_digest"`
	PackageLabel        string `json:"package_label,omitempty"`
	FileLabel           string `json:"file_label,omitempty"`
	SourceSpan          string `json:"source_span,omitempty"`
}
type CodeChange struct {
	CodeSymbolID string `json:"code_symbol_id"`
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
}

// ExternalResourceReceipt is the only admissible source for resource
// observations. The oracle never derives CPU, peak memory, or work from
// fixture cardinality or elapsed wall time.
type ExternalResourceReceipt struct {
	ReceiptID      string `json:"receipt_id"`
	Metric         string `json:"metric"`
	Value          uint64 `json:"value"`
	Unit           string `json:"unit"`
	ProviderDigest string `json:"provider_digest"`
	ObserverDigest string `json:"observer_digest"`
	SnapshotDigest string `json:"snapshot_digest"`
	SourceDigest   string `json:"source_digest"`
	BindingDigest  string `json:"binding_digest"`
	Present        bool   `json:"present"`
	Independent    bool   `json:"independent"`
	State          string `json:"state"`
}
type ResourceObservation struct {
	CPUCoreNS       uint64 `json:"cpu_core_ns"`
	PeakMemoryBytes uint64 `json:"peak_memory_bytes"`
	WorkUnits       uint64 `json:"work_units"`
}
