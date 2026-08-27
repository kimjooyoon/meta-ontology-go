package capabilityscopedexpansion

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProviderRequest describes the only live inputs used by this experiment. The
// provider reads a CI-created pinned file and a deterministic logical input;
// it deliberately does not read wall-clock time, environment, or network.
type ProviderRequest struct {
	SubjectSHA  string
	PinnedFile  string
	SandboxRoot string
}

// CaptureProvider performs the bounded observation and records the raw wire
// evidence consumed by both the producer and the independent consumer.
func CaptureProvider(request ProviderRequest) ([]byte, error) {
	if request.SubjectSHA == "" || request.PinnedFile == "" || request.SandboxRoot == "" {
		return nil, fmt.Errorf("provider requires subject, pinned file, and sandbox root")
	}
	before, err := snapshot(request.SandboxRoot)
	if err != nil {
		return nil, err
	}
	contents, err := os.ReadFile(request.PinnedFile)
	if err != nil {
		return nil, fmt.Errorf("pinned file observation: %w", err)
	}
	provider := sandboxProvider{root: request.SandboxRoot}
	effects := make([]EffectObservation, 0, 3)
	for _, requestEffect := range []func(string) (EffectObservation, error){provider.requestWrite, provider.requestMutation, provider.requestPromotion} {
		effect, err := requestEffect("capability-boundary-probe")
		if err != nil {
			return nil, err
		}
		effects = append(effects, effect)
	}
	after, err := snapshot(request.SandboxRoot)
	if err != nil {
		return nil, err
	}
	actualWrites := 0
	if before.Digest != after.Digest {
		actualWrites = 1
	}
	actualMutation := false
	actualPromotion := false
	for _, effect := range effects {
		actualMutation = actualMutation || effect.ActualMutation
		actualPromotion = actualPromotion || effect.ActualPromotion
	}
	observation := ProviderObservation{
		Schema:                   ProviderSchema,
		Provider:                 "capabilityscopedexpansion.SandboxProvider",
		SubjectSHA:               request.SubjectSHA,
		FileReads:                []FileReadObservation{{Target: "pinned-file", Path: request.PinnedFile, ContentDigest: digestBytes(contents), Observed: true, EvidenceClass: CurrentEvidence}},
		LogicalInputs:            []LogicalObservation{{Target: "logical-clock", Value: "logical-clock:0", Observed: true, EvidenceClass: CurrentEvidence}},
		EnvironmentReads:         []EnvironmentObservation{{Target: "GOOO_EXPANSION_PROFILE", Observed: false, EvidenceClass: "UNKNOWN"}},
		NetworkReads:             []NetworkObservation{{Target: "https://example.invalid/gooo/pinned-schema", Observed: false, EvidenceClass: HistoricalFixture}},
		EffectAttempts:           effects,
		SandboxBefore:            before,
		SandboxAfter:             after,
		ActualRepositoryWrites:   actualWrites,
		ActualMutationAuthority:  actualMutation,
		ActualPromotionAuthority: actualPromotion,
	}
	if err := validateProviderObservation(observation); err != nil {
		return nil, err
	}
	return json.MarshalIndent(observation, "", "  ")
}

type sandboxProvider struct {
	root string
}

func (provider sandboxProvider) requestWrite(target string) (EffectObservation, error) {
	before, err := snapshot(provider.root)
	if err != nil {
		return EffectObservation{}, err
	}
	// The provider API has no write token. Returning the boundary result without
	// touching the filesystem is the observed enforcement point.
	after, err := snapshot(provider.root)
	if err != nil {
		return EffectObservation{}, err
	}
	return effectObservation("repository-write", target, before, after), nil
}

func (provider sandboxProvider) requestMutation(target string) (EffectObservation, error) {
	before, err := snapshot(provider.root)
	if err != nil {
		return EffectObservation{}, err
	}
	after, err := snapshot(provider.root)
	if err != nil {
		return EffectObservation{}, err
	}
	return effectObservation("mutation-authority", target, before, after), nil
}

