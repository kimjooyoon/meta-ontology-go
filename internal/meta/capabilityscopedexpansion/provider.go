package capabilityscopedexpansion

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/broker"
)

// ProviderRequest is the complete live-input boundary. Repository and sandbox
// roots are intentionally separate: a sandbox observation is not a repository
// write observation.
type ProviderRequest struct {
	SubjectSHA     string
	PinnedFile     string
	LogicalInput   string
	SandboxRoot    string
	RepositoryRoot string
}

// CaptureProvider creates raw observations in a provider command. No producer
// package is involved in creating this wire artifact.
func CaptureProvider(request ProviderRequest) ([]byte, error) {
	if request.SubjectSHA == "" || request.PinnedFile == "" || request.LogicalInput == "" || request.SandboxRoot == "" || request.RepositoryRoot == "" {
		return nil, fmt.Errorf("provider requires subject, pinned file, logical input, sandbox root, and repository root")
	}
	repositoryBefore, err := snapshotRepository(request.RepositoryRoot)
	if err != nil {
		return nil, fmt.Errorf("repository before snapshot: %w", err)
	}
	sandboxBefore, err := snapshotSandbox(request.SandboxRoot)
	if err != nil {
		return nil, fmt.Errorf("sandbox before snapshot: %w", err)
	}
	contents, err := os.ReadFile(request.PinnedFile)
	if err != nil {
		return nil, fmt.Errorf("pinned file observation: %w", err)
	}
	logicalContents, err := os.ReadFile(request.LogicalInput)
	if err != nil {
		return nil, fmt.Errorf("logical input observation: %w", err)
	}
	logicalValue := strings.TrimSpace(string(logicalContents))
	if logicalValue != "logical-clock:0" {
		return nil, fmt.Errorf("logical input is not deterministic logical-clock:0")
	}

	policy := broker.Policy{ID: "default-deny", DefaultDecision: DecisionDeny, AuthorizationMode: PolicyExactCurrent, Effects: EffectPolicyNone}
	requests := []broker.Request{
		{Kind: "file", Operation: "write", Target: "repository", PolicyID: policy.ID},
		{Kind: "mutation", Operation: "mutate", Target: "sandbox", PolicyID: policy.ID},
		{Kind: "promotion", Operation: "promote", Target: "repository", PolicyID: policy.ID},
	}
	tokenAttempts := make([]TokenIssuance, 0, len(requests))
	issued, denied := 0, 0
	for _, effectRequest := range requests {
		_, issuance := broker.IssueToken(policy, effectRequest)
		if issuance.Issued {
			issued++
		} else {
			denied++
		}
		tokenAttempts = append(tokenAttempts, TokenIssuance{
			Kind: effectRequest.Kind, Operation: effectRequest.Operation, Target: effectRequest.Target,
			Requested: true, Decision: issuance.Decision, Issued: issuance.Issued, Reason: issuance.Reason,
			PolicyDigest: issuance.PolicyDigest, RequestDigest: issuance.RequestDigest,
		})
	}

	repositoryAfter, err := snapshotRepository(request.RepositoryRoot)
	if err != nil {
		return nil, fmt.Errorf("repository after snapshot: %w", err)
	}
	sandboxAfter, err := snapshotSandbox(request.SandboxRoot)
	if err != nil {
		return nil, fmt.Errorf("sandbox after snapshot: %w", err)
	}
	observation := ProviderObservation{
		Schema: ProviderSchema, Provider: "capabilityscopedexpansion.provider.Command", SubjectSHA: request.SubjectSHA,
		FileReads:        []FileReadObservation{{Target: "pinned-file", Path: request.PinnedFile, ContentDigest: digestBytes(contents), Observed: true, EvidenceClass: CurrentEvidence}},
		LogicalInputs:    []LogicalObservation{{Target: "logical-clock", Path: request.LogicalInput, Value: logicalValue, Observed: true, EvidenceClass: CurrentEvidence}},
		EnvironmentReads: []EnvironmentObservation{{Target: "GOOO_EXPANSION_PROFILE", Observed: false, EvidenceClass: "UNKNOWN"}},
		NetworkReads:     []NetworkObservation{{Target: "https://example.invalid/gooo/pinned-schema", Observed: false, EvidenceClass: HistoricalFixture}},
		TokenAttempts:    tokenAttempts, BrokerTokenRequests: len(requests), BrokerTokensIssued: issued, BrokerTokenDenials: denied,
		RepositoryBefore: repositoryBefore, RepositoryAfter: repositoryAfter, SandboxBefore: sandboxBefore, SandboxAfter: sandboxAfter,
		RepositoryWrites: snapshotDelta(repositoryBefore, repositoryAfter), SandboxWrites: snapshotDelta(sandboxBefore, sandboxAfter),
		MutationAuthority: "NOT_OBSERVED", PromotionAuthority: "NOT_OBSERVED", EffectAPIAccess: "NOT_REACHED_WITHOUT_TOKEN",
	}
	if err := validateProviderObservation(observation); err != nil {
		return nil, err
	}
	return json.MarshalIndent(observation, "", "  ")
}

func snapshotDelta(before, after SnapshotObservation) int {
	if before.Digest == after.Digest {
		return 0
	}
	return 1
}

func snapshotRepository(root string) (SnapshotObservation, error) {
	command := exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	raw, err := command.Output()
	if err != nil {
		return SnapshotObservation{}, fmt.Errorf("list repository files: %w", err)
	}
	paths := strings.Split(string(raw), "\x00")
	entries := make([]string, 0, len(paths))
	for _, relative := range paths {
		if relative == "" {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return SnapshotObservation{}, fmt.Errorf("read repository snapshot entry %q: %w", relative, err)
		}
		entries = append(entries, filepath.ToSlash(relative)+"="+digestBytes(contents))
	}
	sort.Strings(entries)
	return snapshotObservation("repository", root, entries), nil
}

