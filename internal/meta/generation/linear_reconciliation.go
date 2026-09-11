package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"
)

// Linear reconciliation is a read-only proof for the ordinary main-target
// promotion route. It deliberately models observations rather than mutating
// refs, protection, or the historical ledger.
const (
	ReconciliationSchema           = "gooo/linear-tree-reconciliation/v1"
	ReconciliationRepository       = "kimjooyoon/meta-ontology-go"
	ReconciliationOwnerLogin       = "kimjooyoon"
	ReconciliationOwnerID    int64 = 115961382
	ReconciliationOwnerType        = "User"
)

var reconciliationSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

var ReconciliationUnknownFields = [...]string{
	"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by",
}

type ReconciliationDecision string

const (
	ReconciliationClosed  ReconciliationDecision = "CLOSED"
	ReconciliationUnknown ReconciliationDecision = "UNKNOWN"
	ReconciliationRefuted ReconciliationDecision = "REFUTED"
)

type ReconciliationCandidate struct {
	PullRequest        int    `json:"pull_request"`
	Repository         string `json:"repository"`
	BaseBranch         string `json:"base_branch"`
	BaseSHA            string `json:"base_sha"`
	HeadBranch         string `json:"head_branch"`
	HeadSHA            string `json:"head_sha"`
	MergeBaseSHA       string `json:"merge_base_sha"`
	ReplayIdentity     string `json:"replay_identity"`
	ExpectedTreeDigest string `json:"expected_tree_digest"`
}

type ReconciliationTopology struct {
	MainBeforeSHA string `json:"main_before_sha"`
	MainAfterSHA  string `json:"main_after_sha"`
	DevBeforeSHA  string `json:"dev_before_sha"`
	DevAfterSHA   string `json:"dev_after_sha"`
	MergeBaseSHA  string `json:"merge_base_sha"`
	Status        string `json:"status"`
	AheadBy       int    `json:"ahead_by"`
	BehindBy      int    `json:"behind_by"`
}

type ReconciliationTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
}

type ReconciliationTreeEvidence struct {
	SourceBefore *ReconciliationTree `json:"source_before"`
	SourceAfter  *ReconciliationTree `json:"source_after"`
	Reconciled   *ReconciliationTree `json:"reconciled"`
}

type ReconciliationTree []ReconciliationTreeEntry

type ReconciliationOwner struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Type  string `json:"type"`
}

type ReconciliationAuthorization struct {
	State                  string                  `json:"state"`
	TargetBranch           string                  `json:"target_branch"`
	OwnerSelection         ReconciliationOwner     `json:"owner_selection"`
	Actor                  ReconciliationOwner     `json:"actor"`
	Candidate              ReconciliationCandidate `json:"candidate"`
	CandidateDigest        string                  `json:"candidate_digest"`
	Nonce                  string                  `json:"nonce"`
	IssuedAt               string                  `json:"issued_at"`
	ExpiresAt              string                  `json:"expires_at"`
	OneUse                 bool                    `json:"one_use"`
	Reusable               bool                    `json:"reusable"`
	UseCount               int                     `json:"use_count"`
	ReuseAttempts          int                     `json:"reuse_attempts"`
	ProtectionMutation     bool                    `json:"protection_mutation"`
	ForcePush              bool                    `json:"force_push"`
	RepositoryWritesBefore int                     `json:"repository_writes_before"`
}

type ReconciliationReplay struct {
	Identity              string `json:"identity"`
	CurrentRequestDigest  string `json:"current_request_digest"`
	PreviousIdentity      string `json:"previous_identity"`
	PreviousRequestDigest string `json:"previous_request_digest"`
}

type ReconciliationMutationEvidence struct {
	RepositoryWrites    int `json:"repository_writes"`
	ProtectionMutations int `json:"protection_mutations"`
	ForcePushes         int `json:"force_pushes"`
}

