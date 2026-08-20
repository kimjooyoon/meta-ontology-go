//go:build detector_bridge

package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