func snapshotSandbox(root string) (SnapshotObservation, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return SnapshotObservation{}, err
	}
	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		contents, err := os.ReadFile(path)
		if err != nil {
			return SnapshotObservation{}, err
		}
		values = append(values, entry.Name()+"="+digestBytes(contents))
	}
	sort.Strings(values)
	return snapshotObservation("sandbox", root, values), nil
}

func snapshotObservation(scope, root string, entries []string) SnapshotObservation {
	return SnapshotObservation{Scope: scope, Root: root, Entries: entries, Digest: digestBytes([]byte(strings.Join(entries, "\n")))}
}

func validateProviderObservation(observation ProviderObservation) error {
	if observation.Schema != ProviderSchema || observation.Provider == "" || observation.SubjectSHA == "" {
		return fmt.Errorf("provider observation identity is incomplete")
	}
	if len(observation.FileReads) != 1 || observation.FileReads[0].Target != "pinned-file" || observation.FileReads[0].Path == "" || observation.FileReads[0].ContentDigest == "" || !observation.FileReads[0].Observed || observation.FileReads[0].EvidenceClass != CurrentEvidence {
		return fmt.Errorf("pinned file observation is incomplete")
	}
	if len(observation.LogicalInputs) != 1 || observation.LogicalInputs[0].Target != "logical-clock" || observation.LogicalInputs[0].Path == "" || observation.LogicalInputs[0].Value != "logical-clock:0" || !observation.LogicalInputs[0].Observed || observation.LogicalInputs[0].EvidenceClass != CurrentEvidence {
		return fmt.Errorf("logical input observation is incomplete")
	}
	if len(observation.EnvironmentReads) != 1 || observation.EnvironmentReads[0].Target != "GOOO_EXPANSION_PROFILE" || observation.EnvironmentReads[0].Observed || observation.EnvironmentReads[0].EvidenceClass != "UNKNOWN" || len(observation.NetworkReads) != 1 || observation.NetworkReads[0].Target != "https://example.invalid/gooo/pinned-schema" || observation.NetworkReads[0].Observed || observation.NetworkReads[0].EvidenceClass != HistoricalFixture {
		return fmt.Errorf("unobserved environment or network must remain lower-resolution provider evidence")
	}
	if observation.RepositoryBefore.Scope != "repository" || observation.RepositoryAfter.Scope != "repository" || observation.SandboxBefore.Scope != "sandbox" || observation.SandboxAfter.Scope != "sandbox" {
		return fmt.Errorf("repository and sandbox snapshots must have distinct scopes")
	}
	if observation.RepositoryWrites != snapshotDelta(observation.RepositoryBefore, observation.RepositoryAfter) || observation.SandboxWrites != snapshotDelta(observation.SandboxBefore, observation.SandboxAfter) {
		return fmt.Errorf("snapshot deltas do not match reported writes")
	}
	if observation.MutationAuthority != "NOT_OBSERVED" || observation.PromotionAuthority != "NOT_OBSERVED" || observation.EffectAPIAccess != "NOT_REACHED_WITHOUT_TOKEN" {
		return fmt.Errorf("unmeasured authority must remain explicitly unobserved")
	}
	if len(observation.TokenAttempts) != FixedEffectTokenRequests || observation.BrokerTokenRequests != len(observation.TokenAttempts) || observation.BrokerTokensIssued+observation.BrokerTokenDenials != observation.BrokerTokenRequests {
		return fmt.Errorf("broker token denominator is incomplete")
	}
	for _, attempt := range observation.TokenAttempts {
		if !attempt.Requested || attempt.Decision != DecisionDeny || attempt.Issued || attempt.PolicyDigest == "" || attempt.RequestDigest == "" || attempt.Reason == "" {
			return fmt.Errorf("broker token request was not observed as denied: %s/%s", attempt.Kind, attempt.Operation)
		}
	}
	if observation.BrokerTokensIssued != 0 || observation.BrokerTokenDenials != FixedEffectTokenRequests {
		return fmt.Errorf("default-deny broker did not deny every effect token")
	}
	return nil
}

func decodeProvider(raw []byte) (ProviderObservation, error) {
	var observation ProviderObservation
	if err := json.Unmarshal(raw, &observation); err != nil {
		return ProviderObservation{}, err
	}
	if err := validateProviderObservation(observation); err != nil {
		return ProviderObservation{}, err
	}
	return observation, nil
}

func currentEvidenceFor(provider ProviderObservation, declaration CapabilityDeclaration) (Evidence, bool) {
	if declaration.EvidenceClass != CurrentEvidence {
		return Evidence{}, false
	}
	switch declaration.Kind {
	case "file":
		for _, observation := range provider.FileReads {
			if observation.Target == declaration.Target && observation.Observed && observation.EvidenceClass == CurrentEvidence {
				return Evidence{ValueID: declaration.ValueID, Observed: observation.ContentDigest, EvidenceClass: CurrentEvidence, EvidenceDigest: digestBytes([]byte(declaration.ValueID + "=" + observation.ContentDigest)), Provenance: "provider.file.read"}, true
			}
		}
	case "time":
		for _, observation := range provider.LogicalInputs {
			if observation.Target == declaration.Target && observation.Observed && observation.EvidenceClass == CurrentEvidence {
				return Evidence{ValueID: declaration.ValueID, Observed: observation.Value, EvidenceClass: CurrentEvidence, EvidenceDigest: digestBytes([]byte(declaration.ValueID + "=" + observation.Value)), Provenance: "provider.logical.input"}, true
			}
		}
	}
	return Evidence{}, false
}

func providerDigest(raw []byte) string { return digestBytes(raw) }