type ReconciliationInput struct {
	Candidate     ReconciliationCandidate        `json:"candidate"`
	Topology      ReconciliationTopology         `json:"topology"`
	Trees         ReconciliationTreeEvidence     `json:"trees"`
	Authorization *ReconciliationAuthorization   `json:"authorization"`
	Replay        *ReconciliationReplay          `json:"replay"`
	Mutations     ReconciliationMutationEvidence `json:"mutations"`
	Now           string                         `json:"now"`
}

type ReconciliationUnknownEvidence struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type ReconciliationRefutedEvidence struct {
	Reason string `json:"reason"`
	CellID string `json:"cell_id"`
	Detail string `json:"detail"`
}

type ReconciliationCell struct {
	ID    string                 `json:"id"`
	State ReconciliationDecision `json:"state"`
}

type ReconciliationMetrics struct {
	OperationRequests   int `json:"operation_requests"`
	OperationResults    int `json:"operation_results"`
	ReplayComparisons   int `json:"replay_comparisons"`
	ReplayMismatches    int `json:"replay_mismatches"`
	RepositoryWrites    int `json:"repository_writes"`
	ProtectionMutations int `json:"protection_mutations"`
	ForcePushes         int `json:"force_pushes"`
}

type ReconciliationReceipt struct {
	Schema          string                          `json:"schema"`
	Decision        ReconciliationDecision          `json:"decision"`
	Reason          string                          `json:"reason"`
	CandidateDigest string                          `json:"candidate_digest"`
	TreeDigest      string                          `json:"tree_digest"`
	ReplayDigest    string                          `json:"replay_digest"`
	Cells           []ReconciliationCell            `json:"cells"`
	Unknown         *ReconciliationUnknownEvidence  `json:"unknown"`
	Refuted         []ReconciliationRefutedEvidence `json:"refuted"`
	Metrics         ReconciliationMetrics           `json:"metrics"`
}

// ReconciliationOwnerIdentity is the only identity accepted by the normal
// main-target owner authorization route.
func ReconciliationOwnerIdentity() ReconciliationOwner {
	return ReconciliationOwner{Login: ReconciliationOwnerLogin, ID: ReconciliationOwnerID, Type: ReconciliationOwnerType}
}

func validReconciliationSHA(value string) bool {
	return reconciliationSHA.MatchString(value)
}

func validReconciliationDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func reconciliationDigest(value any) string {
	encoded, _ := json.Marshal(value)
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:])
}

// ReconciliationTreeDigest canonicalizes the complete recursive blob tree.
// A missing or malformed tree is an error; an empty valid tree is distinct
// from missing evidence.
func ReconciliationTreeDigest(tree ReconciliationTree) (string, error) {
	entries := append(ReconciliationTree(nil), tree...)
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].Path < entries[right].Path
	})
	canonical := make([]struct {
		Path string `json:"path"`
		Mode string `json:"mode"`
		Type string `json:"type"`
		SHA  string `json:"sha"`
	}, 0, len(entries))
	for index, entry := range entries {
		if entry.Path == "" || entry.Type != "blob" || entry.Mode == "" || !validReconciliationSHA(entry.SHA) {
			return "", fmt.Errorf("malformed tree entry at index %d", index)
		}
		if index > 0 && entries[index-1].Path == entry.Path {
			return "", fmt.Errorf("duplicate tree path %q", entry.Path)
		}
		canonical = append(canonical, struct {
			Path string `json:"path"`
			Mode string `json:"mode"`
			Type string `json:"type"`
			SHA  string `json:"sha"`
		}{entry.Path, entry.Mode, entry.Type, entry.SHA})
	}
	return reconciliationDigest(canonical), nil
}

func ReconciliationCandidateDigest(candidate ReconciliationCandidate) string {
	return reconciliationDigest(candidate)
}

func ReconciliationReplayDigest(replay ReconciliationReplay) string {
	return reconciliationDigest(struct {
		Identity string `json:"identity"`
		Current  string `json:"current_request_digest"`
		Previous string `json:"previous_request_digest"`
	}{replay.Identity, replay.CurrentRequestDigest, replay.PreviousRequestDigest})
}

func reconciliationUnknown(stage, step, reason, class, next string, blocked ...string) *ReconciliationUnknownEvidence {
	return &ReconciliationUnknownEvidence{
		Stage: stage, Step: step, Reason: reason, UnknownClass: class,
		NextOperation: next, BlockedBy: append([]string(nil), blocked...),
	}
}

