package generation

import (
	"bufio"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

// SemanticObservationRecorder is scoped to one bounded compiler run. It is a
// recorder for one semantic operation, rather than a general profiler.
type SemanticObservationRecorder struct {
	events []SemanticObservationEvent
}

// NewSemanticObservationRecorder creates an empty bounded-run recorder.
func NewSemanticObservationRecorder() *SemanticObservationRecorder {
	return &SemanticObservationRecorder{}
}

// Record stores one exact compiler operation observation. Sequence numbers are
// assigned here so callers cannot create a second ordering convention.
func (r *SemanticObservationRecorder) Record(event SemanticObservationEvent) error {
	if r == nil {
		return errors.New("semantic observation recorder is nil")
	}
	if event.Sequence == 0 {
		event.Sequence = len(r.events) + 1
	}
	if event.Sequence != len(r.events)+1 {
		return fmt.Errorf("semantic observation sequence %d is not next sequence %d", event.Sequence, len(r.events)+1)
	}
	event.Effects = append([]string(nil), event.Effects...)
	event.SourceSpans = append([]SemanticObservationSpan(nil), event.SourceSpans...)
	r.events = append(r.events, event)
	return nil
}

// Events returns a copy of the bounded-run event stream.
func (r *SemanticObservationRecorder) Events() []SemanticObservationEvent {
	if r == nil {
		return nil
	}
	events := make([]SemanticObservationEvent, len(r.events))
	copy(events, r.events)
	for index := range events {
		events[index].Effects = append([]string(nil), events[index].Effects...)
		events[index].SourceSpans = append([]SemanticObservationSpan(nil), events[index].SourceSpans...)
	}
	return events
}

// BuildSemanticObservation derives candidates only from repeated pure
// operation/input pairs. It never infers safety, benefit, or a rewrite.
func (r *SemanticObservationRecorder) BuildSemanticObservation(contract SemanticObservationContract, contractDigest, inputSourceDigest string, pair SemanticObservationPairEvidence) (SemanticObservation, error) {
	if r == nil {
		return SemanticObservation{}, errors.New("semantic observation recorder is nil")
	}
	if err := validateSemanticObservationContract(contract); err != nil {
		return SemanticObservation{}, err
	}
	if !knownEnvelopeDigest(contractDigest) {
		return SemanticObservation{}, errors.New("semantic observation contract digest is unknown")
	}
	if inputSourceDigest != "" && !knownEnvelopeDigest(inputSourceDigest) {
		return SemanticObservation{}, errors.New("semantic observation input source digest is unknown")
	}

	type group struct {
		phase       string
		operationID string
		inputDigest string
		count       int
		spans       []SemanticObservationSpan
	}
	groups := make(map[string]*group)
	digests := make(map[string]struct{})
	for index, event := range r.events {
		if event.Sequence != index+1 {
			return SemanticObservation{}, fmt.Errorf("semantic observation event %d has sequence %d", index+1, event.Sequence)
		}
		if event.Phase != contract.Phase || event.OperationID != contract.OperationID || !event.Pure {
			return SemanticObservation{}, fmt.Errorf("event %d is outside the declared pure semantic operation", event.Sequence)
		}
		if !knownEnvelopeDigest(event.InputDigest) {
			return SemanticObservation{}, fmt.Errorf("event %d has unknown canonical input digest", event.Sequence)
		}
		if !subsetStrings(event.Effects, contract.AllowedEffects) {
			return SemanticObservation{}, fmt.Errorf("event %d uses an effect outside the declared observation grant", event.Sequence)
		}
		if len(event.SourceSpans) == 0 {
			return SemanticObservation{}, fmt.Errorf("event %d has no source span", event.Sequence)
		}
		for _, span := range event.SourceSpans {
			if !validSemanticObservationSpan(span) {
				return SemanticObservation{}, fmt.Errorf("event %d has an invalid source span", event.Sequence)
			}
		}
		key := event.Phase + "\x00" + event.OperationID + "\x00" + event.InputDigest
		current := groups[key]
		if current == nil {
			current = &group{phase: event.Phase, operationID: event.OperationID, inputDigest: event.InputDigest}
			groups[key] = current
		}
		current.count++
		current.spans = appendUniqueObservationSpans(current.spans, event.SourceSpans...)
		digests[event.InputDigest] = struct{}{}
	}

	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	candidates := make([]SemanticObservationCandidate, 0, len(keys))
	duplicateEvaluations := 0
	for _, key := range keys {
		current := groups[key]
		if current.count < 2 {
			continue
		}
		duplicateEvaluations += current.count - 1
		spans := sortedObservationSpans(current.spans)
		candidate := SemanticObservationCandidate{
			StableID:               semanticObservationCandidateID(current.phase, current.operationID, current.inputDigest, spans),
			Phase:                  current.phase,
			OperationID:            current.operationID,
			InputDigest:            current.inputDigest,
			SourceSpans:            spans,
			ObservedCount:          current.count,
			ExpectedReducibleCount: current.count - 1,
			SafetyAssessment:       "UNKNOWN_NOT_INFERRED",
			BenefitAssessment:      "UNKNOWN_NOT_INFERRED",
		}
		candidates = append(candidates, candidate)
	}

	metrics := SemanticObservationMetrics{
		ObservedOperations:   len(r.events),
		DistinctInputDigests: len(digests),
		DuplicateEvaluations: duplicateEvaluations,
		CandidatesEmitted:    len(candidates),
		BeforeOperationCount: pair.BeforeOperationCount,
		AfterOperationCount:  pair.AfterOperationCount,
		RepositoryWrites:     0,
		LocalTestExecutions:  0,
	}
	return SemanticObservation{
		Schema:               SemanticObservationSchema,
		Contract:             contract,
		ContractDigest:       contractDigest,
		InputSourceDigest:    inputSourceDigest,
		Events:               r.Events(),
		ObservedOperations:   len(r.events),
		DistinctInputDigests: len(digests),
		DuplicateEvaluations: duplicateEvaluations,
		CandidatesEmitted:    len(candidates),
		Candidates:           candidates,
		Decision:             "CLOSED",
		Reason:               observationReason(len(candidates)),
		PairEvidence:         pair,
		Metrics:              metrics,
	}, nil
}

// ParseSemanticObservationContract extracts the observation declaration from
// the eight-activity .gooo authority. This is intentionally stricter than the
// generic envelope parser: every observation boundary is explicit in .gooo.
func ParseSemanticObservationContract(source []byte) (SemanticObservationContract, error) {
	if len(source) == 0 {
		return SemanticObservationContract{}, errors.New(".gooo observation authority is empty")
	}
	if err := validateSemanticOperationAuthority(source); err != nil {
		return SemanticObservationContract{}, err
	}
	fields := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(source)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "activity ") {
			continue
		}
		_, after, ok := strings.Cut(line, " computes ")
		if !ok {
			continue
		}
		payload, err := strconv.Unquote(strings.TrimSpace(after))
		if err != nil {
			return SemanticObservationContract{}, fmt.Errorf("decode activity computation: %w", err)
		}
		for item := range strings.SplitSeq(payload, ";") {
			key, value, ok := strings.Cut(item, "=")
			if !ok || strings.TrimSpace(key) == "" {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if !observationDeclarationKey(key) {
				continue
			}
			if previous, exists := fields[key]; exists && previous != value {
				return SemanticObservationContract{}, fmt.Errorf("observation declaration key %q conflicts", key)
			}
			fields[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return SemanticObservationContract{}, fmt.Errorf("scan observation authority: %w", err)
	}
	effects := strings.Split(fields["effects"], "|")
	pure, err := strconv.ParseBool(fields["pure"])
	if err != nil {
		return SemanticObservationContract{}, errors.New("observation declaration pure must be true or false")
	}
	contract := SemanticObservationContract{
		Activity:               fields["observation"],
		Phase:                  fields["phase"],
		OperationID:            fields["operation_id"],
		CanonicalInputIdentity: fields["input_identity"],
		AllowedEffects:         effects,
		Pure:                   pure,
		CandidateRule:          fields["candidate_rule"],
	}
	if err := validateSemanticObservationContract(contract); err != nil {
		return SemanticObservationContract{}, err
	}
	if fields["repository_writes"] != "0" || fields["local_test_executions"] != "0" {
		return SemanticObservationContract{}, errors.New("observation authority must declare zero repository writes and local test executions")
	}
	return contract, nil
}

func observationDeclarationKey(key string) bool {
	switch key {
	case "observation", "phase", "operation_id", "input_identity", "effects", "pure", "candidate_rule", "repository_writes", "local_test_executions":
		return true
	default:
		return false
	}
}

func validateSemanticObservationContract(contract SemanticObservationContract) error {
	if contract.Activity != SemanticObservationActivity ||
		contract.Phase != SemanticObservationPhase ||
		contract.OperationID != SemanticObservationOperationID ||
		contract.CanonicalInputIdentity != SemanticObservationInputIdentity ||
		!contract.Pure ||
		contract.CandidateRule != SemanticObservationCandidateRule {
		return errors.New(".gooo observation declaration does not match the released compiler observation contract")
	}
	if len(contract.AllowedEffects) != 2 || contract.AllowedEffects[0] != "read:source" || contract.AllowedEffects[1] != "read:semantic-ir" {
		return errors.New("observation authority must declare exactly read:source and read:semantic-ir effects")
	}
	return nil
}

func knownEnvelopeDigest(value string) bool {
	return cache.Digest(value).Known()
}

func subsetStrings(values, allowed []string) bool {
	for _, value := range values {
		found := slices.Contains(allowed, value)
		if !found {
			return false
		}
	}
	return true
}

func appendUniqueObservationSpans(spans []SemanticObservationSpan, additions ...SemanticObservationSpan) []SemanticObservationSpan {
	for _, addition := range additions {
		found := slices.Contains(spans, addition)
		if !found {
			spans = append(spans, addition)
		}
	}
	return spans
}

func sortedObservationSpans(spans []SemanticObservationSpan) []SemanticObservationSpan {
	result := append([]SemanticObservationSpan(nil), spans...)
	sort.Slice(result, func(left, right int) bool {
		if result[left].File != result[right].File {
			return result[left].File < result[right].File
		}
		if result[left].StartOffset != result[right].StartOffset {
			return result[left].StartOffset < result[right].StartOffset
		}
		if result[left].EndOffset != result[right].EndOffset {
			return result[left].EndOffset < result[right].EndOffset
		}
		return result[left].StartLine < result[right].StartLine
	})
	return result
}

func validSemanticObservationSpan(span SemanticObservationSpan) bool {
	return span.File != "" && span.StartOffset >= 0 && span.EndOffset >= span.StartOffset && span.StartLine > 0 && span.StartColumn > 0 && span.EndLine >= span.StartLine && span.EndColumn > 0
}

func semanticObservationCandidateID(phase, operationID, inputDigest string, _ []SemanticObservationSpan) string {
	digest, _ := cache.DigestOf(struct {
		Phase       string
		OperationID string
		InputDigest string
	}{phase, operationID, inputDigest})
	return "self-observation/candidate/" + digest.String()
}

func observationReason(candidateCount int) string {
	if candidateCount == 0 {
		return "NO_DUPLICATE_SEMANTIC_OPERATION"
	}
	return "EXACT_DUPLICATE_SEMANTIC_OPERATION"
}
