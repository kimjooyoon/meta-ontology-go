package audienceresolutionconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	reportSchema     = "gooo/audience-resolution-consumer/v1"
	ledgerSchema     = "gooo/audience-resolution-ledger/v1"
	sourceKind       = "gooo"
	policyPrefix     = "gooo://audience-resolution/policy/"
	resolutionPrefix = "gooo://audience-resolution/resolution/"
	claimPrefix      = "gooo://audience-resolution/claim-state/"
	relationPrefix   = "gooo://audience-resolution/relation/"
)

type audiencePolicy struct {
	Audience    string
	Resolution  string
	Coordinates []string
}

type sourceModel struct {
	Digest      string
	IRDigest    string
	Denominator int
	Audiences   []audiencePolicy
	ClaimStates map[string]bool
	Relation    bool
}

// Check is an independent reconstruction. It accepts a subject receipt as a
// provisional claim, but derives every audience-local value from raw source,
// raw recipes, and artifact bytes again.
func Check(input Input) Report {
	report := Report{Schema: reportSchema}
	report.RawLedgerFinalFieldsAbsent = rawLedgerHasNoFinalFields(input.LedgerBytes)
	report.RawEvidenceHistoricalOnly = rawEvidenceIsHistorical(input.LedgerBytes, input.Ledger)
	report.ReceiptDigestMatch = receiptDigestMatches(input.ReceiptBytes, input.Receipt.Digest)
	report.ProducerImports = auditProducerImports(input.RepoRoot)

	model, err := reconstruct(input.SourcePath, input.Source)
	if err != nil {
		return sealReport(withIssue(report, "SOURCE_RECONSTRUCTION_UNAVAILABLE"))
	}
	report.SourceReconstruction = SourceReconstruction{
		ParsedAndLowered:  true,
		DeclarationCount:  model.Denominator,
		SemanticDigest:    model.Digest,
		CanonicalIRDigest: model.IRDigest,
		ReceiptMatches: input.Receipt.Source.SemanticDigest == model.Digest &&
			input.Receipt.Source.DeclarationCount == model.Denominator && input.Receipt.Source.Reconstructed,
	}

	values, contradictions, evidenceIssues := independentlyObserve(input, model)
	report.CurrentEvidenceCounts = countEvidence(input.Ledger, input.Receipt.CurrentEvidence)
	report.DistinctPropositions = len(sourceCoordinates(model))
	declaredReplay := input.Receipt.Replay
	report.Replay = verifyReplay(input)
	if declaredReplay.RunADigest != report.Replay.RunADigest || declaredReplay.RunBDigest != report.Replay.RunBDigest ||
		declaredReplay.Equal != report.Replay.Equal || declaredReplay.CombinedDigest != report.Replay.CombinedDigest {
		evidenceIssues = append(evidenceIssues, "REPLAY_RECEIPT_DIGEST_MISMATCH")
	}
	report.CounterexamplesChecked, evidenceIssues = verifyCounterexamples(input, model, evidenceIssues)
	report.Audiences = reconstructAudiences(model, input.Ledger, input.Receipt, values, contradictions)
	report.ClaimTransitionsChecked, evidenceIssues = verifyTransitions(input, model, values, contradictions, evidenceIssues)
	expectedTransitions := len(model.Audiences) * len(sourceCoordinates(model))
	report.Attestation = buildAttestation(input, model, report.ClaimTransitionsChecked == expectedTransitions && len(evidenceIssues) == 0)

	issues := append([]string{}, evidenceIssues...)
	if input.Ledger.Schema != ledgerSchema || input.Ledger.Subject == "" || input.Ledger.Source.Kind != sourceKind ||
		input.Ledger.Source.Digest != rawDigest(input.Source) {
		issues = append(issues, "RAW_SOURCE_BINDING_INVALID")
	}
	if input.Ledger.Source.SemanticDigest != "" && input.Ledger.Source.SemanticDigest != model.Digest {
		issues = append(issues, "RAW_SEMANTIC_DIGEST_INVALID")
	}
	if !report.RawLedgerFinalFieldsAbsent {
		issues = append(issues, "RAW_LEDGER_CONTAINS_FINAL_FIELD")
	}
	if !report.RawEvidenceHistoricalOnly {
		issues = append(issues, "RAW_LEDGER_CONTAINS_CURRENT_OBSERVATION")
	}
	if !report.ReceiptDigestMatch || !report.SourceReconstruction.ReceiptMatches {
		issues = append(issues, "RECEIPT_SOURCE_RECONSTRUCTION_MISMATCH")
	}
	if report.ProducerImports.Numerator != 0 {
		issues = append(issues, "CONSUMER_IMPORTS_PRODUCER")
	}
	if report.Replay.RunAPath == "" || report.Replay.RunBPath == "" || !report.Replay.Equal || report.Replay.RunAPath == report.Replay.RunBPath {
		issues = append(issues, "REPLAY_ARTIFACTS_NOT_INDEPENDENT")
	}
	if len(report.Audiences) != 3 || len(input.Receipt.Views) != 3 {
		issues = append(issues, "AUDIENCE_PROJECTION_UNAVAILABLE")
	} else {
		for index, audience := range report.Audiences {
			view := input.Receipt.Views[index]
			if audience.Audience != view.Audience || audience.Visible != view.Visible || audience.Required != view.Required ||
				audience.Decision != view.LocalDecision || audience.Resolution != view.LocalResolution ||
				audience.SubjectSatisfied != view.SubjectSatisfied || audience.SubjectRequired != view.SubjectRequired {
				issues = append(issues, "AUDIENCE_LOCAL_DECISION_MISMATCH")
			}
		}
	}
	expectedDecision, expectedResolution := subjectDecision(model, values, contradictions)
	if input.Receipt.Decision != expectedDecision || input.Receipt.Resolution != expectedResolution {
		issues = append(issues, "GLOBAL_DECISION_RECONSTRUCTION_MISMATCH")
	}
	if input.Receipt.Summary.DistinctPropositions != report.DistinctPropositions || report.DistinctPropositions <= 0 {
		issues = append(issues, "DISTINCT_PROPOSITION_DENOMINATOR_INVALID")
	}
	if len(input.Receipt.ClaimTransitions) != report.ClaimTransitionsChecked || report.ClaimTransitionsChecked != expectedTransitions {
		issues = append(issues, "CLAIM_TRANSITION_COUNT_MISMATCH")
	}
	if input.Receipt.Summary.SourceDenominator != model.Denominator || input.Receipt.Summary.Coordinates.Total != len(subjectCoordinates(model)) {
		issues = append(issues, "SOURCE_DERIVED_DENOMINATOR_INVALID")
	}
	if !report.AttestationValid(input.Receipt.Digest) {
		issues = append(issues, "VERIFICATION_ATTESTATION_INVALID")
	}
	if len(issues) == 0 {
		report.Decision, report.Reason = "PASS", "INDEPENDENT_RAW_ARTIFACT_RECONSTRUCTION_MATCHED"
	} else {
		report.Decision, report.Reason = "REFUTED", strings.Join(unique(issues), ";")
	}
	return sealReport(report)
}

