//go:build detector_bridge

package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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

type productionReceiptBinding struct {
	Schema                      string
	ReceiptID                   string
	SurfaceID                   string
	SemanticOwnerID             string
	CodeSymbolID                string
	SourceMapBindingDigest      string
	SnapshotDigest              string
	RegistryDigest              string
	ToolchainDigest             string
	ProfileDigest               string
	BeforeBlobDigest            string
	AfterBlobDigest             string
	BeforeAuthoritySourceDigest string
	AfterAuthoritySourceDigest  string
	BeforeSemanticDigest        string
	AfterSemanticDigest         string
	ChangeClaim                 production.ChangeClaim
	ReceiptKind                 semantic.SemanticChangeKind
	OriginPathIDs               []string
	ClaimID                     string
	EvidenceIDs                 []string
	EvidenceDigests             []string
	CanonicalDelta              string
	DeltaDigest                 string
	AuthoritySourceID           string
	AuthoritySourcePath         string
	State                       string
}

type productionExternalBinding struct {
	Schema                 string
	SnapshotDigest         string
	ProviderDigest         string
	ObserverDigest         string
	CPUWorkUnits           *uint64
	PeakMemoryBytes        *uint64
	DeterministicWorkUnits *uint64
	Digest                 string
}

type productionVector struct {
	Schema       string
	Decision     Decision
	Reasons      []production.Reason
	Accepted     []string
	Observation  production.ObservationVector
	FullSuite    bool
	InputDigest  string
	ResultDigest string
	Bindings     productionBindingVector
}

type oracleBridgeVector struct {
	Schema                string
	InputDigest           string
	Decision              Decision
	Reason                Reason
	AcceptedSurfaces      []string
	ChangedSurfaces       []string
	ReceiptSurfaces       []string
	SemanticBeforeDigest  string
	SemanticAfterDigest   string
	SemanticDeltaDigest   string
	PathClosureDigest     string
	ObservationCounts     ObservationCounts
	Resources             ResourceObservation
	CanonicalOutputDigest string
	ReplayDigest          string
}

type bridgeSubjectVector struct {
	Oracle   oracleBridgeVector
	Producer productionVector
}

func oracleBridgeVectorFromResult(output Output) oracleBridgeVector {
	return oracleBridgeVector{
		Schema: output.Schema, InputDigest: output.InputDigest, Decision: output.Decision, Reason: output.Reason,
		AcceptedSurfaces: append([]string(nil), output.AcceptedSurfaces...), ChangedSurfaces: append([]string(nil), output.ChangedSurfaces...), ReceiptSurfaces: append([]string(nil), output.ReceiptSurfaces...),
		SemanticBeforeDigest: output.SemanticBeforeDigest, SemanticAfterDigest: output.SemanticAfterDigest, SemanticDeltaDigest: output.SemanticDeltaDigest, PathClosureDigest: output.PathClosureDigest,
		ObservationCounts: output.ObservationCounts, Resources: output.Resources, CanonicalOutputDigest: output.CanonicalOutputDigest, ReplayDigest: output.ReplayDigest,
	}
}
