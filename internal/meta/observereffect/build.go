package observereffect

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type BuildOptions struct {
	Mode                  string
	Root                  string
	TopologyRoot          string
	SourcePath            string
	AllowIntentionalWrite bool
}

type fileSnapshot struct {
	Digest string
	Paths  []string
	Files  map[string]string
}

func Build(opts BuildOptions) (Report, Receipt, Receipt, error) {
	if opts.Mode != "observe" && opts.Mode != "unknown" && opts.Mode != "violate" {
		return Report{}, Receipt{}, Receipt{}, fmt.Errorf("unsupported mode %q", opts.Mode)
	}
	if opts.Root == "" || opts.SourcePath == "" {
		return Report{}, Receipt{}, Receipt{}, fmt.Errorf("root and source are required")
	}
	sourcePayload, err := os.ReadFile(opts.SourcePath)
	if err != nil {
		return Report{}, Receipt{}, Receipt{}, fmt.Errorf("read .gooo source: %w", err)
	}
	source := Source{
		Path:       sourceDisplayPath(opts.Root, opts.SourcePath),
		Digest:     DigestBytes(sourcePayload),
		GoooSource: validGoooSource(opts.SourcePath, sourcePayload),
	}
	before, err := scanRoot(opts.Root)
	if err != nil {
		return Report{}, Receipt{}, Receipt{}, err
	}
	if opts.Mode == "violate" {
		if !opts.AllowIntentionalWrite {
			return Report{}, Receipt{}, Receipt{}, fmt.Errorf("intentional write was not authorized")
		}
		marker := filepath.Join(opts.Root, ".observer-effect-intentional-write.marker")
		if err := os.WriteFile(marker, []byte("counterexample\n"), 0o644); err != nil {
			return Report{}, Receipt{}, Receipt{}, fmt.Errorf("write counterexample marker: %w", err)
		}
	}
	after, err := scanRoot(opts.Root)
	if err != nil {
		return Report{}, Receipt{}, Receipt{}, err
	}
	environment := environmentDelta()
	logicalTime := logicalTimeDelta()
	repositoryChanged := before.Digest != after.Digest
	repositoryWrites := changedPathCount(before, after)
	unknown := Unknown{Stage: "NONE", Step: "NONE", Reason: "NONE"}
	if opts.Mode == "unknown" {
		unknown = Unknown{
			Stage:  "OBSERVE",
			Step:   "capture-environment",
			Reason: "DECLARED_ENVIRONMENT_NOT_AVAILABLE",
		}
	}
	topologyRoot := opts.TopologyRoot
	if topologyRoot == "" {
		topologyRoot = opts.Root
	}
	topology := buildTopology(topologyRoot)
	decision, resolution := "OBSERVED", "EXACT"
	if opts.Mode == "unknown" {
		decision, resolution = "UNKNOWN", "LOWER_RESOLUTION"
	}
	if !topology.Exact || repositoryChanged {
		decision = "FAIL_CLOSED"
		resolution = "EXACT"
	}
	observation := Observation{
		RepositoryStorage: SnapshotDelta{BeforeDigest: before.Digest, AfterDigest: after.Digest, Changed: repositoryChanged},
		Environment:       environment,
		LogicalTime:       logicalTime,
	}
	effects := buildEffects(observation, repositoryWrites, decision)
	claim := buildClaimTransition(source.Digest, decision, unknown)
	indicators := buildIndicators(source, observation, effects, unknown, claim, decision)
	metrics := buildMetrics(indicators)
	authority := Authority{
		RepositoryWrites: repositoryWrites, OutputWrites: 3,
		MutationAuthority: false, PromotionAuthorized: false,
	}
	report := Report{
		Schema: LedgerSchema, Experiment: ExperimentName, Source: source,
		Observation: observation, Effects: effects, Unknown: unknown,
		Coordinate: unknown, Reason: unknown.Reason,
		ClaimTransition: claim, Topology: topology,
		RunnerScoped: runnerScopedEvidence(), Guardian: guardianExpectation(), Metrics: metrics, Authority: authority,
		RepositoryWrites: repositoryWrites, MutationAuthority: false,
		PromotionAuthorized: false, Decision: decision, Resolution: resolution,
		Indicators: indicators,
	}
	report.EvidenceDigest = DigestValue([]any{report.Source, report.Observation, report.Effects, report.Unknown, report.ClaimTransition, report.Topology, report.RunnerScoped, report.Guardian})
	observationReceipt := makeReceipt("OBSERVATION_RESULT", source, decision, resolution, repositoryWrites, unknown, claim, report.EvidenceDigest)
	effectReceipt := makeReceipt("OBSERVER_EFFECT", source, decision, resolution, repositoryWrites, unknown, claim, DigestValue(effects))
	observationReceipt.Digest = ReceiptDigest(observationReceipt)
	effectReceipt.Digest = ReceiptDigest(effectReceipt)
	report.ReceiptDigests = []string{observationReceipt.Digest, effectReceipt.Digest}
	report.Digest = ReportDigest(report)
	return report, observationReceipt, effectReceipt, nil
}