func (report Report) AttestationValid(subjectDigest string) bool {
	attestation := report.Attestation
	if attestation.Schema == "" || attestation.SubjectReceiptDigest != subjectDigest || attestation.Decision != "PASS" ||
		attestation.EvidenceStatus != "CURRENT_EVIDENCE" || !validDigest(attestation.Evidence.ContentDigest) ||
		attestation.Evidence.ArtifactPath == "" || attestation.Evidence.ObservedValue != "true" || attestation.ClaimTransition.After != "DISCHARGED" ||
		!validDigest(attestation.Digest) {
		return false
	}
	copy := attestation
	copy.Digest = ""
	return digestBytes(canonicalJSON(copy)) == attestation.Digest
}

func reconstruct(filename string, source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceModel{}, fmt.Errorf("parse diagnostics: %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, err
	}
	policies := map[string][]struct {
		ordinal    int
		coordinate string
	}{}
	resolutions := map[string]string{}
	claims := map[string]bool{}
	relation := false
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		switch {
		case strings.HasPrefix(id, policyPrefix):
			parts := strings.Split(strings.TrimPrefix(id, policyPrefix), "/")
			if len(parts) < 3 {
				return sourceModel{}, fmt.Errorf("incomplete policy identity")
			}
			ordinal, err := strconv.Atoi(parts[1])
			if err != nil {
				return sourceModel{}, err
			}
			policies[parts[0]] = append(policies[parts[0]], struct {
				ordinal    int
				coordinate string
			}{ordinal, strings.Join(parts[2:], "/")})
		case strings.HasPrefix(id, resolutionPrefix):
			parts := strings.Split(strings.TrimPrefix(id, resolutionPrefix), "/")
			if len(parts) < 2 {
				return sourceModel{}, fmt.Errorf("incomplete resolution identity")
			}
			resolutions[parts[0]] = strings.Join(parts[1:], "/")
		case strings.HasPrefix(id, claimPrefix):
			claims[strings.TrimPrefix(id, claimPrefix)] = true
		case id == relationPrefix+"evidence-to-claim":
			relation = true
		}
	}
	canonical := []byte(ir.SemanticCanonical())
	model := sourceModel{Digest: digestBytes(canonical), IRDigest: digestBytes(canonical), Denominator: len(ir.Graph.Nodes()), ClaimStates: claims, Relation: relation}
	for _, audience := range []string{"USER", "TOOL_AUTHOR", "GOVERNOR"} {
		values := policies[audience]
		sort.Slice(values, func(i, j int) bool {
			if values[i].ordinal != values[j].ordinal {
				return values[i].ordinal < values[j].ordinal
			}
			return values[i].coordinate < values[j].coordinate
		})
		if len(values) == 0 || resolutions[audience] == "" {
			return sourceModel{}, fmt.Errorf("audience policy incomplete")
		}
		coordinates := make([]string, 0, len(values))
		for _, value := range values {
			coordinates = append(coordinates, value.coordinate)
		}
		model.Audiences = append(model.Audiences, audiencePolicy{Audience: audience, Resolution: resolutions[audience], Coordinates: coordinates})
	}
	if !nested(model.Audiences) || !claims["OPEN"] || !claims["DISCHARGED"] || !claims["REFUTED"] || !relation {
		return sourceModel{}, fmt.Errorf("formal policy, claim state, or relation is incomplete")
	}
	return model, nil
}