func (input ReconciliationInput) now() (time.Time, error) {
	if input.Now == "" {
		return time.Time{}, fmt.Errorf("evaluation time is missing")
	}
	return time.Parse(time.RFC3339Nano, input.Now)
}

func validReconciliationOwner(owner ReconciliationOwner) bool {
	return owner == ReconciliationOwnerIdentity()
}

func validateReconciliationAuthorization(auth ReconciliationAuthorization, candidate ReconciliationCandidate, now time.Time) (string, bool) {
	if auth.State != "AUTHORIZED" || auth.TargetBranch != "main" || !validReconciliationOwner(auth.OwnerSelection) || !validReconciliationOwner(auth.Actor) || auth.Actor != auth.OwnerSelection || auth.Candidate != candidate || auth.CandidateDigest != ReconciliationCandidateDigest(candidate) || len(auth.Nonce) < 16 || auth.OneUse != true || auth.Reusable != false || auth.UseCount != 0 || auth.ReuseAttempts != 0 || auth.ProtectionMutation || auth.ForcePush || auth.RepositoryWritesBefore != 0 {
		return "OWNER_AUTHORIZATION_MISMATCH", false
	}
	issued, issuedErr := time.Parse(time.RFC3339Nano, auth.IssuedAt)
	expires, expiresErr := time.Parse(time.RFC3339Nano, auth.ExpiresAt)
	if issuedErr != nil || expiresErr != nil || !issued.Before(now) || !expires.After(now) {
		return "AUTHORIZATION_NOT_FRESH", false
	}
	return "", true
}

func validReconciliationTopology(topology ReconciliationTopology, candidate ReconciliationCandidate) (unknownReason string, refutedReason string) {
	if !validReconciliationSHA(topology.MainBeforeSHA) || !validReconciliationSHA(topology.MainAfterSHA) || !validReconciliationSHA(topology.DevBeforeSHA) || !validReconciliationSHA(topology.DevAfterSHA) || !validReconciliationSHA(topology.MergeBaseSHA) || topology.Status == "" {
		return "INCOMPLETE_TOPOLOGY", ""
	}
	if topology.MainBeforeSHA != candidate.BaseSHA || topology.DevBeforeSHA != candidate.HeadSHA || topology.MainAfterSHA != candidate.BaseSHA || topology.DevAfterSHA != candidate.HeadSHA || topology.MergeBaseSHA != candidate.BaseSHA || topology.MainBeforeSHA != topology.MainAfterSHA || topology.DevBeforeSHA != topology.DevAfterSHA {
		return "", "TOPOLOGY_REF_DRIFT"
	}
	if topology.Status != "ahead" || topology.AheadBy <= 0 || topology.BehindBy != 0 {
		return "", "NON_LINEAR_TOPOLOGY"
	}
	return "", ""
}

func validateReconciliationCandidate(candidate ReconciliationCandidate) (string, bool) {
	if candidate.PullRequest <= 0 || candidate.Repository == "" || candidate.BaseBranch == "" || candidate.HeadBranch == "" || candidate.ReplayIdentity == "" || !validReconciliationSHA(candidate.BaseSHA) || !validReconciliationSHA(candidate.HeadSHA) || !validReconciliationSHA(candidate.MergeBaseSHA) || !validReconciliationDigest(candidate.ExpectedTreeDigest) {
		return "INCOMPLETE_CANDIDATE", false
	}
	if candidate.Repository != ReconciliationRepository || candidate.BaseBranch != "main" || candidate.HeadBranch != "dev" {
		return "CANDIDATE_ROUTE_MISMATCH", false
	}
	return "", true
}

func addReconciliationCell(cells []ReconciliationCell, id string, state ReconciliationDecision) []ReconciliationCell {
	for index := range cells {
		if cells[index].ID == id {
			cells[index].State = state
			return cells
		}
	}
	return append(cells, ReconciliationCell{ID: id, State: state})
}