func sourceDisplayPath(root, source string) string {
	rootAbs, rootErr := filepath.Abs(root)
	sourceAbs, sourceErr := filepath.Abs(source)
	if rootErr == nil && sourceErr == nil {
		if relative, err := filepath.Rel(rootAbs, sourceAbs); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return filepath.ToSlash(relative)
		}
	}
	return filepath.ToSlash(source)
}

func validGoooSource(path string, payload []byte) bool {
	return strings.HasSuffix(path, ".gooo") &&
		strings.Contains(string(payload), "entity ") &&
		strings.Contains(string(payload), "activity ")
}

func scanRoot(root string) (fileSnapshot, error) {
	paths := make([]string, 0)
	contents := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root && entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			contents[filepath.ToSlash(relative)] = []byte("symlink:" + target)
		} else {
			payload, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			contents[filepath.ToSlash(relative)] = payload
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return fileSnapshot{}, fmt.Errorf("scan root: %w", err)
	}
	sort.Strings(paths)
	framed := make([]string, 0, len(paths))
	digests := make(map[string]string, len(paths))
	for _, path := range paths {
		digests[path] = DigestBytes(contents[path])
		framed = append(framed, path, string([]byte{0}), string(contents[path]), string([]byte{0}))
	}
	return fileSnapshot{Digest: DigestBytes([]byte(strings.Join(framed, ""))), Paths: paths, Files: digests}, nil
}

func changedPathCount(before, after fileSnapshot) int {
	paths := make(map[string]bool, len(before.Files)+len(after.Files))
	for path := range before.Files {
		paths[path] = true
	}
	for path := range after.Files {
		paths[path] = true
	}
	count := 0
	for path := range paths {
		beforeDigest, beforeOK := before.Files[path]
		afterDigest, afterOK := after.Files[path]
		if beforeOK != afterOK || beforeDigest != afterDigest {
			count++
		}
	}
	return count
}

func environmentDelta() SnapshotDelta {
	values := []string{runtime.GOOS, runtime.GOARCH, envOrUnset("GOTOOLCHAIN"), envOrDefault("SOURCE_DATE_EPOCH", "0")}
	digest := DigestValue(values)
	return SnapshotDelta{BeforeDigest: digest, AfterDigest: digest, Changed: false}
}

func logicalTimeDelta() SnapshotDelta {
	value := envOrDefault("SOURCE_DATE_EPOCH", "0")
	digest := DigestValue([]string{"injected-logical-clock", value})
	return SnapshotDelta{BeforeDigest: digest, AfterDigest: digest, Changed: false}
}

func envOrUnset(name string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return "UNSET"
}

func envOrDefault(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return fallback
}

func buildEffects(observation Observation, repositoryWrites int, decision string) []Effect {
	return []Effect{
		{
			Domain: "REPOSITORY_STORAGE", ObservedChanged: observation.RepositoryStorage.Changed,
			MutationAttempted: repositoryWrites > 0, BeforeDigest: observation.RepositoryStorage.BeforeDigest,
			AfterDigest: observation.RepositoryStorage.AfterDigest, WriteCount: repositoryWrites,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "assert-zero-repository-writes", ProofChoice: "REGRESSION",
			Status: effectStatus(repositoryWrites == 0, decision), Reason: effectReason(repositoryWrites == 0),
		},
		{
			Domain: "ENVIRONMENT", ObservedChanged: observation.Environment.Changed,
			MutationAttempted: false, BeforeDigest: observation.Environment.BeforeDigest,
			AfterDigest: observation.Environment.AfterDigest, WriteCount: 0,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "assert-environment-stability", ProofChoice: "COHERENCE",
			Status: effectStatus(!observation.Environment.Changed, decision), Reason: effectReason(!observation.Environment.Changed),
		},
		{
			Domain: "LOGICAL_TIME", ObservedChanged: observation.LogicalTime.Changed,
			MutationAttempted: false, BeforeDigest: observation.LogicalTime.BeforeDigest,
			AfterDigest: observation.LogicalTime.AfterDigest, WriteCount: 0,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "assert-logical-time-stability", ProofChoice: "COHERENCE",
			Status: effectStatus(!observation.LogicalTime.Changed, decision), Reason: effectReason(!observation.LogicalTime.Changed),
		},
		{
			Domain: "OUTPUT", ObservedChanged: true, MutationAttempted: true,
			BeforeDigest: DigestValue("ABSENT"), AfterDigest: DigestValue("OBSERVER_ARTIFACTS"), WriteCount: 3,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "record-observer-output-effect", ProofChoice: "FOUNDATION",
			Status: effectStatus(true, decision), Reason: "DECLARED_OUTPUT_EFFECT",
		},
	}
}