func sourceCoordinates(model sourceModel) []string {
	if len(model.Audiences) != 3 {
		return nil
	}
	return append([]string(nil), model.Audiences[2].Coordinates...)
}

func subjectCoordinates(model sourceModel) map[string]bool {
	result := map[string]bool{}
	for _, coordinate := range sourceCoordinates(model) {
		if coordinate != "projection.shared-decision" && coordinate != "receipt.seal" {
			result[coordinate] = true
		}
	}
	return result
}

func audienceCoordinates(model sourceModel, audience string) []string {
	for _, value := range model.Audiences {
		if value.Audience == audience {
			return value.Coordinates
		}
	}
	return nil
}

func nested(values []audiencePolicy) bool {
	if len(values) != 3 {
		return false
	}
	for index := 1; index < len(values); index++ {
		if len(values[index].Coordinates) <= len(values[index-1].Coordinates) {
			return false
		}
		for coordinateIndex, coordinate := range values[index-1].Coordinates {
			if values[index].Coordinates[coordinateIndex] != coordinate {
				return false
			}
		}
	}
	return true
}

func independentlyObserve(input Input, model sourceModel) (map[string]bool, map[string]bool, []string) {
	values := map[string]bool{}
	contradictions := map[string]bool{}
	issues := []string{}
	recipes, duplicate := recordMap(input.Ledger.Records)
	if duplicate {
		issues = append(issues, "RAW_RECIPE_DUPLICATE")
	}
	for _, current := range input.Receipt.CurrentEvidence {
		recipe, ok := recipes[current.Coordinate]
		if !ok {
			issues = append(issues, "CURRENT_EVIDENCE_HAS_UNKNOWN_COORDINATE:"+current.Coordinate)
			continue
		}
		if current.ClaimID != recipe.ClaimID || current.Proposition != recipe.Proposition || current.PropositionDigest != recipe.PropositionDigest ||
			current.TargetAddress != recipe.TargetAddress || current.Provider == "" || current.ArtifactPath == "" || current.PriorClaim != recipe.PriorClaim ||
			current.Producer != recipe.Producer || current.Consumer != recipe.Consumer || current.MetaOperation != recipe.MetaOperation ||
			current.ProofChoice != recipe.ProofChoice || current.Stage != recipe.Stage || current.Step != recipe.Step || current.Reason != recipe.Reason {
			issues = append(issues, "CURRENT_EVIDENCE_RECIPE_RELATION_MISMATCH:"+current.Coordinate)
		}
		if current.EvidenceStatus == "UNKNOWN" {
			values[current.Coordinate] = false
			continue
		}
		if current.EvidenceStatus != "CURRENT_EVIDENCE" {
			issues = append(issues, "CURRENT_EVIDENCE_STATUS_INVALID:"+current.Coordinate)
			continue
		}
		artifact, err := readArtifact(input.ArtifactRoot, current.ArtifactPath)
		if err != nil {
			issues = append(issues, "CURRENT_ARTIFACT_UNAVAILABLE:"+current.Coordinate)
			continue
		}
		if digestBytes(canonicalJSON(artifact)) != current.ContentDigest {
			issues = append(issues, "CURRENT_ARTIFACT_DIGEST_MISMATCH:"+current.Coordinate)
		}
		if current.Coordinate == "ledger.replay" {
			if len(current.ArtifactPaths) != 2 || len(current.ContentDigests) != 2 || current.ArtifactPaths[0] == current.ArtifactPaths[1] {
				issues = append(issues, "REPLAY_EVIDENCE_PATHS_INVALID")
			} else {
				for index, path := range current.ArtifactPaths {
					run, runErr := readArtifact(input.ArtifactRoot, path)
					if runErr != nil || digestBytes(canonicalJSON(run)) != current.ContentDigests[index] {
						issues = append(issues, "REPLAY_EVIDENCE_DIGEST_INVALID")
					}
				}
			}
		}
		if stringValue(artifact["coordinate"]) != current.Coordinate || stringValue(artifact["proposition_digest"]) != current.PropositionDigest ||
			stringValue(artifact["target_address"]) != current.TargetAddress || stringValue(artifact["evidence_status"]) != "CURRENT_EVIDENCE" {
			issues = append(issues, "CURRENT_ARTIFACT_BINDING_INVALID:"+current.Coordinate)
		}
		observed, contradiction, ok := observeArtifactPredicate(current.Coordinate, artifact, input, model, recipes)
		if !ok {
			issues = append(issues, "CURRENT_PREDICATE_UNRECONSTRUCTABLE:"+current.Coordinate)
			continue
		}
		values[current.Coordinate] = observed
		if contradiction {
			contradictions[current.Coordinate] = true
		}
		if stringValue(artifact["observed_value"]) != boolString(observed) || current.ObservedValue != boolString(observed) {
			issues = append(issues, "CURRENT_PREDICATE_VALUE_MISMATCH:"+current.Coordinate)
		}
	}
	return values, contradictions, unique(issues)
}

