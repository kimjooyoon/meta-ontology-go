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
		Path: sourceDisplayPath(opts.Root, opts.SourcePath),
	}
	source = canonicalSource(source.Path, opts.SourcePath, sourcePayload)
	environmentBefore := readEnvironmentSample()
	logicalTimeBefore := readLogicalTimeSample()
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
	environmentAfter := readEnvironmentSample()
	logicalTimeAfter := readLogicalTimeSample()
	environment := environmentDelta(environmentBefore, environmentAfter)
	logicalTime := logicalTimeDelta(logicalTimeBefore, logicalTimeAfter)
	repositoryChanged := before.Digest != after.Digest
	repositoryWrites := changedPathCount(before, after)
	unknown := primaryUnknown(environment, logicalTime)
	if opts.Mode == "unknown" {
		unknown = Unknown{
			Stage:  "OBSERVE",
			Step:   "capture-environment",
			Reason: "DECLARED_ENVIRONMENT_NOT_AVAILABLE",
		}
		environment.Status = "UNKNOWN"
		environment.Resolution = "LOWER_RESOLUTION"
		environment.BeforeObserved = false
		environment.AfterObserved = false
		environment.Stage = unknown.Stage
		environment.Step = unknown.Step
		environment.Reason = unknown.Reason
	}
	topologyRoot := opts.TopologyRoot
	if topologyRoot == "" {
		topologyRoot = opts.Root
	}
	topology := buildTopology(topologyRoot)
	decision, resolution := "OBSERVED", "EXACT"
	if !isCanonicalSource(source) || !topology.Exact || repositoryChanged || environment.Changed || logicalTime.Changed {
		decision = "FAIL_CLOSED"
		resolution = "EXACT"
	} else if unknown.Reason != "NONE" || hasOpenOrUnknownCoordinate(environment, logicalTime) {
		decision = "UNKNOWN"
		resolution = "LOWER_RESOLUTION"
	}
	observation := Observation{
		RepositoryStorage: SnapshotDelta{
			BeforeDigest: before.Digest, AfterDigest: after.Digest, Changed: repositoryChanged,
			BeforeObserved: true, AfterObserved: true, Status: snapshotStatus(repositoryChanged, true),
			Resolution: resolutionForStatus(snapshotStatus(repositoryChanged, true)), Stage: "OBSERVE",
			Step: "scan-repository-boundaries", Reason: snapshotReason(repositoryChanged, true),
		},
		Environment: environment,
		LogicalTime: logicalTime,
	}
	effects := buildEffects(observation, repositoryWrites)
	coordinates := buildCoordinateAdjudications(observation)
	claim := buildClaimTransition(source.Digest, decision, unknown)
	indicators := buildIndicators(source, observation, effects, claim)
	metrics := buildMetrics(indicators)
	authority := Authority{
		RepositoryWrites: repositoryWrites, OutputWrites: 0,
		MutationAuthority: false, PromotionAuthorized: false,
	}
	report := Report{
		Schema: LedgerSchema, Experiment: ExperimentName, Source: source,
		Observation: observation, Effects: effects, Unknown: unknown,
		Coordinate: unknown, Reason: unknown.Reason,
		ClaimTransition: claim, Coordinates: coordinates, Topology: topology,
		RunnerScoped: runnerScopedEvidence(), Guardian: guardianExpectation(), Metrics: metrics, Authority: authority,
		RepositoryWrites: repositoryWrites, MutationAuthority: false,
		PromotionAuthorized: false, Decision: decision, Resolution: resolution,
		Indicators: indicators,
	}
	report.EvidenceDigest = DigestValue([]any{report.Source, report.Observation, report.Effects, report.Unknown, report.ClaimTransition, report.Coordinates, report.Topology, report.RunnerScoped, report.Guardian})
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

type environmentSample struct {
	Values []string
}

type logicalTimeSample struct {
	Value    string
	Observed bool
}

func readEnvironmentSample() environmentSample {
	return environmentSample{Values: []string{
		runtime.GOOS, runtime.GOARCH, envOrUnset("GOTOOLCHAIN"), envOrUnset("SOURCE_DATE_EPOCH"),
	}}
}