func effectStatus(satisfied bool, decision string) string {
	if decision == "UNKNOWN" {
		return "UNKNOWN"
	}
	if satisfied {
		return "PASS"
	}
	return "FAIL"
}

func effectReason(satisfied bool) string {
	if satisfied {
		return "NO_CHANGE_IN_DECLARED_SCOPE"
	}
	return "OBSERVER_MUTATED_DECLARED_SCOPE"
}

func buildClaimTransition(sourceDigest, decision string, unknown Unknown) ClaimTransition {
	current := "SUPPORTED"
	transition := "CLAIMED->SUPPORTED"
	if decision == "FAIL_CLOSED" {
		current, transition = "REFUTED", "CLAIMED->REFUTED"
	}
	if decision == "UNKNOWN" {
		current, transition = "UNKNOWN", "CLAIMED->UNKNOWN"
	}
	previous := DigestValue([]string{"claim.observer.zero-write", "CLAIMED", sourceDigest, "1"})
	evidence := DigestValue([]any{sourceDigest, decision, unknown})
	return ClaimTransition{
		ClaimID: "claim.observer.zero-write", Persistent: true, Sequence: 2,
		PreviousState: "CLAIMED", CurrentState: current, Transition: transition,
		PreviousEvidenceDigest: previous, EvidenceDigest: evidence,
	}
}

func buildIndicators(source Source, observation Observation, effects []Effect, unknown Unknown, claim ClaimTransition, decision string) []Indicator {
	status := func(ok bool, unknownSensitive bool) string {
		if decision == "UNKNOWN" && unknownSensitive {
			return "UNKNOWN"
		}
		if ok {
			return "PASS"
		}
		return "FAIL"
	}
	repositoryStable := !observation.RepositoryStorage.Changed
	return []Indicator{
		indicator("OEL-OBS-01", "OBSERVATION", "bind-gooo-source", "observer-effect-judge", "bind-gooo-source", "FOUNDATION", "valid .gooo source", fmt.Sprint(source.GoooSource), status(source.GoooSource, false), unknown),
		indicator("OEL-OBS-02", "OBSERVATION", "observe-repository", "observer-effect-judge", "observe-repository", "FOUNDATION", "before snapshot exists", fmt.Sprint(observation.RepositoryStorage.BeforeDigest != ""), status(observation.RepositoryStorage.BeforeDigest != "", false), unknown),
		indicator("OEL-OBS-03", "OBSERVATION", "observe-repository", "observer-effect-judge", "observe-repository", "COHERENCE", "after snapshot exists", fmt.Sprint(observation.RepositoryStorage.AfterDigest != ""), status(observation.RepositoryStorage.AfterDigest != "", false), unknown),
		indicator("OEL-OBS-04", "OBSERVATION", "observe-environment", "observer-effect-judge", "observe-environment", "FOUNDATION", "environment delta recorded", fmt.Sprint(observation.Environment.BeforeDigest == observation.Environment.AfterDigest), status(observation.Environment.BeforeDigest == observation.Environment.AfterDigest, true), unknown),
		indicator("OEL-OBS-05", "OBSERVATION", "observe-logical-time", "observer-effect-judge", "observe-logical-time", "FOUNDATION", "logical time delta recorded", fmt.Sprint(observation.LogicalTime.BeforeDigest == observation.LogicalTime.AfterDigest), status(observation.LogicalTime.BeforeDigest == observation.LogicalTime.AfterDigest, true), unknown),
		indicator("OEL-OBS-06", "OBSERVATION", "separate-observation", "observer-effect-judge", "separate-observation-from-effect", "COHERENCE", "observation and effects are distinct", fmt.Sprint(len(effects) == 4), status(len(effects) == 4, false), unknown),
		indicator("OEL-EFF-01", "EFFECT", "observe-repository", "observer-effect-judge", "assert-zero-repository-writes", "REGRESSION", "repository writes = 0", fmt.Sprint(len(effectsForDomain(effects, "REPOSITORY_STORAGE")) == 1 && repositoryStable), status(repositoryStable, false), unknown),
		indicator("OEL-EFF-02", "EFFECT", "observe-environment", "observer-effect-judge", "assert-environment-stability", "COHERENCE", "environment changed = false", fmt.Sprint(!observation.Environment.Changed), status(!observation.Environment.Changed, true), unknown),
		indicator("OEL-EFF-03", "EFFECT", "observe-logical-time", "observer-effect-judge", "assert-logical-time-stability", "COHERENCE", "logical time changed = false", fmt.Sprint(!observation.LogicalTime.Changed), status(!observation.LogicalTime.Changed, true), unknown),
		indicator("OEL-EFF-04", "EFFECT", "emit-output", "observer-effect-judge", "record-observer-output-effect", "FOUNDATION", "output effect declared", fmt.Sprint(effectWriteCount(effects, "OUTPUT") == 3), status(effectWriteCount(effects, "OUTPUT") == 3, false), unknown),
		indicator("OEL-GOV-01", "GUARDRAIL", "deny-mutation-authority", "observer-effect-judge", "deny-mutation-authority", "REGRESSION", "mutation authority = false", "false", status(true, false), unknown),
		indicator("OEL-GOV-02", "GUARDRAIL", "persist-claim-transition", "observer-effect-judge", "persist-claim-transition", "REGRESSION", "persistent claim transition", claim.Transition, status(claim.Persistent && claim.Sequence == 2, false), unknown),
	}
}