func observeArtifactPredicate(coordinate string, artifact map[string]any, input Input, model sourceModel, recipes map[string]RawRecord) (bool, bool, bool) {
	details, _ := artifact["details"].(map[string]any)
	switch coordinate {
	case "source.binding":
		sourcePath, sourceErr := readRawArtifact(input.ArtifactRoot, stringValue(details["source_artifact_path"]))
		return input.Ledger.Source.Path != "" && input.Ledger.Source.Kind == sourceKind && input.Ledger.Source.Digest == rawDigest(input.Source) &&
			sourceErr == nil && rawDigest(sourcePath) == stringValue(details["source_artifact_digest"]) && string(sourcePath) == string(input.Source) &&
			stringValue(details["semantic_digest"]) == model.Digest && intValue(details["declaration_count"]) == model.Denominator, false, true
	case "ledger.coverage":
		coords := sourceCoordinates(model)
		seen := map[string]bool{}
		for _, record := range input.Ledger.Records {
			seen[record.Coordinate] = true
		}
		valid := len(input.Ledger.Records) == len(coords)
		for _, coordinate := range coords {
			valid = valid && seen[coordinate]
		}
		contradiction := stringValue(artifact["observed_predicate"]) == "forced_contradiction" && stringValue(details["mutation"]) == "CONTRADICTORY"
		return valid && !contradiction, contradiction, true
	case "ledger.replay":
		replay := input.Receipt.Replay
		a, aErr := readArtifact(input.ArtifactRoot, replay.RunAPath)
		b, bErr := readArtifact(input.ArtifactRoot, replay.RunBPath)
		if aErr != nil || bErr != nil || replay.RunAPath == replay.RunBPath {
			return false, false, false
		}
		return digestBytes(canonicalJSON(a)) == replay.RunADigest && digestBytes(canonicalJSON(b)) == replay.RunBDigest && replay.Equal && string(canonicalJSON(a)) == string(canonicalJSON(b)), false, true
	case "user.coordinates", "author.coordinates", "governor.coordinates":
		audience := map[string]string{"user.coordinates": "USER", "author.coordinates": "TOOL_AUTHOR", "governor.coordinates": "GOVERNOR"}[coordinate]
		valid := len(audienceCoordinates(model, audience)) > 0
		for _, value := range audienceCoordinates(model, audience) {
			valid = valid && recipes[value].Coordinate == value
		}
		return valid, false, true
	case "projection.nesting":
		return nested(model.Audiences), false, true
	case "projection.resolution":
		return len(model.Audiences) == 3 && model.Audiences[0].Resolution == "USER_VISIBLE_COORDINATES" && model.Audiences[1].Resolution == "TOOL_CONTRACT_COORDINATES" && model.Audiences[2].Resolution == "GOVERNOR_FULL_LEDGER", false, true
	case "projection.shared-decision":
		return input.Receipt.Decision == "PASS", false, true
	case "counterexample.omission", "counterexample.contradiction":
		for _, value := range input.Receipt.Counterexamples {
			if (coordinate == "counterexample.omission" && value.ID == "counterexample.missing-information") ||
				(coordinate == "counterexample.contradiction" && value.ID == "counterexample.decision-contradiction") {
				return value.ExecutionValidated, false, true
			}
		}
		return false, false, true
	case "receipt.seal":
		return false, false, true
	default:
		return false, false, false
	}
}

