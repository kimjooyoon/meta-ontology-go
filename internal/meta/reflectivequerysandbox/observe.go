package reflectivequerysandbox

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Observe(path string, subjectSHA, repositoryBeforePath, repositoryAfterPath string) (Observation, error) {
	return ObserveWithCheckoutEvidence(path, subjectSHA, "", repositoryBeforePath, repositoryAfterPath)
}

func ObserveWithCheckoutEvidence(path, subjectSHA, checkoutEvidence, repositoryBeforePath, repositoryAfterPath string) (Observation, error) {
	if path != SourcePath {
		return Observation{}, fmt.Errorf("source path is not canonical: got %q want %q", path, SourcePath)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Observation{}, fmt.Errorf("read source: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(data))
	if diagnostics.HasErrors() {
		return Observation{}, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Observation{}, fmt.Errorf("lower source: %w", err)
	}
	graph, err := query.FromSemanticIR(ir)
	if err != nil {
		return Observation{}, fmt.Errorf("project query view: %w", err)
	}
	model, err := deriveSourceModel(ir)
	if err != nil {
		return Observation{}, fmt.Errorf("derive source contract: %w", err)
	}

	semanticDigest := ir.StableHash()
	source := Snapshot{
		Path: path, SourceDigest: semantic.StableHash(data), SemanticDigest: semanticDigest,
		GraphDigest: graph.StableHash(), NodeCount: len(graph.Nodes()), FactCount: len(graph.AllFacts()), GoooLines: countLines(data),
	}
	effects := observeEffects(repositoryBeforePath, repositoryAfterPath)
	attempts, immutableAccepted, detachedCapability, mutationAPI, mutationOutcome, mutationError, err := buildAttempts(ir, graph, model, semanticDigest, source.GraphDigest, effects)
	if err != nil {
		return Observation{}, err
	}
	effects.ImmutableIDPatchAccepted = immutableAccepted
	effects.DetachedGraphPatchCapability = detachedCapability
	effects.OverallAuthority = "UNKNOWN"
	effects.MutationAPI = mutationAPI
	effects.MutationOutcome = mutationOutcome
	effects.MutationError = mutationError

	// The producer records material but does not seal it. The receipt.seal attempt
	// remains provisional until the independent consumer binds this digest.
	for index := range attempts {
		if attempts[index].ID == "receipt.seal" {
			attempts[index].ObservedMaterialDigest = ""
		}
	}
	claims := buildClaimTransitions(model.Claims, attempts, effects, "")
	chainDigest, err := validateTransitionChain(claims)
	if err != nil {
		return Observation{}, fmt.Errorf("build provisional claim chain: %w", err)
	}
	value := Observation{
		Schema: Schema, SubjectSHA: subjectSHA, Contract: buildContract(model, source, attempts, claims),
		Source: source, Attempts: attempts, Claims: claims, Effects: effects, SubjectBinding: bindSubjectSHA(subjectSHA, checkoutEvidence),
		Provisional: true, TransitionChainDigest: chainDigest, Producer: ProducerName,
	}
	value.ProvisionalDigest = observationDigest(value)
	return value, nil
}

func bindSubjectSHA(value, checkoutEvidence string) SubjectBinding {
	binding := SubjectBinding{Value: value}
	binding.Format = SubjectEvidence{Stage: "SUBJECT", Step: "validate-sha-format"}
	if len(value) == 40 {
		if _, err := hex.DecodeString(value); err == nil {
			binding.Format.Decision, binding.Format.Resolution, binding.Format.Reason = "PASS", "EXACT", "FORMAT_VALID"
			binding.Format.Digest = semantic.StableHashString(value + "|FORMAT_VALID")
		} else {
			binding.Format.Decision, binding.Format.Resolution, binding.Format.Reason = "UNKNOWN", "LOWER_RESOLUTION", "FORMAT_INVALID"
			binding.Format.Digest = semantic.StableHashString(value + "|FORMAT_INVALID")
		}
	} else {
		binding.Format.Decision, binding.Format.Resolution, binding.Format.Reason = "UNKNOWN", "LOWER_RESOLUTION", "FORMAT_INVALID"
		binding.Format.Digest = semantic.StableHashString(value + "|FORMAT_INVALID")
	}

	binding.Checkout = SubjectEvidence{Stage: "SUBJECT", Step: "verify-checkout-evidence"}
	if checkoutEvidence == "" {
		binding.Checkout.Decision, binding.Checkout.Resolution, binding.Checkout.Reason = "UNKNOWN", "LOWER_RESOLUTION", "SUBJECT_SHA_CHECKOUT_UNOBSERVED"
		binding.Checkout.Digest = semantic.StableHashString(value + "|SUBJECT_SHA_CHECKOUT_UNOBSERVED")
		return binding
	}
	binding.Checkout.EvidenceDigest = semantic.StableHash([]byte(checkoutEvidence))
	lines := strings.Split(strings.TrimSuffix(checkoutEvidence, "\n"), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "subject_sha=") || !strings.HasPrefix(lines[1], "checkout_head=") || !strings.HasPrefix(lines[2], "subject_matches_checkout=") {
		binding.Checkout.Decision, binding.Checkout.Resolution, binding.Checkout.Reason = "UNKNOWN", "LOWER_RESOLUTION", "SUBJECT_SHA_CHECKOUT_EVIDENCE_INVALID"
		binding.Checkout.Digest = semantic.StableHashString(binding.Checkout.EvidenceDigest + "|SUBJECT_SHA_CHECKOUT_EVIDENCE_INVALID")
		return binding
	}
	evidenceSubject := strings.TrimPrefix(lines[0], "subject_sha=")
	observedSHA := strings.TrimPrefix(lines[1], "checkout_head=")
	matchValue := strings.TrimPrefix(lines[2], "subject_matches_checkout=")
	binding.Checkout.ObservedSHA = observedSHA
	match := matchValue == "true"
	if evidenceSubject != value || len(observedSHA) != 40 {
		binding.Checkout.Decision, binding.Checkout.Resolution, binding.Checkout.Reason = "UNKNOWN", "LOWER_RESOLUTION", "SUBJECT_SHA_CHECKOUT_EVIDENCE_INVALID"
		binding.Checkout.Digest = semantic.StableHashString(binding.Checkout.EvidenceDigest + "|SUBJECT_SHA_CHECKOUT_EVIDENCE_INVALID")
		return binding
	}
	if _, err := hex.DecodeString(observedSHA); err != nil || (match != (value == observedSHA)) {
		binding.Checkout.Decision, binding.Checkout.Resolution, binding.Checkout.Reason = "UNKNOWN", "LOWER_RESOLUTION", "SUBJECT_SHA_CHECKOUT_EVIDENCE_INVALID"
		binding.Checkout.Digest = semantic.StableHashString(binding.Checkout.EvidenceDigest + "|SUBJECT_SHA_CHECKOUT_EVIDENCE_INVALID")
		return binding
	}
	if binding.Format.Decision == "PASS" && value == observedSHA && match {
		binding.Checkout.Decision, binding.Checkout.Resolution, binding.Checkout.Reason = "PASS", "EXACT", "CHECKOUT_BOUND"
		binding.Checkout.Digest = semantic.StableHashString(binding.Checkout.EvidenceDigest + "|CHECKOUT_BOUND")
		return binding
	}
	if binding.Format.Decision == "PASS" {
		binding.Checkout.Decision, binding.Checkout.Resolution, binding.Checkout.Reason = "REFUTED", "EXACT", "SUBJECT_SHA_CHECKOUT_MISMATCH"
		binding.Checkout.Digest = semantic.StableHashString(binding.Checkout.EvidenceDigest + "|SUBJECT_SHA_CHECKOUT_MISMATCH")
		return binding
	}
	binding.Checkout.Decision, binding.Checkout.Resolution, binding.Checkout.Reason = "UNKNOWN", "LOWER_RESOLUTION", "SUBJECT_SHA_CHECKOUT_UNRESOLVED"
	binding.Checkout.Digest = semantic.StableHashString(binding.Checkout.EvidenceDigest + "|SUBJECT_SHA_CHECKOUT_UNRESOLVED")
	return binding
}

func buildAttempts(ir semantic.IR, graph *query.Graph, model sourceModel, semanticDigest, graphDigest string, effects Effects) ([]Attempt, bool, string, string, string, string, error) {
	attempts := make([]Attempt, 0, len(model.Operations)+1)
	var immutableAccepted bool
	detachedCapability, mutationAPI, mutationOutcome, mutationError := "UNKNOWN", "", "", ""
	for _, operation := range model.Operations {
		target, err := targetForOperation(ir, operation, model)
		if err != nil {
			return nil, false, "UNKNOWN", "", "", "", err
		}
		attemptID := attemptIDForProgram(operation.Program)
		switch {
		case strings.HasPrefix(operation.Program, "reflect.query:"):
			attempts = append(attempts, exactAttempt(graph, operation, attemptID, target, semanticDigest, graphDigest, "QUERY", model.Claims))
		case operation.Program == "reflect.observation:net-repository-status-unchanged":
			attempts = append(attempts, repositoryAttempt(operation, attemptID, target, semanticDigest, effects, model.Claims))
		case strings.HasPrefix(operation.Program, "reflect.observation:"):
			attempts = append(attempts, exactAttempt(graph, operation, attemptID, target, semanticDigest, graphDigest, "RECEIPT", model.Claims))
		case strings.HasPrefix(operation.Program, "reflect.attempt:"):
			attempt, accepted, capability, api, outcome, apiError, err := mutationAttempt(ir, operation, attemptID, target, semanticDigest, graphDigest, model, model.Claims)
			if err != nil {
				return nil, false, "UNKNOWN", "", "", "", err
			}
			attempts = append(attempts, attempt)
			immutableAccepted, detachedCapability, mutationAPI, mutationOutcome, mutationError = accepted, capability, api, outcome, apiError
		}
	}
	metricOperation, ok := findOperation(model.Operations, "reflect.query:metrics")
	if !ok {
		return nil, false, "UNKNOWN", "", "", "", fmt.Errorf("source declares no metrics query operation")
	}
	attempts = append(attempts, unknownAttempt(graph, metricOperation, model.UnknownTarget, semanticDigest, graphDigest, model.Claims))
	for index := range attempts {
		for _, claim := range model.Claims {
			if claim.EvidenceAttempt == attempts[index].ID {
				attempts[index].EvidenceClaimIDs = append(attempts[index].EvidenceClaimIDs, claim.ID)
			}
		}
	}
	sort.SliceStable(attempts, func(i, j int) bool { return attempts[i].ID < attempts[j].ID })
	return attempts, immutableAccepted, detachedCapability, mutationAPI, mutationOutcome, mutationError, nil
}

func exactAttempt(graph *query.Graph, operation operationSpec, id string, target semantic.ID, semanticDigest, graphDigest, stage string, claims []claimSpec) Attempt {
	before := graph.StableHash()
	attempt := Attempt{
		ID: id, Class: classForAttempt(id, claims), Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(),
		MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName,
		ProofChoice: proofForAttempt(id, claims), Stage: stage, Step: "match-source-relation",
		SemanticDigestBefore: semanticDigest, GraphDigestBefore: before,
	}
	result, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, query.ID(target.String())))
	after := graph.StableHash()
	attempt.SemanticDigestAfter, attempt.GraphDigestAfter = semanticDigest, after
	attempt.ObservedMaterialDigest = after
	if err != nil {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "QUERY_API_ERROR"
		return attempt
	}
	attempt.MatchedFacts = len(result.All())
	if attempt.MatchedFacts == 1 {
		attempt.Decision, attempt.Resolution, attempt.Reason = "PASS", "EXACT", "EXACT_RELATION_MATCH"
	} else {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "RELATION_NOT_OBSERVED"
	}
	return attempt
}