func readLogicalTimeSample() logicalTimeSample {
	value, observed := os.LookupEnv("SOURCE_DATE_EPOCH")
	if !observed {
		return logicalTimeSample{Value: "UNAVAILABLE", Observed: false}
	}
	return logicalTimeSample{Value: value, Observed: true}
}

func environmentDelta(before, after environmentSample) SnapshotDelta {
	beforeDigest := DigestValue(before.Values)
	afterDigest := DigestValue(after.Values)
	changed := beforeDigest != afterDigest
	return SnapshotDelta{
		BeforeDigest: beforeDigest, AfterDigest: afterDigest, Changed: changed,
		BeforeObserved: len(before.Values) == 4, AfterObserved: len(after.Values) == 4,
		Status: snapshotStatus(changed, true), Resolution: "EXACT",
		Stage: "OBSERVE", Step: "capture-environment-boundaries", Reason: snapshotReason(changed, true),
	}
}

func logicalTimeDelta(before, after logicalTimeSample) SnapshotDelta {
	beforeDigest := DigestValue([]string{"injected-logical-clock", before.Value})
	afterDigest := DigestValue([]string{"injected-logical-clock", after.Value})
	changed := beforeDigest != afterDigest
	status := snapshotStatus(changed, before.Observed && after.Observed)
	reason := snapshotReason(changed, before.Observed && after.Observed)
	if !before.Observed || !after.Observed {
		status = "UNKNOWN"
		reason = "SOURCE_DATE_EPOCH_NOT_DECLARED"
	}
	return SnapshotDelta{
		BeforeDigest: beforeDigest, AfterDigest: afterDigest, Changed: changed,
		BeforeObserved: before.Observed, AfterObserved: after.Observed,
		Status: status, Resolution: resolutionForStatus(status),
		Stage: "OBSERVE", Step: "capture-logical-time-boundaries", Reason: reason,
	}
}

func snapshotStatus(changed, observed bool) string {
	if !observed {
		return "UNKNOWN"
	}
	if changed {
		return "FAIL"
	}
	return "PASS"
}

func snapshotReason(changed, observed bool) string {
	if !observed {
		return "DECLARED_COORDINATE_NOT_OBSERVED"
	}
	if changed {
		return "DECLARED_COORDINATE_CHANGED"
	}
	return "INDEPENDENT_BOUNDARY_READS_MATCHED"
}

func resolutionForStatus(status string) string {
	if status == "PASS" || status == "FAIL" {
		return "EXACT"
	}
	return "LOWER_RESOLUTION"
}

func primaryUnknown(environment, logicalTime SnapshotDelta) Unknown {
	if logicalTime.Status == "UNKNOWN" {
		return Unknown{Stage: logicalTime.Stage, Step: logicalTime.Step, Reason: logicalTime.Reason}
	}
	return Unknown{Stage: "EMIT_OUTPUT", Step: "artifact-write", Reason: "ACTUAL_OUTPUT_WRITES_NOT_INSTRUMENTED"}
}