func reconstructAudiences(model sourceModel, ledger RawLedger, receipt Receipt, values map[string]bool, contradictions map[string]bool) []AudienceCheck {
	result := make([]AudienceCheck, 0, 3)
	subject := subjectCoordinates(model)
	for _, policy := range model.Audiences {
		visible, subjectSatisfied := 0, 0
		for _, coordinate := range policy.Coordinates {
			if _, ok := receiptRecord(receipt.CurrentEvidence, coordinate); ok {
				visible++
			}
			if subject[coordinate] && values[coordinate] {
				subjectSatisfied++
			}
		}
		decision, resolution := "PASS", "EXACT"
		for coordinate := range subject {
			if contains(policy.Coordinates, coordinate) && contradictions[coordinate] {
				decision, resolution = "REFUTED", "INVARIANT_ONLY"
				break
			}
		}
		if decision != "REFUTED" && subjectSatisfied != len(subject) {
			decision, resolution = "UNKNOWN", "LOWER_RESOLUTION"
		}
		result = append(result, AudienceCheck{Audience: policy.Audience, Visible: visible, Required: len(sourceCoordinates(model)), Decision: decision, Resolution: resolution, SubjectSatisfied: subjectSatisfied, SubjectRequired: len(subject)})
	}
	return result
}

func verifyReplay(input Input) ReplayVerification {
	replay := input.Receipt.Replay
	a, aErr := readArtifact(input.ArtifactRoot, replay.RunAPath)
	b, bErr := readArtifact(input.ArtifactRoot, replay.RunBPath)
	if aErr != nil || bErr != nil {
		return replay
	}
	replay.RunADigest = digestBytes(canonicalJSON(a))
	replay.RunBDigest = digestBytes(canonicalJSON(b))
	replay.Equal = replay.RunADigest == replay.RunBDigest && string(canonicalJSON(a)) == string(canonicalJSON(b))
	replay.CombinedDigest = digestBytes(append(append([]byte{}, canonicalJSON(a)...), canonicalJSON(b)...))
	return replay
}

func verifyCounterexamples(input Input, model sourceModel, issues []string) (int, []string) {
	seen := map[string]bool{}
	for _, value := range input.Receipt.Counterexamples {
		seen[value.ID] = true
		artifact, err := readArtifact(input.ArtifactRoot, value.ArtifactPath)
		if err != nil || digestBytes(canonicalJSON(artifact)) != value.ContentDigest {
			issues = append(issues, "COUNTEREXAMPLE_ARTIFACT_INVALID:"+value.ID)
			continue
		}
		expectedAfter := "OPEN"
		expectedDecision, expectedResolution := "UNKNOWN", "LOWER_RESOLUTION"
		if value.Kind == "DECISION_CONTRADICTION" {
			expectedAfter, expectedDecision, expectedResolution = "REFUTED", "REFUTED", "INVARIANT_ONLY"
		}
		if value.TargetCoordinate == "" || value.PropositionDigest == "" || value.Global != expectedDecision || value.Resolution != expectedResolution ||
			value.BeforeClaim != "OPEN" || value.AfterClaim != expectedAfter || !value.ExecutionValidated ||
			stringValue(artifact["target_coordinate"]) != value.TargetCoordinate || stringValue(artifact["proposition_digest"]) != value.PropositionDigest ||
			stringValue(artifact["target_address"]) != value.TargetAddress || stringValue(artifact["global_decision"]) != value.Global || stringValue(artifact["resolution"]) != value.Resolution ||
			stringValue(artifact["stage"]) != value.Stage || stringValue(artifact["step"]) != value.Step || stringValue(artifact["reason"]) != value.Reason || !boolValue(artifact["execution_validated"]) {
			issues = append(issues, "COUNTEREXAMPLE_EXECUTION_NOT_VALIDATED:"+value.ID)
		}
		if len(value.Views) != len(model.Audiences) {
			issues = append(issues, "COUNTEREXAMPLE_AUDIENCE_VIEWS_MISSING:"+value.ID)
		}
	}
	if !seen["counterexample.missing-information"] || !seen["counterexample.decision-contradiction"] {
		issues = append(issues, "COUNTEREXAMPLE_SET_INCOMPLETE")
	}
	return len(input.Receipt.Counterexamples), unique(issues)
}