func mutationAttempt(ir semantic.IR, operation operationSpec, id string, target semantic.ID, semanticDigest, graphDigest string, model sourceModel, claims []claimSpec) (Attempt, bool, string, string, string, string, error) {
	node, ok := ir.Graph.Node(target)
	if !ok {
		return Attempt{}, false, "UNKNOWN", "", "", "", fmt.Errorf("mutation target %q disappeared from semantic IR", target)
	}
	field := tail(model.MutationField)
	payload := tail(model.MutationPayload)
	intent := tail(model.MutationIntent)
	locality := tail(model.MutationLocality)
	fieldHash := ""
	if field != "id" {
		var fieldErr error
		fieldHash, fieldErr = semantic.NodeFieldHash(node, field)
		if fieldErr != nil {
			return Attempt{}, false, "UNKNOWN", "", "", "", fmt.Errorf("derive mutation field: %w", fieldErr)
		}
	}
	beforeSemantic := ir.StableHash()
	beforeGraph := ir.Graph.StableHash()
	request := semantic.GraphPatchRequest{
		SchemaVersion: semantic.GraphPatchSchemaVersion, Operation: semantic.GraphPatchSetNodeField,
		ExpectedGraphHash: beforeGraph, NodeID: node.ID, ExpectedNodeHash: node.StableHash(), Field: field, ExpectedFieldHash: fieldHash,
		ExpectedSourceDigest: semanticDigest, ExpectedIRDigest: semanticDigest,
		AllowedIntent: intent, Locality: locality,
	}
	base := semantic.GraphPatchBase{SourceDigest: semanticDigest, IRDigest: semanticDigest}
	patched, callErr := ir.Graph.ApplyGraphPatch(base, request, semantic.GraphPatchMutation{Name: payload})
	originalSemanticAfter := ir.StableHash()
	originalGraphAfter := ir.Graph.StableHash()
	api := "semantic.Graph.ApplyGraphPatch"
	attempt := Attempt{
		ID: id, Class: classForAttempt(id, claims), Operation: "mutate", Root: operation.ID.String(), Relation: "set", Target: target.String(),
		MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName,
		ProofChoice: proofForAttempt(id, claims), Stage: "MUTATION_BOUNDARY", Step: "apply-typed-request", API: api,
		MutationField: field, MutationPayload: payload, MutationIntent: intent, MutationLocality: locality,
		SemanticDigestBefore: beforeSemantic, SemanticDigestAfter: originalSemanticAfter,
		GraphDigestBefore: beforeGraph, GraphDigestAfter: originalGraphAfter,
		OriginalSemanticDigestAfter: originalSemanticAfter, OriginalGraphDigestAfter: originalGraphAfter,
	}
	if callErr != nil {
		attempt.APIErrorCode = mutationErrorCode(callErr)
		attempt.APIError = callErr.Error()
		if originalSemanticAfter != beforeSemantic || originalGraphAfter != beforeGraph {
			attempt.APIOutcome = "ERROR"
			attempt.Decision, attempt.Resolution, attempt.Reason = "REFUTED", "EXACT", "MUTATION_API_CHANGED_OR_PARTIALLY_MUTATED"
			return attempt, false, "UNKNOWN", api, attempt.APIOutcome, attempt.APIError, nil
		}
		var conflict semantic.GraphPatchConflict
		if errors.As(callErr, &conflict) && conflict.Code == semantic.PatchImmutableField && conflict.Detail == field {
			attempt.APIOutcome = "REJECTED"
			attempt.Decision, attempt.Resolution, attempt.Reason = "DENIED", "EXACT_REJECTION", "IMMUTABLE_ID_PATCH_REJECTED"
			return attempt, false, "UNKNOWN", api, attempt.APIOutcome, attempt.APIError, nil
		}
		attempt.APIOutcome = "ERROR"
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "MUTATION_API_ERROR_"+strings.ToUpper(attempt.APIErrorCode)
		return attempt, false, "UNKNOWN", api, attempt.APIOutcome, attempt.APIError, nil
	}
	patchedIR := ir
	patchedIR.Graph = patched
	attempt.ReturnedSemanticDigest, attempt.ReturnedGraphDigest = patchedIR.StableHash(), patched.StableHash()
	attempt.Decision, attempt.Resolution, attempt.Reason, attempt.APIOutcome = "REFUTED", "EXACT", "DETACHED_GRAPH_PATCH_ACCEPTED", "ACCEPTED"
	return attempt, field == "id", "OBSERVED", api, attempt.APIOutcome, "", nil
}