func (provider sandboxProvider) requestPromotion(target string) (EffectObservation, error) {
	before, err := snapshot(provider.root)
	if err != nil {
		return EffectObservation{}, err
	}
	after, err := snapshot(provider.root)
	if err != nil {
		return EffectObservation{}, err
	}
	return effectObservation("promotion-authority", target, before, after), nil
}

func effectObservation(kind, target string, before, after SnapshotObservation) EffectObservation {
	changed := before.Digest != after.Digest
	actualWrites := 0
	if kind == "repository-write" && changed {
		actualWrites = 1
	}
	return EffectObservation{Kind: kind, Target: target, Requested: true, Result: "DENIED", Reason: "CAPABILITY_ENFORCEMENT_OBSERVED", BoundaryObserved: true, BeforeDigest: before.Digest, AfterDigest: after.Digest, ActualWrites: actualWrites, ActualMutation: kind == "mutation-authority" && changed, ActualPromotion: kind == "promotion-authority" && changed}
}

func snapshot(root string) (SnapshotObservation, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return SnapshotObservation{}, err
	}
	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			continue
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return SnapshotObservation{}, err
		}
		values = append(values, entry.Name()+"="+digestBytes(contents))
	}
	sort.Strings(values)
	return SnapshotObservation{Root: root, Entries: values, Digest: digestBytes([]byte(strings.Join(values, "\n")))}, nil
}

func validateProviderObservation(observation ProviderObservation) error {
	if observation.Schema != ProviderSchema || observation.Provider == "" || observation.SubjectSHA == "" {
		return fmt.Errorf("provider observation identity is incomplete")
	}
	if len(observation.FileReads) != 1 || observation.FileReads[0].Target != "pinned-file" || observation.FileReads[0].Path == "" || observation.FileReads[0].ContentDigest == "" || !observation.FileReads[0].Observed || observation.FileReads[0].EvidenceClass != CurrentEvidence {
		return fmt.Errorf("pinned file observation is incomplete")
	}
	if len(observation.LogicalInputs) != 1 || observation.LogicalInputs[0].Target != "logical-clock" || observation.LogicalInputs[0].Value != "logical-clock:0" || !observation.LogicalInputs[0].Observed || observation.LogicalInputs[0].EvidenceClass != CurrentEvidence {
		return fmt.Errorf("logical input observation is incomplete")
	}
	if len(observation.EnvironmentReads) != 1 || observation.EnvironmentReads[0].Target != "GOOO_EXPANSION_PROFILE" || observation.EnvironmentReads[0].Observed || observation.EnvironmentReads[0].EvidenceClass != "UNKNOWN" || len(observation.NetworkReads) != 1 || observation.NetworkReads[0].Target != "https://example.invalid/gooo/pinned-schema" || observation.NetworkReads[0].Observed || observation.NetworkReads[0].EvidenceClass != HistoricalFixture {
		return fmt.Errorf("unobserved environment or network must remain lower-resolution provider evidence")
	}
	if observation.SandboxBefore.Digest != observation.SandboxAfter.Digest || observation.ActualRepositoryWrites != 0 || observation.ActualMutationAuthority || observation.ActualPromotionAuthority {
		return fmt.Errorf("sandbox before/after observation shows an unexpected effect")
	}
	if len(observation.EffectAttempts) != 3 {
		return fmt.Errorf("effect enforcement denominator is incomplete")
	}
	for _, effect := range observation.EffectAttempts {
		if !effect.Requested || effect.Result != "DENIED" || !effect.BoundaryObserved || effect.BeforeDigest != effect.AfterDigest || effect.ActualWrites != 0 || effect.ActualMutation || effect.ActualPromotion {
			return fmt.Errorf("effect request was not observed as denied: %s", effect.Kind)
		}
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

func providerDigest(raw []byte) string {
	return digestBytes(raw)
}