func verifyTransitions(input Input, model sourceModel, values, contradictions map[string]bool, issues []string) (int, []string) {
	recipes, duplicate := recordMap(input.Ledger.Records)
	if duplicate {
		issues = append(issues, "RAW_RECIPE_DUPLICATE")
	}
	previous := digestBytes([]byte("gooo://audience-resolution/claim-event/genesis"))
	index := 0
	distinct := map[string]bool{}
	for _, audience := range model.Audiences {
		for _, coordinate := range sourceCoordinates(model) {
			if index >= len(input.Receipt.ClaimTransitions) {
				return index, unique(append(issues, "CLAIM_TRANSITION_CHAIN_SHORT"))
			}
			actual := input.Receipt.ClaimTransitions[index]
			recipe := recipes[coordinate]
			if recipe.Coordinate == "" {
				recipe = fallbackRecipe(coordinate)
			}
			visible := contains(audience.Coordinates, coordinate) && recipe.Coordinate == coordinate && hasCurrent(input.Receipt.CurrentEvidence, coordinate)
			before := recipe.PriorClaim
			if before == "" {
				before = "OPEN"
			}
			after := "OPEN"
			if visible && contradictions[coordinate] {
				after = "REFUTED"
			} else if visible && values[coordinate] {
				after = "DISCHARGED"
			}
			status, evidenceDigest := "UNKNOWN", digestBytes([]byte("unobserved:"+coordinate))
			current := findCurrent(input.Receipt.CurrentEvidence, coordinate)
			if current.Coordinate != "" {
				evidenceDigest = current.ContentDigest
				if visible {
					status = current.EvidenceStatus
				}
				if evidenceDigest == "" {
					evidenceDigest = digestBytes([]byte("unobserved:" + coordinate))
				}
			}
			claimID := recipe.ClaimID
			if claimID == "" {
				claimID = "claim/" + coordinate
			}
			producer, consumer, metaOperation, proofChoice := recipe.Producer, recipe.Consumer, recipe.MetaOperation, recipe.ProofChoice
			if producer == "" {
				producer = "audience-resolution.policy"
			}
			if consumer == "" {
				consumer = audience.Audience
			}
			if metaOperation == "" {
				metaOperation = "project-audience-claim"
			}
			if proofChoice == "" {
				proofChoice = "COHERENCE"
			}
			expected := ReceiptTransition{ClaimID: claimID + "/audience/" + audience.Audience, Audience: audience.Audience, IndicatorID: coordinate,
				Proposition: recipe.Proposition, PropositionDigest: recipe.PropositionDigest, TargetAddress: recipe.TargetAddress, Before: before, After: after,
				Visibility: visibility(visible), EvidenceStatus: status, EvidenceDigest: evidenceDigest, PreviousEventDigest: previous,
				SourceDigest: input.Ledger.Source.Digest, SemanticSourceDigest: model.Digest, Producer: producer, Consumer: consumer,
				MetaOperation: metaOperation, ProofChoice: proofChoice, Stage: recipe.Stage, Step: recipe.Step, Reason: recipe.Reason}
			expected.EventDigest = transitionDigest(expected)
			if actual != expected {
				issues = append(issues, "CLAIM_TRANSITION_MISMATCH:"+strconv.Itoa(index))
			}
			if recipe.PropositionDigest != "" {
				distinct[recipe.PropositionDigest] = true
			}
			previous = actual.EventDigest
			index++
		}
	}
	if len(distinct) != len(sourceCoordinates(model)) {
		issues = append(issues, "CLAIM_PROPOSITION_DENOMINATOR_INVALID")
	}
	if index != len(input.Receipt.ClaimTransitions) {
		issues = append(issues, "CLAIM_TRANSITION_CHAIN_EXTRA_EVENTS")
	}
	return index, unique(issues)
}