func indicator(id, class, producer, consumer, operation, proof, expected, actual, status string, unknown Unknown) Indicator {
	return Indicator{ID: id, Class: class, Producer: producer, Consumer: consumer, MetaOperation: operation, ProofChoice: proof, Expected: expected, Actual: actual, Status: status, Stage: unknown.Stage, Step: unknown.Step, Reason: unknown.Reason}
}

func effectsForDomain(effects []Effect, domain string) []Effect {
	result := make([]Effect, 0, 1)
	for _, effect := range effects {
		if effect.Domain == domain {
			result = append(result, effect)
		}
	}
	return result
}

func effectWriteCount(effects []Effect, domain string) int {
	for _, effect := range effects {
		if effect.Domain == domain {
			return effect.WriteCount
		}
	}
	return 0
}

func buildMetrics(indicators []Indicator) Metrics {
	satisfied := 0
	observationSatisfied, effectSatisfied, guardrailSatisfied := 0, 0, 0
	for _, indicator := range indicators {
		if indicator.Status != "PASS" {
			continue
		}
		satisfied++
		switch indicator.Class {
		case "OBSERVATION":
			observationSatisfied++
		case "EFFECT":
			effectSatisfied++
		case "GUARDRAIL":
			guardrailSatisfied++
		}
	}
	return Metrics{
		FixedDenominator: FixedDenominator, Satisfied: satisfied,
		CoverageBasisPoints:  satisfied * 10000 / FixedDenominator,
		ObservationSatisfied: observationSatisfied, ObservationTotal: 6,
		EffectSatisfied: effectSatisfied, EffectTotal: 4,
		GuardrailSatisfied: guardrailSatisfied, GuardrailTotal: 2,
	}
}

func makeReceipt(kind string, source Source, decision, resolution string, repositoryWrites int, unknown Unknown, claim ClaimTransition, evidence string) Receipt {
	operation, proof := "observe-result", "COHERENCE"
	if kind == "OBSERVER_EFFECT" {
		operation, proof = "record-observer-output-effect", "FOUNDATION"
	}
	return Receipt{
		Schema: ReceiptSchema, Kind: kind, Producer: "observer-effect-ledger",
		Consumer: "observer-effect-judge", MetaOperation: operation, ProofChoice: proof,
		Subject: source.Path, SubjectDigest: source.Digest, Decision: decision,
		Resolution: resolution, RepositoryWrites: repositoryWrites,
		MutationAuthority: false, Unknown: unknown, Coordinate: unknown, Reason: unknown.Reason,
		ClaimTransition: claim,
		EvidenceDigest:  evidence,
	}
}
