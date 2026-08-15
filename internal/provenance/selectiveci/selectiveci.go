package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const SchemaVersion = "gooo-selective-ci-evidence/v1"

type DecisionStatus string

const (
	Verified   DecisionStatus = "VERIFIED"
	Unknown    DecisionStatus = "UNKNOWN"
	FailClosed DecisionStatus = "FAIL_CLOSED"
)

type FallbackMode string

const (
	NoFallback FallbackMode = "NONE"
	FullSuite  FallbackMode = "FULL_SUITE"
)

type ReceiptStatus string

const (
	ReceiptVerified  ReceiptStatus = "VERIFIED"
	ReceiptCandidate ReceiptStatus = "CANDIDATE"
	ReceiptDeferred  ReceiptStatus = "DEFERRED"
	ReceiptNotRun    ReceiptStatus = "NOT_RUN"
)

const (
	CodeVerified        = "SELECTIVE_CI_V1_VERIFIED"
	CodeMissing         = "SELECTIVE_CI_V1_MISSING_INPUT"
	CodeStaleSnapshot   = "SELECTIVE_CI_V1_STALE_SNAPSHOT"
	CodeCandidate       = "SELECTIVE_CI_V1_CANDIDATE_ONLY"
	CodeDuplicate       = "SELECTIVE_CI_V1_DUPLICATE"
	CodeAmbiguous       = "SELECTIVE_CI_V1_AMBIGUOUS"
	CodeDisconnected    = "SELECTIVE_CI_V1_DISCONNECTED"
	CodeWrongEndpoint   = "SELECTIVE_CI_V1_WRONG_ENDPOINT"
	CodeCycle           = "SELECTIVE_CI_V1_CYCLE"
	CodeReceiptMismatch = "SELECTIVE_CI_V1_RECEIPT_MISMATCH"
	CodeDigestMismatch  = "SELECTIVE_CI_V1_DIGEST_MISMATCH"
	CodeMalformed       = "SELECTIVE_CI_V1_MALFORMED"
)

type SnapshotBinding struct {
	Base semantic.SnapshotDigests
	Head semantic.SnapshotDigests
}

type Path struct {
	PathID        semantic.ID
	RootID        semantic.ID
	ObligationID  semantic.ID
	CommandID     semantic.ID
	ReceiptID     semantic.ID
	RecordIDs     []semantic.ID
	ExpectedKinds []semantic.InferenceKind
}

type CommandReceipt struct {
	CommandID             semantic.ID
	ReceiptID             semantic.ID
	Status                ReceiptStatus
	ProviderReceiptDigest string
	PhaseReceiptDigest    string
	ResourceReceiptDigest string
	RegistryDigest        string
	PlanDigest            string
	Digest                string
}

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

func (r Receipt) canonical() string {
	copy := canonicalReceipt(r)
	copy.Digest = ""
	data, err := marshalReceipt(copy)
	if err != nil {
		return ""
	}
	return string(data)
}

func (r Receipt) expectedDigest() string {
	return semantic.StableHashString("selective-ci-receipt/v1\x00" + r.canonical())
}

// Canonical returns the deterministic receipt payload without its self-digest.
func (r Receipt) Canonical() string { return r.canonical() }

// ExpectedDigest returns the digest that seals the canonical receipt payload.
func (r Receipt) ExpectedDigest() string { return r.expectedDigest() }

func canonicalReceipt(value Receipt) Receipt {
	copy := value
	copy.SelectedCommandIDs = append([]semantic.ID(nil), value.SelectedCommandIDs...)
	copy.ObligationIDs = append([]semantic.ID(nil), value.ObligationIDs...)
	copy.PathIDs = append([]semantic.ID(nil), value.PathIDs...)
	copy.VerifiedCommandIDs = append([]semantic.ID(nil), value.VerifiedCommandIDs...)
	copy.VerifiedObligationIDs = append([]semantic.ID(nil), value.VerifiedObligationIDs...)
	copy.VerifiedPathIDs = append([]semantic.ID(nil), value.VerifiedPathIDs...)
	sort.Slice(copy.SelectedCommandIDs, func(i, j int) bool { return copy.SelectedCommandIDs[i] < copy.SelectedCommandIDs[j] })
	sort.Slice(copy.ObligationIDs, func(i, j int) bool { return copy.ObligationIDs[i] < copy.ObligationIDs[j] })
	sort.Slice(copy.PathIDs, func(i, j int) bool { return copy.PathIDs[i] < copy.PathIDs[j] })
	sort.Slice(copy.VerifiedCommandIDs, func(i, j int) bool { return copy.VerifiedCommandIDs[i] < copy.VerifiedCommandIDs[j] })
	sort.Slice(copy.VerifiedObligationIDs, func(i, j int) bool { return copy.VerifiedObligationIDs[i] < copy.VerifiedObligationIDs[j] })
	sort.Slice(copy.VerifiedPathIDs, func(i, j int) bool { return copy.VerifiedPathIDs[i] < copy.VerifiedPathIDs[j] })
	return copy
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && strings.ToLower(value) == value
}

func normalizeDigest(value, label string) (string, error) {
	value = strings.TrimSpace(value)
	if !validDigest(value) {
		return "", fmt.Errorf("%s must be a lowercase SHA-256 digest", label)
	}
	return value, nil
}

func normalizeID(value semantic.ID, label string) (semantic.ID, error) {
	id, err := semantic.ParseIdentity(value.String())
	if err != nil {
		return "", fmt.Errorf("%s: %w", label, err)
	}
	return id, nil
}

func normalizeIDs(values []semantic.ID, label string) ([]semantic.ID, error) {
	if len(values) == 0 {
		return []semantic.ID{}, nil
	}
	out := make([]semantic.ID, 0, len(values))
	seen := make(map[semantic.ID]struct{}, len(values))
	for _, value := range values {
		id, err := normalizeID(value, label)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate %s %s", label, id)
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func equalIDs(left, right []semantic.ID) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func containsID(values []semantic.ID, value semantic.ID) bool {
	index := sort.Search(len(values), func(i int) bool { return values[i] >= value })
	return index < len(values) && values[index] == value
}