func buildAttestation(input Input, model sourceModel, valid bool) Attestation {
	recipe := findRaw(input.Ledger.Records, "receipt.seal")
	path := "current/receipt.seal.attestation-evidence.json"
	evidencePayload := map[string]any{"schema": "gooo/audience-resolution/attestation-evidence/v1", "subject_receipt_digest": input.Receipt.Digest,
		"receipt_digest_recomputed": valid, "observed_predicate": "receipt_digest_recomputed", "observed_value": boolString(valid), "evidence_status": "CURRENT_EVIDENCE"}
	evidenceDigest := digestBytes(canonicalJSON(evidencePayload))
	transition := ReceiptTransition{ClaimID: "claim/receipt.seal/conformance", Audience: "CONFORMANCE", IndicatorID: "receipt.seal",
		Proposition: recipe.Proposition, PropositionDigest: recipe.PropositionDigest, TargetAddress: recipe.TargetAddress, Before: "OPEN", After: "DISCHARGED", Visibility: "VISIBLE",
		EvidenceStatus: "CURRENT_EVIDENCE", EvidenceDigest: evidenceDigest, PreviousEventDigest: finalEvent(input.Receipt.ClaimTransitions), SourceDigest: input.Ledger.Source.Digest,
		SemanticSourceDigest: model.Digest, Producer: "independent.checker", Consumer: "audience-resolution.conformance", MetaOperation: "verify-receipt-seal",
		ProofChoice: "REGRESSION", Stage: "receipt", Step: "seal", Reason: "receipt digest independently recomputed after subject evaluation"}
	transition.EventDigest = transitionDigest(transition)
	attestation := Attestation{Schema: "gooo/audience-resolution-verification-attestation/v1", SubjectReceiptDigest: input.Receipt.Digest,
		Decision: "PASS", Resolution: "EXACT", Stage: "receipt", Step: "seal", Reason: "receipt digest independently recomputed after subject evaluation",
		EvidenceStatus: "CURRENT_EVIDENCE", Evidence: EvidenceRecord{ID: "receipt.seal.attestation", Coordinate: "receipt.seal", Audience: "independent.checker",
			ClaimID: recipe.ClaimID, Proposition: recipe.Proposition, PropositionDigest: recipe.PropositionDigest, TargetAddress: recipe.TargetAddress,
			Provider: "independent.checker", ArtifactPath: path, ContentDigest: evidenceDigest, ObservedPredicate: "receipt_digest_recomputed", ObservedValue: boolString(valid), EvidenceStatus: "CURRENT_EVIDENCE",
			Producer: "independent.checker", Consumer: "audience-resolution.conformance", MetaOperation: "verify-receipt-seal", ProofChoice: "REGRESSION", Stage: "receipt", Step: "seal", Reason: "receipt digest independently recomputed after subject evaluation", PriorClaim: "OPEN"}, ClaimTransition: transition}
	attestation.Digest = digestBytes(canonicalJSON(attestation))
	return attestation
}

// AttestationEvidencePayload is the exact raw artifact sealed by the report.
func AttestationEvidencePayload(attestation Attestation) []byte {
	return canonicalJSON(map[string]any{"schema": "gooo/audience-resolution/attestation-evidence/v1", "subject_receipt_digest": attestation.SubjectReceiptDigest,
		"receipt_digest_recomputed": attestation.Evidence.ObservedValue == "true", "observed_predicate": attestation.Evidence.ObservedPredicate,
		"observed_value": attestation.Evidence.ObservedValue, "evidence_status": attestation.Evidence.EvidenceStatus})
}

func recordMap(records []RawRecord) (map[string]RawRecord, bool) {
	result, duplicate := map[string]RawRecord{}, false
	for _, record := range records {
		if _, exists := result[record.Coordinate]; exists {
			duplicate = true
		}
		result[record.Coordinate] = record
	}
	return result, duplicate
}

func findRaw(records []RawRecord, coordinate string) RawRecord {
	for _, record := range records {
		if record.Coordinate == coordinate {
			return record
		}
	}
	return RawRecord{Coordinate: coordinate}
}

func fallbackRecipe(coordinate string) RawRecord {
	proposition := "semantic policy coordinate " + coordinate
	return RawRecord{ID: coordinate, Coordinate: coordinate, ClaimID: "claim/" + coordinate, Proposition: proposition,
		PropositionDigest: digestBytes([]byte(proposition)), TargetAddress: "gooo://audience-resolution/claim/" + coordinate,
		Provider: "audience-resolution.policy", Producer: "audience-resolution.policy", Consumer: "audience-resolution.policy",
		MetaOperation: "project-audience-claim", ProofChoice: "COHERENCE", Stage: "projection", Step: "policy",
		Reason: "formal source policy coordinate has no raw recipe", PriorClaim: "OPEN"}
}

func findCurrent(records []EvidenceRecord, coordinate string) EvidenceRecord {
	for _, record := range records {
		if record.Coordinate == coordinate {
			return record
		}
	}
	return EvidenceRecord{}
}

func receiptRecord(records []EvidenceRecord, coordinate string) (EvidenceRecord, bool) {
	value := findCurrent(records, coordinate)
	return value, value.Coordinate != ""
}

func hasCurrent(records []EvidenceRecord, coordinate string) bool {
	value, ok := receiptRecord(records, coordinate)
	return ok && value.EvidenceStatus != "HISTORICAL_FIXTURE"
}

func countEvidence(ledger RawLedger, current []EvidenceRecord) EvidenceCounts {
	counts := EvidenceCounts{}
	for _, record := range ledger.Records {
		if record.EvidenceStatus == "HISTORICAL_FIXTURE" {
			counts.Historical++
		}
	}
	for _, record := range current {
		switch record.EvidenceStatus {
		case "CURRENT_EVIDENCE":
			counts.Current++
		case "HISTORICAL_FIXTURE":
			counts.Historical++
		default:
			counts.Unknown++
		}
	}
	return counts
}

func distinctPropositions(records []RawRecord) int {
	values := map[string]bool{}
	for _, record := range records {
		if record.PropositionDigest != "" {
			values[record.PropositionDigest] = true
		}
	}
	return len(values)
}

