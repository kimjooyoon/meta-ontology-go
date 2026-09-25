package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

type Input struct {
	Schema             string
	Snapshots          SnapshotBinding
	RegistryDigest     string
	PlanDigest         string
	ChangedRootIDs     []semantic.ID
	SelectedCommandIDs []semantic.ID
	ObligationIDs      []semantic.ID
	Paths              []Path
	CommandReceipts    []CommandReceipt
	EvidenceIDs        []semantic.ID
	InferencePath      semantic.InferencePathV1
}
type Receipt struct {
	Schema                  string
	Status                  DecisionStatus
	Fallback                FallbackMode
	Code                    string
	Snapshots               SnapshotBinding
	RegistryDigest          string
	PlanDigest              string
	SelectedCommandIDs      []semantic.ID
	ObligationIDs           []semantic.ID
	PathIDs                 []semantic.ID
	RequiredCommandCount    int
	RequiredObligationCount int
	VerifiedCommandCount    int
	VerifiedObligationCount int
	VerifiedPathCount       int
	VerifiedCommandIDs      []semantic.ID
	VerifiedObligationIDs   []semantic.ID
	VerifiedPathIDs         []semantic.ID
	Digest                  string
}

func (s SnapshotBinding) equal(other SnapshotBinding) bool {
	return s.Base == other.Base && s.Head == other.Head
}
func (s SnapshotBinding) canonical() string {
	return strings.Join([]string{
		s.Base.Source, s.Base.Semantic, s.Head.Source, s.Head.Semantic,
	}, "\x00")
}
func (r CommandReceipt) canonical(binding SnapshotBinding) string {
	return strings.Join([]string{
		r.CommandID.String(), r.ReceiptID.String(), string(r.Status),
		r.ProviderReceiptDigest, r.PhaseReceiptDigest, r.ResourceReceiptDigest,
		r.RegistryDigest, r.PlanDigest, binding.canonical(),
	}, "\x00")
}
func (r CommandReceipt) expectedDigest(binding SnapshotBinding) string {
	return semantic.StableHashString("selective-ci-command-receipt/v1\x00" + r.canonical(binding))
}

// ExpectedDigest returns the digest bound to the immutable command receipt
// fields and the caller-supplied snapshot binding.
func (r CommandReceipt) ExpectedDigest(binding SnapshotBinding) string {
	return r.expectedDigest(binding)
}