func tail(id semantic.ID) string {
	parts := strings.Split(strings.TrimSuffix(id.String(), "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func mutationErrorCode(err error) string {
	var conflict semantic.GraphPatchConflict
	if errors.As(err, &conflict) {
		return string(conflict.Code)
	}
	return "unknown"
}

func unknownAttempt(graph *query.Graph, operation operationSpec, target query.ID, semanticDigest, graphDigest string, claims []claimSpec) Attempt {
	before := graph.StableHash()
	attempt := Attempt{
		ID: "unknown.target", Class: classForAttempt("unknown.target", claims), Operation: "query", Root: operation.ID.String(), Relation: "used", Target: target.String(),
		MetaOperation: metaForAttempt("unknown.target", operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName,
		ProofChoice: proofForAttempt("unknown.target", claims), Stage: "UNKNOWN", Step: "resolve-unknown-subject", SemanticDigestBefore: semanticDigest, GraphDigestBefore: before,
	}
	_, err := graph.ExactMatch(query.NewExactQuery(query.ID(operation.ID.String()), query.Used, target))
	attempt.SemanticDigestAfter, attempt.GraphDigestAfter = semanticDigest, graph.StableHash()
	attempt.ObservedMaterialDigest = attempt.GraphDigestAfter
	if err != nil && errors.Is(err, query.ErrUnknownEndpoint) {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "UNKNOWN_TARGET"
	} else if err != nil {
		attempt.Decision, attempt.Resolution, attempt.Reason = "UNKNOWN", "LOWER_RESOLUTION", "QUERY_API_ERROR"
	} else {
		attempt.Decision, attempt.Resolution, attempt.Reason = "REFUTED", "EXACT", "UNKNOWN_TARGET_BECAME_KNOWN"
	}
	return attempt
}

func proofForAttempt(id string, claims []claimSpec) string {
	for _, claim := range claims {
		if claim.EvidenceAttempt == id {
			return claim.ProofChoice
		}
	}
	return "SOURCE_DERIVED"
}

func classForAttempt(id string, claims []claimSpec) string {
	for _, claim := range claims {
		if claim.EvidenceAttempt == id {
			return claim.Class
		}
	}
	return "SOURCE_DERIVED"
}

func metaForAttempt(id, fallback string, claims []claimSpec) string {
	for _, claim := range claims {
		if claim.EvidenceAttempt == id {
			return claim.MetaOperation
		}
	}
	return fallback
}

func findOperation(operations []operationSpec, program string) (operationSpec, bool) {
	for _, operation := range operations {
		if operation.Program == program {
			return operation, true
		}
	}
	return operationSpec{}, false
}

func attemptIDForProgram(program string) string {
	switch {
	case strings.HasPrefix(program, "reflect.query:"):
		return "reflect." + strings.TrimPrefix(program, "reflect.query:")
	case strings.HasPrefix(program, "reflect.attempt:"):
		return "mutation.attempt"
	case strings.HasPrefix(program, "reflect.observation:"):
		if strings.HasSuffix(program, ":net-repository-status-unchanged") {
			return "repository.net-status-unchanged"
		}
		return "receipt.seal"
	default:
		return program
	}
}

func observeEffects(beforePath, afterPath string) Effects {
	base := Effects{RepositoryObservation: "UNOBSERVED", OverallAuthority: "UNKNOWN", DetachedGraphPatchCapability: "UNKNOWN"}
	if beforePath == "" || afterPath == "" {
		base.RepositoryObservationStage, base.RepositoryObservationStep, base.RepositoryObservationReason = "REPOSITORY", "read-status", "REPOSITORY_EVIDENCE_MISSING"
		return base
	}
	before, err := readStatusLines(beforePath)
	if err != nil {
		base.RepositoryObservationStage, base.RepositoryObservationStep, base.RepositoryObservationReason = "REPOSITORY", "read-status", "REPOSITORY_EVIDENCE_UNAVAILABLE"
		return base
	}
	after, err := readStatusLines(afterPath)
	if err != nil {
		base.RepositoryObservationStage, base.RepositoryObservationStep, base.RepositoryObservationReason = "REPOSITORY", "read-status", "REPOSITORY_EVIDENCE_UNAVAILABLE"
		return base
	}
	base.RepositoryStatusBefore, base.RepositoryStatusAfter = before, after
	base.NetRepositoryChanges = changedLines(before, after)
	base.RepositoryEvidenceAvailable = true
	base.RepositoryObservationStage, base.RepositoryObservationStep = "REPOSITORY", "compare-normalized-status"
	if len(base.NetRepositoryChanges) == 0 {
		base.RepositoryObservation, base.RepositoryObservationReason = "net_repository_status_unchanged", "NET_REPOSITORY_STATUS_UNCHANGED"
	} else {
		base.RepositoryObservation, base.RepositoryObservationReason = "net_repository_status_changed", "NET_REPOSITORY_STATUS_CHANGED"
	}
	return base
}

func repositoryAttempt(operation operationSpec, id string, target semantic.ID, semanticDigest string, effects Effects, claims []claimSpec) Attempt {
	value := Attempt{ID: id, Class: classForAttempt(id, claims), Operation: "repository", Root: operation.ID.String(), Relation: "net", Target: target.String(), MetaOperation: metaForAttempt(id, operation.Program, claims), Producer: ProducerName, Consumer: ConsumerName, ProofChoice: proofForAttempt(id, claims), Stage: "REPOSITORY", Step: "compare-normalized-status", SemanticDigestBefore: semanticDigest, SemanticDigestAfter: semanticDigest}
	if !effects.RepositoryEvidenceAvailable {
		value.Step, value.Reason, value.Decision, value.Resolution = "read-status", effects.RepositoryObservationReason, "UNKNOWN", "LOWER_RESOLUTION"
		return value
	}
	material := semantic.StableHashString(strings.Join(effects.RepositoryStatusBefore, "\n") + "\x00" + strings.Join(effects.RepositoryStatusAfter, "\n"))
	value.GraphDigestBefore, value.GraphDigestAfter, value.ObservedMaterialDigest = material, material, material
	if reflectRepositoryNet(effects) {
		value.Decision, value.Resolution, value.Reason = "PASS", "EXACT", "NET_REPOSITORY_STATUS_UNCHANGED"
	} else {
		value.Decision, value.Resolution, value.Reason = "REFUTED", "EXACT", "NET_REPOSITORY_STATUS_CHANGED"
	}
	return value
}

func readStatusLines(path string) ([]string, error) {
	if path == "" {
		return nil, errors.New("repository evidence missing")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read repository status %q: %w", path, err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return []string{}, nil
	}
	return strings.Split(text, "\n"), nil
}

func changedLines(before, after []string) []string {
	left, right := make(map[string]struct{}), make(map[string]struct{})
	for _, line := range before {
		left[line] = struct{}{}
	}
	for _, line := range after {
		right[line] = struct{}{}
	}
	changed := make([]string, 0)
	for line := range left {
		if _, ok := right[line]; !ok {
			changed = append(changed, "BEFORE:"+line)
		}
	}
	for line := range right {
		if _, ok := left[line]; !ok {
			changed = append(changed, "AFTER:"+line)
		}
	}
	sort.Strings(changed)
	return changed
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	text := string(data)
	count := strings.Count(text, "\n")
	if !strings.HasSuffix(text, "\n") {
		count++
	}
	return count
}