func rawLedgerHasNoFinalFields(raw []byte) bool {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	forbidden := map[string]bool{"decision": true, "satisfied": true, "claim_after": true, "blocked": true, "expected_decision": true, "observed_decision": true, "observation": true}
	var visit func(any) bool
	visit = func(current any) bool {
		switch value := current.(type) {
		case map[string]any:
			for key, child := range value {
				if forbidden[key] || !visit(child) {
					return false
				}
			}
		case []any:
			for _, child := range value {
				if !visit(child) {
					return false
				}
			}
		}
		return true
	}
	return visit(value)
}

func rawEvidenceIsHistorical(raw []byte, ledger RawLedger) bool {
	if len(raw) == 0 || !rawLedgerHasNoFinalFields(raw) {
		return false
	}
	for _, record := range ledger.Records {
		if record.EvidenceStatus != "HISTORICAL_FIXTURE" || record.Provider == "" || record.ContentDigest != "" || record.ObservedValue == "OBSERVED" {
			return false
		}
	}
	return len(ledger.Records) > 0
}

func readArtifact(root, relative string) (map[string]any, error) {
	if root == "" || relative == "" {
		return nil, fmt.Errorf("artifact root or path is empty")
	}
	bytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(bytes, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func readRawArtifact(root, relative string) ([]byte, error) {
	if root == "" || relative == "" {
		return nil, fmt.Errorf("artifact root or path is empty")
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
}

func auditProducerImports(root string) ImportAudit {
	directory := filepath.Join(root, "internal", "meta", "audienceresolutionconsumer")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return ImportAudit{Numerator: 1, Denominator: 1, Forbidden: []string{"consumer package unavailable"}}
	}
	audit := ImportAudit{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			audit.Numerator++
			audit.Forbidden = append(audit.Forbidden, entry.Name()+":parse")
			continue
		}
		for _, imported := range file.Imports {
			audit.Denominator++
			pathValue, err := strconv.Unquote(imported.Path.Value)
			if err == nil && (pathValue == "github.com/kimjooyoon/meta-ontology-go/internal/meta/audienceresolution" || pathValue == "github.com/kimjooyoon/meta-ontology-go/cmd/audience-resolution-witness") {
				audit.Numerator++
				audit.Forbidden = append(audit.Forbidden, pathValue)
			}
		}
		if fileHasForbiddenSymbol(file) {
			audit.Numerator++
			audit.Forbidden = append(audit.Forbidden, entry.Name()+":forbidden-symbol")
		}
	}
	sort.Strings(audit.Forbidden)
	return audit
}

func fileHasForbiddenSymbol(file *ast.File) bool {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		found := false
		ast.Inspect(function.Body, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if ok && (identifier.Name == "CanonicalContract" || identifier.Name == "ValidateReceipt" || identifier.Name == "fixedSpecs") {
				found = true
			}
			return !found
		})
		if found {
			return true
		}
	}
	return false
}

func receiptDigestMatches(raw []byte, expected string) bool {
	if expected == "" || len(raw) == 0 {
		return false
	}
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	delete(value, "digest")
	return digestBytes(canonicalJSON(value)) == expected
}

func transitionDigest(value ReceiptTransition) string {
	value.EventDigest = ""
	return digestBytes(canonicalJSON(value))
}

func finalEvent(values []ReceiptTransition) string {
	if len(values) == 0 {
		return digestBytes([]byte("gooo://audience-resolution/claim-event/genesis"))
	}
	return values[len(values)-1].EventDigest
}

func visibility(value bool) string {
	if value {
		return "VISIBLE"
	}
	return "OMITTED"
}

func subjectDecision(model sourceModel, values, contradictions map[string]bool) (string, string) {
	for coordinate := range subjectCoordinates(model) {
		if contradictions[coordinate] {
			return "REFUTED", "INVARIANT_ONLY"
		}
	}
	for coordinate := range subjectCoordinates(model) {
		if !values[coordinate] {
			return "UNKNOWN", "LOWER_RESOLUTION"
		}
	}
	return "PASS", "EXACT"
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}

func intValue(value any) int {
	result, _ := value.(float64)
	return int(result)
}

func boolValue(value any) bool {
	result, _ := value.(bool)
	return result
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, character := range value[7:] {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func rawDigest(value []byte) string { return digestBytes(value) }

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func canonicalJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	var normalized any
	if json.Unmarshal(raw, &normalized) != nil {
		return raw
	}
	result, _ := json.Marshal(normalized)
	return result
}

func sealReport(report Report) Report {
	report.Digest = ""
	report.Digest = digestBytes(canonicalJSON(report))
	return report
}

func withIssue(report Report, reason string) Report {
	report.Decision, report.Reason = "REFUTED", reason
	return report
}