func ReconcileLinearTree(input ReconciliationInput) ReconciliationReceipt {
	cells := []ReconciliationCell{
		{ID: "CANDIDATE_IDENTITY", State: ReconciliationClosed},
		{ID: "LINEAR_TOPOLOGY", State: ReconciliationClosed},
		{ID: "SOURCE_TREE_STABILITY", State: ReconciliationClosed},
		{ID: "EXACT_TREE_EQUIVALENCE", State: ReconciliationClosed},
		{ID: "FRESH_OWNER_AUTHORIZATION", State: ReconciliationClosed},
		{ID: "REPLAY_IDENTITY", State: ReconciliationClosed},
		{ID: "NO_MUTATING_EFFECT", State: ReconciliationClosed},
	}
	unknowns := make([]ReconciliationUnknownEvidence, 0, 4)
	refuted := make([]ReconciliationRefutedEvidence, 0, 4)
	markUnknown := func(id string, evidence *ReconciliationUnknownEvidence) {
		unknowns = append(unknowns, *evidence)
		cells = addReconciliationCell(cells, id, ReconciliationUnknown)
	}
	markRefuted := func(id, reason, detail string) {
		refuted = append(refuted, ReconciliationRefutedEvidence{Reason: reason, CellID: id, Detail: detail})
		cells = addReconciliationCell(cells, id, ReconciliationRefuted)
	}

	candidateDigest := ReconciliationCandidateDigest(input.Candidate)
	if reason, valid := validateReconciliationCandidate(input.Candidate); !valid {
		if reason == "INCOMPLETE_CANDIDATE" {
			markUnknown("CANDIDATE_IDENTITY", reconciliationUnknown("FOUNDATION", "BIND_CANDIDATE", reason, "INCOMPLETE_EVIDENCE", "PROVIDE_EXACT_CANDIDATE", "candidate"))
		} else {
			markRefuted("CANDIDATE_IDENTITY", reason, "candidate is not the exact same-repository main-target route")
		}
	}

	if topologyUnknown, topologyRefuted := validReconciliationTopology(input.Topology, input.Candidate); topologyUnknown != "" {
		markUnknown("LINEAR_TOPOLOGY", reconciliationUnknown("FOUNDATION", "READ_LIVE_REFS", topologyUnknown, "INCOMPLETE_EVIDENCE", "READ_CURRENT_MAIN_AND_DEV"))
	} else if topologyRefuted != "" {
		markRefuted("LINEAR_TOPOLOGY", topologyRefuted, "main and dev observations are not a stable fast-forward topology")
	}

	treeDigest := ""
	if input.Trees.SourceBefore == nil || input.Trees.SourceAfter == nil || input.Trees.Reconciled == nil {
		markUnknown("SOURCE_TREE_STABILITY", reconciliationUnknown("COHERENCE", "READ_SOURCE_TREES", "TREE_OBSERVATION_MISSING", "DIRECT_MISSING", "READ_COMPLETE_RECURSIVE_TREES", "tree-observation"))
	} else {
		beforeDigest, beforeErr := ReconciliationTreeDigest(*input.Trees.SourceBefore)
		afterDigest, afterErr := ReconciliationTreeDigest(*input.Trees.SourceAfter)
		reconciledDigest, reconciledErr := ReconciliationTreeDigest(*input.Trees.Reconciled)
		if beforeErr != nil || afterErr != nil || reconciledErr != nil {
			markRefuted("SOURCE_TREE_STABILITY", "MALFORMED_TREE", "recursive tree evidence is malformed")
		} else {
			treeDigest = afterDigest
			if beforeDigest != afterDigest {
				markRefuted("SOURCE_TREE_STABILITY", "SOURCE_TREE_DRIFT", "source tree changed between exact observations")
			}
			if afterDigest != reconciledDigest || afterDigest != input.Candidate.ExpectedTreeDigest {
				markRefuted("EXACT_TREE_EQUIVALENCE", "TREE_MISMATCH", "reconciled tree is not byte-for-byte equivalent to current dev")
			}
		}
	}

	if input.Authorization == nil {
		markUnknown("FRESH_OWNER_AUTHORIZATION", reconciliationUnknown("FOUNDATION", "READ_OWNER_AUTHORIZATION", "OWNER_AUTHORIZATION_MISSING", "DIRECT_MISSING", "OBTAIN_FRESH_MAIN_TARGET_AUTHORIZATION", "owner-authorization"))
	} else if now, err := input.now(); err != nil {
		markUnknown("FRESH_OWNER_AUTHORIZATION", reconciliationUnknown("FOUNDATION", "READ_OWNER_AUTHORIZATION", "EVALUATION_TIME_MISSING", "INCOMPLETE_EVIDENCE", "PROVIDE_EVALUATION_TIME", "clock"))
	} else if reason, valid := validateReconciliationAuthorization(*input.Authorization, input.Candidate, now); !valid {
		markRefuted("FRESH_OWNER_AUTHORIZATION", reason, "owner authorization is stale, consumed, or not exact for this main-target candidate")
	}

	if input.Replay == nil {
		markUnknown("REPLAY_IDENTITY", reconciliationUnknown("REGRESSION", "COMPARE_REPLAY", "REPLAY_OBSERVATION_MISSING", "DIRECT_MISSING", "PROVIDE_REPLAY_OBSERVATION", "replay"))
	} else if input.Replay.Identity != input.Candidate.ReplayIdentity || input.Replay.CurrentRequestDigest != candidateDigest || !validReconciliationDigest(input.Replay.CurrentRequestDigest) {
		markRefuted("REPLAY_IDENTITY", "REQUEST_DIGEST_MISMATCH", "replay identity or current request digest is not bound to the candidate")
	} else if input.Replay.PreviousIdentity != "" {
		if input.Replay.PreviousIdentity != input.Replay.Identity || input.Replay.PreviousRequestDigest != input.Replay.CurrentRequestDigest {
			markRefuted("REPLAY_IDENTITY", "REPLAY_COLLISION", "the same replay identity has a different canonical request digest")
		}
	}

	if input.Mutations.RepositoryWrites != 0 || input.Mutations.ProtectionMutations != 0 || input.Mutations.ForcePushes != 0 || input.Mutations.RepositoryWrites < 0 || input.Mutations.ProtectionMutations < 0 || input.Mutations.ForcePushes < 0 {
		markRefuted("NO_MUTATING_EFFECT", "MUTATING_EFFECT_OBSERVED", "reconciliation evidence contains a write, protection mutation, or force push")
	}

	metrics := ReconciliationMetrics{
		OperationRequests:   1,
		OperationResults:    1,
		RepositoryWrites:    input.Mutations.RepositoryWrites,
		ProtectionMutations: input.Mutations.ProtectionMutations,
		ForcePushes:         input.Mutations.ForcePushes,
	}
	if input.Replay != nil {
		metrics.ReplayComparisons = 1
		if len(refuted) > 0 {
			for _, failure := range refuted {
				if failure.CellID == "REPLAY_IDENTITY" && failure.Reason == "REPLAY_COLLISION" {
					metrics.ReplayMismatches = 1
				}
			}
		}
	}

	decision := ReconciliationClosed
	reason := "EXACT_LINEAR_TREE_RECONCILIATION_CLOSED"
	if len(unknowns) > 0 {
		decision = ReconciliationUnknown
		reason = unknowns[0].Reason
	}
	if len(refuted) > 0 {
		decision = ReconciliationRefuted
		reason = refuted[0].Reason
	}
	unknown := (*ReconciliationUnknownEvidence)(nil)
	if len(unknowns) > 0 {
		unknown = &unknowns[0]
	}
	return ReconciliationReceipt{
		Schema:          ReconciliationSchema,
		Decision:        decision,
		Reason:          reason,
		CandidateDigest: candidateDigest,
		TreeDigest:      treeDigest,
		ReplayDigest: func() string {
			if input.Replay == nil {
				return ""
			}
			return ReconciliationReplayDigest(*input.Replay)
		}(),
		Cells:   cells,
		Unknown: unknown,
		Refuted: refuted,
		Metrics: metrics,
	}
}