func hasOpenOrUnknownCoordinate(environment, logicalTime SnapshotDelta) bool {
	return environment.Status == "UNKNOWN" || logicalTime.Status == "UNKNOWN"
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

func buildEffects(observation Observation, repositoryWrites int) []Effect {
	return []Effect{
		{
			Domain: "REPOSITORY_STORAGE", ObservedChanged: observation.RepositoryStorage.Changed,
			MutationAttempted: repositoryWrites > 0, BeforeDigest: observation.RepositoryStorage.BeforeDigest,
			AfterDigest: observation.RepositoryStorage.AfterDigest, WriteCount: repositoryWrites,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "assert-zero-repository-writes", ProofChoice: "REGRESSION",
			Status: statusFromSnapshot(observation.RepositoryStorage.Status),
			Stage:  observation.RepositoryStorage.Stage, Step: observation.RepositoryStorage.Step,
			Reason: observation.RepositoryStorage.Reason,
		},
		{
			Domain: "ENVIRONMENT", ObservedChanged: observation.Environment.Changed,
			MutationAttempted: false, BeforeDigest: observation.Environment.BeforeDigest,
			AfterDigest: observation.Environment.AfterDigest, WriteCount: 0,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "assert-environment-stability", ProofChoice: "COHERENCE",
			Status: statusFromSnapshot(observation.Environment.Status),
			Stage:  observation.Environment.Stage, Step: observation.Environment.Step,
			Reason: observation.Environment.Reason,
		},
		{
			Domain: "LOGICAL_TIME", ObservedChanged: observation.LogicalTime.Changed,
			MutationAttempted: false, BeforeDigest: observation.LogicalTime.BeforeDigest,
			AfterDigest: observation.LogicalTime.AfterDigest, WriteCount: 0,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "assert-logical-time-stability", ProofChoice: "COHERENCE",
			Status: statusFromSnapshot(observation.LogicalTime.Status),
			Stage:  observation.LogicalTime.Stage, Step: observation.LogicalTime.Step,
			Reason: observation.LogicalTime.Reason,
		},
		{
			Domain: "OUTPUT", ObservedChanged: false, MutationAttempted: false, Planned: true,
			BeforeDigest: "UNOBSERVED", AfterDigest: "UNOBSERVED", WriteCount: 0,
			Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "plan-observer-output-effect", ProofChoice: "FOUNDATION",
			Status: "OPEN", Stage: "EMIT_OUTPUT", Step: "artifact-write",
			Reason: "ACTUAL_OUTPUT_WRITES_NOT_INSTRUMENTED",
		},
	}
}

func statusFromSnapshot(status string) string {
	if status == "PASS" || status == "FAIL" || status == "UNKNOWN" {
		return status
	}
	return "UNKNOWN"
}

func buildCoordinateAdjudications(observation Observation) []CoordinateAdjudication {
	return []CoordinateAdjudication{
		coordinateFromSnapshot("REPOSITORY_STORAGE", observation.RepositoryStorage, "assert-zero-repository-writes", "REGRESSION"),
		coordinateFromSnapshot("ENVIRONMENT", observation.Environment, "assert-environment-stability", "COHERENCE"),
		coordinateFromSnapshot("LOGICAL_TIME", observation.LogicalTime, "assert-logical-time-stability", "COHERENCE"),
		{
			Coordinate: "OUTPUT", Status: "OPEN", Resolution: "LOWER_RESOLUTION",
			BeforeObserved: false, AfterObserved: false, Stage: "EMIT_OUTPUT", Step: "artifact-write",
			Reason: "ACTUAL_OUTPUT_WRITES_NOT_INSTRUMENTED", Producer: "observer-effect-ledger",
			Consumer: "observer-effect-judge", MetaOperation: "plan-observer-output-effect", ProofChoice: "FOUNDATION",
		},
	}
}

func coordinateFromSnapshot(coordinate string, snapshot SnapshotDelta, operation, proof string) CoordinateAdjudication {
	return CoordinateAdjudication{
		Coordinate: coordinate, Status: snapshot.Status, Resolution: snapshot.Resolution,
		BeforeObserved: snapshot.BeforeObserved, AfterObserved: snapshot.AfterObserved,
		Stage: snapshot.Stage, Step: snapshot.Step, Reason: snapshot.Reason,
		Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
		MetaOperation: operation, ProofChoice: proof,
	}
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

func buildIndicators(source Source, observation Observation, effects []Effect, claim ClaimTransition) []Indicator {
	repository := effectByDomain(effects, "REPOSITORY_STORAGE")
	output := effectByDomain(effects, "OUTPUT")
	sourceEvidence := Unknown{Stage: "BIND", Step: "parse-and-lower-source", Reason: "CANONICAL_GOOO_PARSE_LOWERING_BOUND"}
	separationEvidence := Unknown{Stage: "ADJUDICATE", Step: "separate-observation-effect", Reason: "OBSERVATION_AND_EFFECT_SECTIONS_SEPARATE"}
	mutationEvidence := Unknown{Stage: "GUARD", Step: "deny-mutation-authority", Reason: "MUTATION_AND_PROMOTION_AUTHORITY_FALSE"}
	claimEvidence := Unknown{Stage: "CLAIM", Step: "persist-claim-transition", Reason: "CLAIM_TRANSITION_PERSISTED"}
	return []Indicator{
		indicator("OEL-OBS-01", "OBSERVATION", "bind-gooo-source", "observer-effect-judge", "bind-gooo-source", "FOUNDATION", "canonical .gooo parse and lowering", fmt.Sprint(isCanonicalSource(source)), indicatorStatus(boolStatus(isCanonicalSource(source))), sourceEvidence),
		indicator("OEL-OBS-02", "OBSERVATION", "observe-repository", "observer-effect-judge", "observe-repository", "FOUNDATION", "before snapshot exists", fmt.Sprint(observation.RepositoryStorage.BeforeObserved), indicatorStatus(boolStatus(observation.RepositoryStorage.BeforeObserved)), snapshotUnknown(observation.RepositoryStorage)),
		indicator("OEL-OBS-03", "OBSERVATION", "observe-repository", "observer-effect-judge", "observe-repository", "COHERENCE", "after snapshot exists", fmt.Sprint(observation.RepositoryStorage.AfterObserved), indicatorStatus(boolStatus(observation.RepositoryStorage.AfterObserved)), snapshotUnknown(observation.RepositoryStorage)),
		indicator("OEL-OBS-04", "OBSERVATION", "observe-environment", "observer-effect-judge", "observe-environment", "FOUNDATION", "environment delta recorded", observation.Environment.Status, indicatorStatus(observation.Environment.Status), snapshotUnknown(observation.Environment)),
		indicator("OEL-OBS-05", "OBSERVATION", "observe-logical-time", "observer-effect-judge", "observe-logical-time", "FOUNDATION", "logical time delta recorded", observation.LogicalTime.Status, indicatorStatus(observation.LogicalTime.Status), snapshotUnknown(observation.LogicalTime)),
		indicator("OEL-OBS-06", "OBSERVATION", "separate-observation", "observer-effect-judge", "separate-observation-from-effect", "COHERENCE", "observation and effects are distinct", fmt.Sprint(len(effects) == 4), indicatorStatus(boolStatus(len(effects) == 4)), separationEvidence),
		indicator("OEL-EFF-01", "EFFECT", "observe-repository", "observer-effect-judge", "assert-zero-repository-writes", "REGRESSION", "repository writes = 0", repository.Status, indicatorStatus(repository.Status), snapshotUnknown(observation.RepositoryStorage)),
		indicator("OEL-EFF-02", "EFFECT", "observe-environment", "observer-effect-judge", "assert-environment-stability", "COHERENCE", "environment changed = false", observation.Environment.Status, indicatorStatus(observation.Environment.Status), snapshotUnknown(observation.Environment)),
		indicator("OEL-EFF-03", "EFFECT", "observe-logical-time", "observer-effect-judge", "assert-logical-time-stability", "COHERENCE", "logical time changed = false", observation.LogicalTime.Status, indicatorStatus(observation.LogicalTime.Status), snapshotUnknown(observation.LogicalTime)),
		indicator("OEL-EFF-04", "EFFECT", "emit-output", "observer-effect-judge", "plan-observer-output-effect", "FOUNDATION", "actual output effect observed", output.Status, indicatorStatus(output.Status), effectUnknown(output)),
		indicator("OEL-GOV-01", "GUARDRAIL", "deny-mutation-authority", "observer-effect-judge", "deny-mutation-authority", "REGRESSION", "mutation authority = false", "false", indicatorStatus("PASS"), mutationEvidence),
		indicator("OEL-GOV-02", "GUARDRAIL", "persist-claim-transition", "observer-effect-judge", "persist-claim-transition", "REGRESSION", "persistent claim transition", claim.Transition, indicatorStatus(boolStatus(claim.Persistent && claim.Sequence == 2)), claimEvidence),
	}
}

func indicatorStatus(status string) string {
	if status == "PASS" {
		return "PASS"
	}
	if status == "FAIL" {
		return "FAIL"
	}
	return "UNKNOWN"
}

func boolStatus(value bool) string {
	if value {
		return "PASS"
	}
	return "FAIL"
}

func snapshotUnknown(snapshot SnapshotDelta) Unknown {
	return Unknown{Stage: snapshot.Stage, Step: snapshot.Step, Reason: snapshot.Reason}
}

func effectUnknown(effect Effect) Unknown {
	return Unknown{Stage: effect.Stage, Step: effect.Step, Reason: effect.Reason}
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

func effectByDomain(effects []Effect, domain string) Effect {
	for _, effect := range effects {
		if effect.Domain == domain {
			return effect
		}
	}
	return Effect{}
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
