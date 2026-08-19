package coupling

const detectorAuthorityHead = "02e35a01946c20c5de67f2ec71eeca5ac6c3aedb"

type productionBindingVector struct {
	AuthoritySchema                  string
	AuthorityRegistryDigest          string
	AuthorityToolchainDigest         string
	AuthorityProfileDigest           string
	AuthoritySnapshotDigest          string
	AuthorityProviderDigest          string
	AuthorityObserverDigest          string
	AuthorityBaselineDigest          string
	AuthorityExternalReceiptRequired bool
	PacketRegistryDigest             string
	ConfigRegistryDigest             string
	ConfigToolchainDigest            string
	ConfigProfileDigest              string
	ConfigSnapshotDigest             string
	ExpectedProviderDigest           string
	ExpectedObserverDigest           string
	BaselineDigest                   string
	ManifestComplete                 bool
	ManifestZeroChange               bool
	ManifestRegistryDigest           string
	ManifestToolchainDigest          string
	ManifestProfileDigest            string
	ManifestBeforeSnapshotDigest     string
	ManifestAfterSnapshotDigest      string
	ManifestDigest                   string
	PathDigest                       string
	ExternalReceiptDigest            string
	Surfaces                         []productionSurfaceBinding
	ManifestEntries                  []productionManifestBinding
	Receipts                         []productionReceiptBinding
	ExternalReceipt                  productionExternalBinding
}
type productionSurfaceBinding struct {
	SurfaceID       string
	CodeSymbolID    string
	SemanticOwnerID string
	SourceMapID     string
	BindingDigest   string
}
type productionManifestBinding struct {
	SurfaceID           string
	CodeSymbolID        string
	SemanticOwnerID     string
	BeforeBindingDigest string
	AfterBindingDigest  string
	BeforeBlobDigest    string
	AfterBlobDigest     string
}
