package claimdependency

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const (
	evidenceProcedure       = "RAW_ARTIFACT_OBSERVATION_BINDING_V3"
	observationSchema       = "gooo.meta.claim-dependency-observation/v2"
	observationBundleSchema = "gooo.meta.claim-dependency-observation-bundle/v1"
	observationProcedure    = "CI_TARGET_BYTES_COMPARISON_V2"
)

// BuildCurrentEvidence keeps the original provider API for callers that do
// not have an external target observation. Parsing/lowering a source is only a
// declared recipe in that case, so the resulting evidence is UNKNOWN.
func BuildCurrentEvidence(artifactPath, operation, capabilityPath, repositoryRoot, outputPath string) (EvidenceReceipt, error) {
	return BuildCurrentEvidenceWithObservation(artifactPath, operation, capabilityPath, repositoryRoot, outputPath, "")
}

// BuildCurrentEvidenceWithObservation binds a current evidence receipt to a
// separately produced observation of the target artifact. Operation is a
// CLAIMED_INPUT/REQUEST and never selects the observed predicate.
func BuildCurrentEvidenceWithObservation(artifactPath, operation, capabilityPath, repositoryRoot, outputPath, observationPath string) (EvidenceReceipt, error) {
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	artifactGraph, err := graphFromSource(artifact, artifactPath)
	if err != nil {
		return EvidenceReceipt{}, fmt.Errorf("provider artifact reconstruction: %w", err)
	}
	capability, err := readCapability(capabilityPath)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	if capability.Status != CurrentEvidence {
		return EvidenceReceipt{}, fmt.Errorf("capability is not current evidence")
	}
	capability.Digest, err = capabilityDigest(capability)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	snapshot, err := repositorySnapshot(repositoryRoot, outputPath)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	if operation != "availability" && operation != "acceptance" && operation != "contradiction" {
		return EvidenceReceipt{}, fmt.Errorf("unknown provider operation %q", operation)
	}
	observations := []ObservationReceipt{}
	observationBundleDigest := ""
	status := UnknownEvidence
	predicate := ObservationUnknown
	observedValue := fmt.Sprintf("observation:ABSENT|stage:OBSERVE|step:current-evidence-provider|reason:EXTERNAL_TARGET_OBSERVATION_MISSING|artifact_path_digest:%s|artifact_bytes_digest:%s", digestBytes([]byte(artifactPath)), digestBytes(artifact))
	if observationPath != "" {
		data, readErr := os.ReadFile(observationPath)
		if readErr != nil {
			return EvidenceReceipt{}, fmt.Errorf("target observation: %w", readErr)
		}
		var bundle ObservationBundle
		if err := json.Unmarshal(data, &bundle); err != nil {
			return EvidenceReceipt{}, fmt.Errorf("target observation decode: %w", err)
		}
		if err := validateObservationBundle(bundle, artifactPath, artifact, artifactGraph.Graph); err != nil {
			return EvidenceReceipt{}, err
		}
		observations = append([]ObservationReceipt(nil), bundle.Observations...)
		observationBundleDigest = bundle.Digest
		if hasClaimOrEdgeObservation(observations) {
			status = CurrentEvidence
		}
		predicate = summaryPredicate(observations)
		observedValue = observationBundleValue(bundle)
	}
	claims := make([]EvidenceClaim, len(artifactGraph.Graph.Nodes))
	for i, claim := range artifactGraph.Graph.Nodes {
		claimObservation, found := claimObservationFor(claim, observations)
		claimStatus, claimPredicate, value, coordinate := UnknownEvidence, ObservationUnknown, absentClaimValue(claim), Coordinate{Stage: "OBSERVE", Step: claim.ActivityName, Reason: "CLAIM_OBSERVATION_MISSING"}
		if found {
			claimStatus, claimPredicate, value, coordinate = CurrentEvidence, claimObservation.ObservedPredicate, claimObservation.ObservedValue, claimObservation.Coordinate
		}
		claims[i] = EvidenceClaim{ClaimID: claim.ClaimID, PropositionDigest: claim.PropositionDigest, ObservedPredicate: claimPredicate, ObservedValue: value, Status: claimStatus, Coordinate: coordinate}
		claims[i].Digest, err = claimEvidenceDigest(claims[i])
		if err != nil {
			return EvidenceReceipt{}, err
		}
	}
	receipt := EvidenceReceipt{Schema: EvidenceSchema, Provider: "github-actions-current-evidence-provider/v3", ArtifactPath: artifactPath, ArtifactBytesDigest: digestBytes(artifact), Operation: operation, RequestStatus: "CLAIMED_INPUT", Procedure: evidenceProcedure, ObservationPath: observationPath, ObservationBundleDigest: observationBundleDigest, Observations: observations, ObservedPredicate: predicate, ObservedValue: observedValue, Status: status, Coordinate: Coordinate{Stage: "OBSERVE", Step: "current-evidence-provider", Reason: observationReason(status, predicate)}, Claims: claims, Capability: capability, Snapshot: snapshot}
	receipt.Digest, err = evidenceReceiptDigest(receipt)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	return receipt, nil
}

func BuildObservationBundle(sourcePath string, source []byte, artifactPath, expectedBytesDigest, output, profile string) (ObservationBundle, error) {
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		return ObservationBundle{}, err
	}
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return ObservationBundle{}, fmt.Errorf("observer source reconstruction: %w", err)
	}
	actualDigest := digestBytes(artifact)
	if profile == "accepted" && expectedBytesDigest != actualDigest {
		return ObservationBundle{}, fmt.Errorf("accepted profile requires matching target bytes")
	}
	if (profile == "contradiction" || profile == "contradiction-no-failure" || profile == "contradiction-single") && expectedBytesDigest == actualDigest {
		return ObservationBundle{}, fmt.Errorf("contradiction profile requires mismatching target bytes")
	}
	observations := []ObservationReceipt{}
	switch profile {
	case "accepted":
		for _, claim := range parsed.Graph.Nodes {
			observations = append(observations, makeObservation("CLAIM", claim.ClaimID, claim.PropositionDigest, "", "", "", "", claim.Target, artifactPath, actualDigest, actualDigest, ObservationEvidence, ObservationEvidence, "CLAIM_TARGET_BYTES_MATCH", output))
		}
	case "contradiction", "contradiction-no-failure", "contradiction-single":
		targeted := map[string]bool{}
		singleEdgeID := ""
		if profile == "contradiction-single" {
			for _, edge := range parsed.Graph.Edges {
				if edge.Kind == Contradicts {
					singleEdgeID = edge.EdgeID
					break
				}
			}
			if singleEdgeID == "" {
				return ObservationBundle{}, fmt.Errorf("single contradiction profile requires a CONTRADICTS edge")
			}
		}
		for _, edge := range parsed.Graph.Edges {
			if edge.Kind != Contradicts && edge.Kind != FailureEntailment {
				continue
			}
			if profile == "contradiction-single" && edge.EdgeID != singleEdgeID {
				continue
			}
			targeted[edge.ToClaimID] = true
		}
		for _, claim := range parsed.Graph.Nodes {
			if !targeted[claim.ClaimID] {
				observations = append(observations, makeObservation("CLAIM", claim.ClaimID, claim.PropositionDigest, "", "", "", "", claim.Target, artifactPath, actualDigest, actualDigest, ObservationEvidence, ObservationEvidence, "CLAIM_TARGET_BYTES_MATCH", output))
			}
		}
		for _, edge := range parsed.Graph.Edges {
			if edge.Kind != Contradicts && edge.Kind != FailureEntailment {
				continue
			}
			if profile == "contradiction-single" && edge.EdgeID != singleEdgeID {
				continue
			}
			if profile == "contradiction-no-failure" && edge.Kind == FailureEntailment {
				continue
			}
			to := parsed.Graph.Nodes[indexOfClaim(edge.ToClaimID, parsed.Graph)]
			predicate := ObservationContradiction
			if edge.Kind == FailureEntailment {
				predicate = ObservationFailure
			}
			observations = append(observations, makeObservation("EDGE", to.ClaimID, to.PropositionDigest, edge.EdgeID, edge.FromClaimID, edge.ToClaimID, edge.Kind, to.Target, artifactPath, expectedBytesDigest, actualDigest, predicate, predicate, "EDGE_ACTIVATION_BYTES_MISMATCH", output))
		}
	case "unrelated":
		observations = append(observations, makeObservation("UNRELATED_ARTIFACT", "", "", "", "", "", "", TargetAddress{Artifact: artifactPath}, artifactPath, expectedBytesDigest, actualDigest, ObservationUnknown, ObservationUnknown, "UNRELATED_ARTIFACT_BYTES_COMPARISON", output))
	default:
		return ObservationBundle{}, fmt.Errorf("unknown observation profile %q", profile)
	}
	for i := range observations {
		observations[i].Digest, err = observationReceiptDigest(observations[i])
		if err != nil {
			return ObservationBundle{}, err
		}
	}
	bundle := ObservationBundle{Schema: observationBundleSchema, Provider: "github-actions-target-observer/v2", SourcePath: sourcePath, SourceDigest: digestBytes(source), ArtifactPath: artifactPath, ArtifactBytesDigest: actualDigest, Profile: profile, Observations: observations}
	bundle.Digest, err = observationBundleDigest(bundle)
	if err != nil {
		return ObservationBundle{}, err
	}
	return bundle, nil
}

func makeObservation(binding, claimID, propositionDigest, edgeID, fromClaimID, toClaimID string, edgeKind EdgeKind, target TargetAddress, artifactPath, expectedValue, observedValue string, expectedPredicate, observedPredicate ObservationPredicate, reason, output string) ObservationReceipt {
	comparison := "MISMATCH"
	if expectedValue == observedValue {
		comparison = "MATCH"
	}
	return ObservationReceipt{Schema: observationSchema, Provider: "github-actions-target-observer/v2", Binding: binding, ClaimID: claimID, PropositionDigest: propositionDigest, EdgeID: edgeID, FromClaimID: fromClaimID, ToClaimID: toClaimID, EdgeKind: edgeKind, Target: target, TargetPath: artifactPath, TargetBytesDigest: observedValue, ExpectedPredicate: expectedPredicate, ExpectedValue: expectedValue, ObservedPredicate: observedPredicate, ObservedValue: observedValue, ComparisonResult: comparison, Procedure: observationProcedure, ProcedureDigest: digestBytes([]byte(observationProcedure)), Output: output, OutputDigest: digestBytes([]byte(output)), Coordinate: Coordinate{Stage: "OBSERVE", Step: "target-observer", Reason: reason}}
}

func validateObservationBundle(bundle ObservationBundle, artifactPath string, artifact []byte, graph Graph) error {
	if bundle.Schema != observationBundleSchema || bundle.Provider == "" || bundle.SourcePath != artifactPath || bundle.SourceDigest != digestBytes(artifact) || bundle.ArtifactPath != artifactPath || bundle.ArtifactBytesDigest != digestBytes(artifact) || bundle.Profile == "" || bundle.Digest == "" || len(bundle.Observations) == 0 {
		return fmt.Errorf("target observation bundle identity or target binding is invalid")
	}
	if digest, err := observationBundleDigest(bundle); err != nil || digest != bundle.Digest {
		return fmt.Errorf("target observation bundle digest is invalid")
	}
	seen := map[string]bool{}
	for _, observation := range bundle.Observations {
		if err := validateObservation(observation, artifactPath, artifact, graph); err != nil {
			return err
		}
		key := observation.Binding + "|" + observation.ClaimID + "|" + observation.EdgeID
		if seen[key] {
			return fmt.Errorf("target observation bundle has duplicate binding %q", key)
		}
		seen[key] = true
	}
	return nil
}

func validateObservation(value ObservationReceipt, artifactPath string, artifact []byte, graph Graph) error {
	if value.Schema != observationSchema || value.Provider == "" || value.TargetPath != artifactPath || value.TargetBytesDigest != digestBytes(artifact) || value.Procedure == "" || value.ProcedureDigest != digestBytes([]byte(value.Procedure)) || value.OutputDigest != digestBytes([]byte(value.Output)) || value.Coordinate.Stage == "" || value.Digest == "" {
		return fmt.Errorf("target observation identity or target binding is invalid")
	}
	if value.ComparisonResult != "MATCH" && value.ComparisonResult != "MISMATCH" || value.ObservedValue != value.TargetBytesDigest {
		return fmt.Errorf("target observation comparison is invalid")
	}
	if value.ComparisonResult == "MATCH" && value.ExpectedValue != value.ObservedValue || value.ComparisonResult == "MISMATCH" && value.ExpectedValue == value.ObservedValue {
		return fmt.Errorf("target observation comparison result does not match values")
	}
	switch value.Binding {
	case "CLAIM":
		claimIndex := indexOfClaim(value.ClaimID, graph)
		validEvidence := value.ExpectedPredicate == ObservationEvidence && value.ObservedPredicate == ObservationEvidence && value.ComparisonResult == "MATCH"
		validContradiction := value.ExpectedPredicate == ObservationContradiction && value.ObservedPredicate == ObservationContradiction && value.ComparisonResult == "MISMATCH"
		if claimIndex < 0 || value.PropositionDigest != graph.Nodes[claimIndex].PropositionDigest || !reflect.DeepEqual(value.Target, graph.Nodes[claimIndex].Target) || (!validEvidence && !validContradiction) {
			return fmt.Errorf("claim-scoped target observation is not bound to its claim")
		}
	case "EDGE":
		edgeIndex := indexOfEdge(value.EdgeID, graph)
		if edgeIndex < 0 {
			return fmt.Errorf("edge-scoped target observation references an unknown edge")
		}
		edge := graph.Edges[edgeIndex]
		to := indexOfClaim(edge.ToClaimID, graph)
		if value.FromClaimID != edge.FromClaimID || value.ToClaimID != edge.ToClaimID || value.EdgeKind != edge.Kind || value.ClaimID != edge.ToClaimID || value.PropositionDigest != graph.Nodes[to].PropositionDigest || !reflect.DeepEqual(value.Target, graph.Nodes[to].Target) || value.ExpectedPredicate != value.ObservedPredicate || value.ComparisonResult != "MISMATCH" || (edge.Kind == Contradicts && value.ObservedPredicate != ObservationContradiction) || (edge.Kind == FailureEntailment && value.ObservedPredicate != ObservationFailure) {
			return fmt.Errorf("edge-scoped target observation is not bound to its edge")
		}
	case "UNRELATED_ARTIFACT":
		if value.ClaimID != "" || value.EdgeID != "" || value.PropositionDigest != "" || value.ObservedPredicate != ObservationUnknown || value.ExpectedPredicate != ObservationUnknown {
			return fmt.Errorf("unrelated observation carries a claim or edge binding")
		}
	default:
		return fmt.Errorf("unknown target observation binding %q", value.Binding)
	}
	if digest, err := observationReceiptDigest(value); err != nil || digest != value.Digest {
		return fmt.Errorf("target observation digest is invalid")
	}
	return nil
}

func claimObservationFor(claim Claim, observations []ObservationReceipt) (ObservationReceipt, bool) {
	for _, observation := range observations {
		if observation.Binding == "CLAIM" && observation.ClaimID == claim.ClaimID && observation.PropositionDigest == claim.PropositionDigest && reflect.DeepEqual(observation.Target, claim.Target) {
			return observation, true
		}
	}
	return ObservationReceipt{}, false
}

func hasClaimOrEdgeObservation(observations []ObservationReceipt) bool {
	for _, observation := range observations {
		if observation.Binding == "CLAIM" || observation.Binding == "EDGE" {
			return true
		}
	}
	return false
}

func summaryPredicate(observations []ObservationReceipt) ObservationPredicate {
	for _, observation := range observations {
		if observation.Binding == "EDGE" && observation.ObservedPredicate == ObservationContradiction {
			return ObservationContradiction
		}
	}
	for _, observation := range observations {
		if observation.Binding == "EDGE" && observation.ObservedPredicate == ObservationFailure {
			return ObservationFailure
		}
	}
	for _, observation := range observations {
		if observation.Binding == "CLAIM" && observation.ObservedPredicate == ObservationContradiction {
			return ObservationContradiction
		}
	}
	for _, observation := range observations {
		if observation.Binding == "CLAIM" && observation.ObservedPredicate == ObservationEvidence {
			return ObservationEvidence
		}
	}
	return ObservationUnknown
}

func observationBundleValue(bundle ObservationBundle) string {
	return fmt.Sprintf("bundle:%s|profile:%s|source_digest:%s|artifact_path:%s|artifact_bytes_digest:%s|observation_total:%d", bundle.Digest, bundle.Profile, bundle.SourceDigest, bundle.ArtifactPath, bundle.ArtifactBytesDigest, len(bundle.Observations))
}

func absentClaimValue(claim Claim) string {
	return fmt.Sprintf("observation:ABSENT|claim_id:%s|proposition_digest:%s|target_artifact:%s|stage:OBSERVE|step:claim-observation|reason:CLAIM_OBSERVATION_MISSING", claim.ClaimID, claim.PropositionDigest, claim.Target.Artifact)
}

func observationReason(status EvidenceStatus, predicate ObservationPredicate) string {
	if status == UnknownEvidence {
		return "EXTERNAL_TARGET_OBSERVATION_MISSING"
	}
	return "CURRENT_TARGET_OBSERVATION_" + string(predicate)
}

func readCapability(path string) (CapabilityEvidence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CapabilityEvidence{}, fmt.Errorf("capability evidence: %w", err)
	}
	var capability CapabilityEvidence
	if err := json.Unmarshal(data, &capability); err != nil {
		return CapabilityEvidence{}, fmt.Errorf("capability evidence decode: %w", err)
	}
	if capability.Provider == "" || capability.Permission == "" || capability.Coordinate.Stage == "" {
		return CapabilityEvidence{}, fmt.Errorf("capability evidence is incomplete")
	}
	return capability, nil
}

func repositorySnapshot(root, outputPath string) (RepositorySnapshot, error) {
	tracked, err := gitRead(root, "ls-files", "-s")
	if err != nil {
		return RepositorySnapshot{}, err
	}
	untracked, err := gitRead(root, "status", "--short", "--untracked-files=all")
	if err != nil {
		return RepositorySnapshot{}, err
	}
	state := tracked + "\x00" + untracked
	outputPath, err = filepath.Abs(outputPath)
	if err != nil {
		return RepositorySnapshot{}, err
	}
	snapshot := RepositorySnapshot{RepositoryRoot: root, TrackedDigest: digestBytes([]byte(tracked)), UntrackedDigest: digestBytes([]byte(untracked)), BeforeDigest: digestBytes([]byte(state)), OutputPath: outputPath, OutputDigest: digestBytes([]byte(outputPath)), Coordinate: Coordinate{Stage: "OBSERVE", Step: "repository-snapshot", Reason: "TRACKED_UNTRACKED_PRE_POST"}}
	trackedAfter, err := gitRead(root, "ls-files", "-s")
	if err != nil {
		return RepositorySnapshot{}, err
	}
	untrackedAfter, err := gitRead(root, "status", "--short", "--untracked-files=all")
	if err != nil {
		return RepositorySnapshot{}, err
	}
	after := trackedAfter + "\x00" + untrackedAfter
	snapshot.AfterDigest = digestBytes([]byte(after))
	if snapshot.BeforeDigest != snapshot.AfterDigest {
		snapshot.RepositoryWrites = 1
	}
	return snapshot, nil
}

func gitRead(root string, args ...string) (string, error) {
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("repository observation %v: %w", args, err)
	}
	return string(output), nil
}

// Evaluate applies the typed state algebra to a source-derived graph and a
// provider receipt. prior, when present, must be the exact UNKNOWN receipt;
// recovery appends to its transition chain and never creates a new ledger.
func Evaluate(source []byte, sourcePath string, evidence EvidenceReceipt, prior *Receipt) (Receipt, error) {
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return Receipt{}, err
	}
	truthTable := TruthTableCases()
	if err := validateTruthTable(truthTable); err != nil {
		return Receipt{}, err
	}
	if err := validateEvidence(parsed, evidence); err != nil {
		return Receipt{}, err
	}
	runtimeSnapshot, err := repositorySnapshot(evidence.Snapshot.RepositoryRoot, evidence.Snapshot.OutputPath)
	if err != nil {
		return Receipt{}, fmt.Errorf("producer execution snapshot: %w", err)
	}
	if runtimeSnapshot.BeforeDigest != evidence.Snapshot.BeforeDigest || runtimeSnapshot.AfterDigest != evidence.Snapshot.AfterDigest || runtimeSnapshot.RepositoryWrites != evidence.Snapshot.RepositoryWrites {
		return Receipt{}, fmt.Errorf("producer execution crossed observed repository snapshot boundary")
	}
	if prior != nil {
		if err := validatePrior(parsed, *prior); err != nil {
			return Receipt{}, err
		}
	}
	sourceDigest, semanticDigest := digestBytes(source), parsed.Graph.CanonicalIRDigest
	provenance := fmt.Sprintf("source:%s|ir:%s|evidence:%s|producer:%s|consumer:%s", sourceDigest, semanticDigest, evidence.Digest, ProducerID, ConsumerID)
	states, outcomes, local := classify(parsed.Graph, evidence)
	transitions, err := buildTransitions(parsed.Graph, outcomes, local, provenance, prior)
	if err != nil {
		return Receipt{}, err
	}
	currentOutcomes := outcomes
	if prior != nil {
		currentOutcomes = transitions[len(transitions)-ClaimTotal:]
	}
	resolutions := buildResolutions(parsed.Graph, states, currentOutcomes, provenance)
	metrics := deriveMetrics(parsed.Graph, states, resolutions, currentOutcomes, evidence, prior != nil)
	decision := decisionFor(states, evidence, prior != nil)
	authorityCases := AuthorityCases()
	if err := validateAuthorityCases(authorityCases); err != nil {
		return Receipt{}, err
	}
	subjectAuthority := authorityResolution(evidence)
	subject := Subject{SourcePath: sourcePath, SourceDigest: sourceDigest, SemanticDigest: semanticDigest, Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID, ProofChoice: ProofChoice, ReadOnly: subjectAuthority == "NET_REPOSITORY_STATE_UNCHANGED", RepositoryWrites: evidence.Snapshot.RepositoryWrites, AuthorityResolution: subjectAuthority, AuthorityCoordinate: evidence.Capability.Coordinate}
	receipt := Receipt{Schema: ReceiptSchema, Scope: Scope, Subject: subject, Evidence: evidence, Graph: parsed.Graph, TruthTable: truthTable, AuthorityCases: authorityCases, EvidenceDigest: evidence.Digest, Transitions: transitions, TransitionHeadDigest: transitions[len(transitions)-1].TransitionDigest, Resolutions: resolutions, Metrics: metrics, Decision: decision}
	if prior != nil {
		receipt.PriorReceiptDigest, err = receiptDigest(*prior)
		if err != nil {
			return Receipt{}, err
		}
		receipt.PreviousTransitionDigest = prior.TransitionHeadDigest
		receipt.PriorClaimStates = resolutionStates(prior.Resolutions)
		receipt.Metrics.AppendOnlyTransitionTotal = ClaimTotal
	}
	receipt.Digest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func validateEvidence(parsed sourceGraph, evidence EvidenceReceipt) error {
	if evidence.Schema != EvidenceSchema || (evidence.Status != CurrentEvidence && evidence.Status != UnknownEvidence) || evidence.Provider == "" || evidence.ArtifactPath == "" || evidence.ArtifactBytesDigest == "" || evidence.Digest == "" || evidence.RequestStatus != "CLAIMED_INPUT" || evidence.Procedure != evidenceProcedure {
		return fmt.Errorf("evidence receipt identity is invalid")
	}
	computed, err := evidenceReceiptDigest(evidence)
	if err != nil || computed != evidence.Digest {
		return fmt.Errorf("current evidence receipt digest is invalid")
	}
	if evidence.Snapshot.BeforeDigest == "" || evidence.Snapshot.AfterDigest == "" || evidence.Snapshot.OutputPath == "" {
		return fmt.Errorf("repository snapshot is incomplete")
	}
	if evidence.Snapshot.RepositoryWrites != 0 || evidence.Snapshot.BeforeDigest != evidence.Snapshot.AfterDigest {
		return fmt.Errorf("evidence crossed repository write boundary")
	}
	capDigest, err := capabilityDigest(evidence.Capability)
	if err != nil || capDigest != evidence.Capability.Digest || evidence.Capability.Status != CurrentEvidence {
		return fmt.Errorf("capability evidence is invalid")
	}
	artifact, err := os.ReadFile(evidence.ArtifactPath)
	if err != nil {
		return fmt.Errorf("producer cannot re-observe artifact: %w", err)
	}
	if digestBytes(artifact) != evidence.ArtifactBytesDigest {
		return fmt.Errorf("artifact bytes digest changed")
	}
	artifactGraph, err := graphFromSource(artifact, evidence.ArtifactPath)
	if err != nil {
		return fmt.Errorf("producer artifact re-observation: %w", err)
	}

	observations := []ObservationReceipt(nil)
	if evidence.ObservationPath != "" {
		observationBytes, err := os.ReadFile(evidence.ObservationPath)
		if err != nil {
			return fmt.Errorf("producer cannot re-observe target observation: %w", err)
		}
		var bundle ObservationBundle
		if err := json.Unmarshal(observationBytes, &bundle); err != nil {
			return fmt.Errorf("target observation bundle decode: %w", err)
		}
		if err := validateObservationBundle(bundle, evidence.ArtifactPath, artifact, artifactGraph.Graph); err != nil {
			return err
		}
		if bundle.Digest != evidence.ObservationBundleDigest || !reflect.DeepEqual(bundle.Observations, evidence.Observations) {
			return fmt.Errorf("embedded target observation bundle differs from raw bundle")
		}
		observations = bundle.Observations
	} else if len(evidence.Observations) != 0 || evidence.ObservationBundleDigest != "" {
		return fmt.Errorf("evidence has observations without a raw observation bundle")
	}

	if len(evidence.Claims) != artifactGraph.Graph.NodeTotal {
		return fmt.Errorf("evidence claim count does not match artifact graph")
	}
	expectedStatus := UnknownEvidence
	if hasClaimOrEdgeObservation(observations) {
		expectedStatus = CurrentEvidence
	}
	if evidence.Status != expectedStatus {
		return fmt.Errorf("evidence status is not derived from claim-scoped observations")
	}
	expectedPredicate := summaryPredicate(observations)
	expectedValue := fmt.Sprintf("observation:ABSENT|stage:OBSERVE|step:current-evidence-provider|reason:EXTERNAL_TARGET_OBSERVATION_MISSING|artifact_path_digest:%s|artifact_bytes_digest:%s", digestBytes([]byte(evidence.ArtifactPath)), digestBytes(artifact))
	if evidence.ObservationPath != "" {
		var bundle ObservationBundle
		observationBytes, _ := os.ReadFile(evidence.ObservationPath)
		if err := json.Unmarshal(observationBytes, &bundle); err != nil {
			return fmt.Errorf("target observation bundle decode: %w", err)
		}
		expectedValue = observationBundleValue(bundle)
	}
	if evidence.ObservedPredicate != expectedPredicate || evidence.ObservedValue != expectedValue {
		return fmt.Errorf("evidence predicate is not computed by claim-scoped observation procedure")
	}
	for i, claim := range artifactGraph.Graph.Nodes {
		ec := evidence.Claims[i]
		if ec.ClaimID != claim.ClaimID || ec.PropositionDigest != claim.PropositionDigest || ec.Digest == "" {
			return fmt.Errorf("evidence claim %d is not source-derived", i+1)
		}
		claimObservation, found := claimObservationFor(claim, observations)
		expectedClaim := EvidenceClaim{ClaimID: claim.ClaimID, PropositionDigest: claim.PropositionDigest, ObservedPredicate: ObservationUnknown, ObservedValue: absentClaimValue(claim), Status: UnknownEvidence, Coordinate: Coordinate{Stage: "OBSERVE", Step: claim.ActivityName, Reason: "CLAIM_OBSERVATION_MISSING"}}
		if found {
			expectedClaim.ObservedPredicate = claimObservation.ObservedPredicate
			expectedClaim.ObservedValue = claimObservation.ObservedValue
			expectedClaim.Status = CurrentEvidence
			expectedClaim.Coordinate = claimObservation.Coordinate
		}
		expectedClaim.Digest, err = claimEvidenceDigest(expectedClaim)
		if err != nil || !reflect.DeepEqual(ec, expectedClaim) {
			return fmt.Errorf("evidence claim %d is not derived from its exact observation binding", i+1)
		}
	}
	_ = parsed
	return nil
}

func validatePrior(parsed sourceGraph, prior Receipt) error {
	if prior.Schema != ReceiptSchema || prior.Scope != Scope || prior.Evidence.ObservedPredicate != ObservationUnknown || prior.Graph.Digest != parsed.Graph.Digest || len(prior.Resolutions) != ClaimTotal || len(prior.PriorClaimStates) != 0 {
		return fmt.Errorf("prior ledger is not an UNKNOWN receipt for this graph")
	}
	if d, _ := receiptDigest(prior); d != prior.Digest {
		return fmt.Errorf("prior receipt digest is invalid")
	}
	priorSource, err := os.ReadFile(prior.Subject.SourcePath)
	if err != nil {
		return fmt.Errorf("prior source cannot be re-observed: %w", err)
	}
	replayed, err := Evaluate(priorSource, prior.Subject.SourcePath, prior.Evidence, nil)
	if err != nil {
		return fmt.Errorf("prior receipt replay failed: %w", err)
	}
	if !reflect.DeepEqual(replayed, prior) {
		return fmt.Errorf("prior receipt is not a complete replay of raw source and evidence")
	}
	if err := validateTransitionChain(prior.Transitions, prior.TransitionHeadDigest); err != nil {
		return err
	}
	for i, r := range prior.Resolutions {
		if r.State != "OPEN" || r.ClaimID != parsed.Graph.Nodes[i].ClaimID {
			return fmt.Errorf("prior claim %d was not OPEN", i+1)
		}
	}
	return nil
}

type localObservation struct {
	Predicate ObservationPredicate
	Digest    string
	Available bool
}

func classify(graph Graph, evidence EvidenceReceipt) ([]string, []Transition, []localObservation) {
	states := make([]string, len(graph.Nodes))
	outcomes := make([]Transition, len(graph.Nodes))
	local := make([]localObservation, len(graph.Nodes))
	for i, claim := range graph.Nodes {
		for _, observed := range evidence.Claims {
			if observed.ClaimID == claim.ClaimID && observed.PropositionDigest == claim.PropositionDigest && observed.Status == CurrentEvidence {
				local[i] = localObservation{Predicate: observed.ObservedPredicate, Digest: observed.Digest, Available: true}
				break
			}
		}
		state, event, reason := "OPEN", "DEPENDENCY_BLOCKED", "UPSTREAM_UNKNOWN_OR_NON_REFUTING"
		incoming := incomingEdges(i, graph)
		refuting := []string{}
		hasRequires, allRequires := false, true
		for _, edge := range incoming {
			from := indexOfClaim(edge.FromClaimID, graph)
			if from < 0 {
				continue
			}
			relation := edgeRelation(edge.Kind, states[from], local[i].Predicate, true, edgeObservationActive(edge, evidence.Observations, ObservationContradiction), edgeObservationActive(edge, evidence.Observations, ObservationFailure))
			if relation == relationRefuted {
				refuting = append(refuting, edge.EdgeID)
			}
			if edge.Kind == Requires {
				hasRequires = true
				if relation != relationDischarged {
					allRequires = false
				}
			}
		}
		// Explicit root contradiction is the observation itself; downstream
		// refutation requires a typed, directional upstream edge.
		if local[i].Available && local[i].Predicate == ObservationContradiction {
			state, event, reason = "REFUTED", "EXPLICIT_CONTRADICTION", "CURRENT_EVIDENCE_EXPLICIT_CONTRADICTION"
		} else if len(refuting) > 0 {
			state, event, reason = "REFUTED", "DEPENDENCY_REFUTED", "EXPLICIT_TYPED_REFUTING_EDGE"
		} else if local[i].Available && local[i].Predicate == ObservationEvidence {
			if !hasRequires || allRequires {
				state, event, reason = "DISCHARGED", "EVIDENCE_ACCEPTED", "LOCAL_CLAIM_EVIDENCE_PREDICATE"
				if hasRequires {
					event, reason = "DEPENDENCY_DISCHARGED", "ALL_REQUIRES_UPSTREAM_AND_LOCAL_EVIDENCE"
				}
			}
		}
		states[i] = state
		outcomes[i] = Transition{ClaimID: claim.ClaimID, Event: event, Before: "OPEN", After: state, Coordinate: Coordinate{Stage: outcomeStage(i), Step: claim.ActivityName, Reason: reason}, EvidenceDigest: local[i].Digest, UpstreamEdgeIDs: transitionEdges(i, graph, states, state, refuting, local[i].Predicate, evidence.Observations), Provenance: "pending"}
	}
	return states, outcomes, local
}

// edgeObservationActive is intentionally claim/edge scoped.  A mismatch on
// another artifact, or an observation attached to another proposition, cannot
// activate this edge even when its predicate has the same string value.
func edgeObservationActive(edge Edge, observations []ObservationReceipt, predicate ObservationPredicate) bool {
	for _, observation := range observations {
		if observation.Binding == "EDGE" && observation.EdgeID == edge.EdgeID && observation.FromClaimID == edge.FromClaimID && observation.ToClaimID == edge.ToClaimID && observation.EdgeKind == edge.Kind && observation.ClaimID == edge.ToClaimID && observation.ObservedPredicate == predicate {
			return true
		}
	}
	return false
}

func outcomeStage(index int) string {
	if index == 0 {
		return "OBSERVE"
	}
	return "PROPAGATE"
}
func incomingEdges(index int, graph Graph) []Edge {
	result := []Edge{}
	for _, edge := range graph.Edges {
		if edge.ToClaimID == graph.Nodes[index].ClaimID {
			result = append(result, edge)
		}
	}
	return result
}
func transitionEdges(index int, graph Graph, states []string, state string, refuting []string, local ObservationPredicate, observations []ObservationReceipt) []string {
	if len(refuting) > 0 {
		return refuting
	}
	if state == "DISCHARGED" {
		var result []string
		for _, edge := range incomingEdges(index, graph) {
			from := indexOfClaim(edge.FromClaimID, graph)
			if edge.Kind == Requires && from >= 0 && edgeRelation(edge.Kind, states[from], local, true, edgeObservationActive(edge, observations, ObservationContradiction), edgeObservationActive(edge, observations, ObservationFailure)) == relationDischarged {
				result = append(result, edge.EdgeID)
			}
		}
		return result
	}
	if state != "OPEN" {
		return nil
	}
	var result []string
	for _, edge := range incomingEdges(index, graph) {
		from := indexOfClaim(edge.FromClaimID, graph)
		if from >= 0 && (edge.Kind == Supports || edge.Kind == Requires) && edgeRelation(edge.Kind, states[from], local, true, edgeObservationActive(edge, observations, ObservationContradiction), edgeObservationActive(edge, observations, ObservationFailure)) == relationOpen && (states[from] == "OPEN" || states[from] == "REFUTED") {
			result = append(result, edge.EdgeID)
		}
	}
	return result
}

func buildTransitions(graph Graph, outcomes []Transition, local []localObservation, provenance string, prior *Receipt) ([]Transition, error) {
	result := []Transition{}
	previous := ""
	if prior != nil {
		result = append(result, prior.Transitions...)
		previous = prior.TransitionHeadDigest
	}
	if prior == nil {
		for _, claim := range graph.Nodes {
			value := Transition{Sequence: len(result) + 1, ClaimID: claim.ClaimID, Event: "CLAIM_REGISTERED", Before: "UNRECORDED", After: "OPEN", Coordinate: Coordinate{Stage: "DECLARE", Step: claim.ActivityName, Reason: "CLAIM_REGISTERED"}, Provenance: provenance, PreviousTransitionDigest: previous}
			value.TransitionDigest, _ = transitionDigest(value)
			result = append(result, value)
			previous = value.TransitionDigest
		}
	}
	for i, outcome := range outcomes {
		outcome.Sequence = len(result) + 1
		outcome.Provenance = provenance
		outcome.PreviousTransitionDigest = previous
		outcome.UpstreamTransitionDigests = upstreamTransitionDigests(outcome.UpstreamEdgeIDs, graph, result, prior)
		if prior != nil {
			outcome.Before = "OPEN"
		}
		outcome.TransitionDigest, _ = transitionDigest(outcome)
		result = append(result, outcome)
		previous = outcome.TransitionDigest
		_ = local[i]
	}
	return result, nil
}

func upstreamTransitionDigests(edgeIDs []string, graph Graph, transitions []Transition, prior *Receipt) []string {
	var result []string
	for _, edgeID := range edgeIDs {
		for _, edge := range graph.Edges {
			if edge.EdgeID != edgeID {
				continue
			}
			for i := len(transitions) - 1; i >= 0; i-- {
				if transitions[i].ClaimID == edge.FromClaimID && transitions[i].Event != "CLAIM_REGISTERED" {
					result = append(result, transitions[i].TransitionDigest)
					break
				}
			}
			if prior != nil {
				_ = prior
			}
		}
	}
	return result
}

func buildResolutions(graph Graph, states []string, outcomes []Transition, provenance string) []Resolution {
	result := make([]Resolution, len(graph.Nodes))
	for i, claim := range graph.Nodes {
		path, edgeIDs, kinds := shortestPath(i, graph, states[i])
		transitionDigests := []string{}
		for _, node := range path {
			for _, outcome := range outcomes {
				if outcome.ClaimID == graph.Nodes[node].ClaimID {
					transitionDigests = append(transitionDigests, outcome.TransitionDigest)
					break
				}
			}
		}
		causePath := claimIDs(path, graph)
		responsibility, owner := failureAttribution(i, states[i], causePath)
		resolution := Resolution{ClaimID: claim.ClaimID, Axis: claim.Axis, PropositionDigest: claim.PropositionDigest, State: states[i], Kind: resolutionKind(i, states[i], outcomes[i]), ObservedEvent: outcomes[i].Event, Coordinate: outcomes[i].Coordinate, EvidenceDigest: outcomes[i].EvidenceDigest, Provenance: provenance, FailureResponsibility: responsibility, FailureOwnerClaimID: owner, CausePath: causePath, CauseEdgeIDs: edgeIDs, CauseEdgeKinds: kinds, CauseTransitionDigests: transitionDigests, CauseCoordinate: coordinatePointer(outcomes[i].Coordinate)}
		if states[i] == "OPEN" {
			resolution.MissingEvidenceIDs = []string{"evidence:" + claim.ClaimID}
			resolution.BlockedByClaimIDs, resolution.BlockedByEdgeIDs = blockedFrontier(i, graph, states)
		}
		result[i] = resolution
	}
	return result
}

func failureAttribution(index int, state string, path []string) (string, string) {
	if state == "DISCHARGED" {
		return "N/A", ""
	}
	if index == 0 || len(path) <= 1 {
		if index == 0 {
			return "DIRECT_CLAIM", pathOrEmpty(path, index)
		}
		return "DIRECT_CLAIM", pathOrEmpty(path, index)
	}
	return "UPSTREAM_CLAIM", path[0]
}

func pathOrEmpty(path []string, index int) string {
	if len(path) > 0 {
		return path[0]
	}
	return ""
}
func coordinatePointer(value Coordinate) *Coordinate { return &value }
func resolutionKind(index int, state string, outcome Transition) string {
	if index == 0 {
		switch state {
		case "REFUTED":
			return "DIRECT_REFUTED"
		case "DISCHARGED":
			return "DIRECT_DISCHARGED"
		default:
			return "DIRECT_UNKNOWN"
		}
	}
	switch state {
	case "REFUTED":
		return "DEPENDENCY_REFUTED"
	case "DISCHARGED":
		if len(outcome.UpstreamEdgeIDs) == 0 {
			return "DIRECT_DISCHARGED"
		}
		return "DEPENDENCY_DISCHARGED"
	default:
		return "DEPENDENCY_BLOCKED"
	}
}

func shortestPath(index int, graph Graph, state string) ([]int, []string, []EdgeKind) {
	if index == 0 {
		return []int{0}, nil, nil
	}
	allowed := map[EdgeKind]bool{Supports: state == "OPEN", Requires: state == "OPEN" || state == "DISCHARGED", Contradicts: state == "REFUTED", FailureEntailment: state == "REFUTED"}
	best := []int(nil)
	bestEdges := []Edge(nil)
	var walk func(int, []int, []Edge)
	walk = func(current int, path []int, edges []Edge) {
		if current == index {
			if best == nil || len(path) < len(best) || (len(path) == len(best) && pathKey(path, graph) < pathKey(best, graph)) {
				best = append([]int(nil), path...)
				bestEdges = append([]Edge(nil), edges...)
			}
			return
		}
		for _, edge := range graph.Edges {
			if edge.ToClaimID != graph.Nodes[index].ClaimID && edge.FromClaimID != graph.Nodes[current].ClaimID {
				continue
			}
			if edge.FromClaimID != graph.Nodes[current].ClaimID || !allowed[edge.Kind] {
				continue
			}
			next := indexOfClaim(edge.ToClaimID, graph)
			seen := false
			for _, n := range path {
				if n == next {
					seen = true
				}
			}
			if !seen {
				walk(next, append(path, next), append(edges, edge))
			}
		}
	}
	walk(0, []int{0}, nil)
	if best == nil {
		return []int{index}, nil, nil
	}
	ids, kinds := []string{}, []EdgeKind{}
	for _, edge := range bestEdges {
		ids = append(ids, edge.EdgeID)
		kinds = append(kinds, edge.Kind)
	}
	return best, ids, kinds
}
func pathKey(path []int, graph Graph) string {
	values := make([]string, len(path))
	for i, value := range path {
		values[i] = graph.Nodes[value].ClaimID
	}
	return strings.Join(values, "\x00")
}
func claimIDs(path []int, graph Graph) []string {
	result := make([]string, len(path))
	for i, index := range path {
		result[i] = graph.Nodes[index].ClaimID
	}
	return result
}
func indexOfClaim(id string, graph Graph) int {
	for i, claim := range graph.Nodes {
		if claim.ClaimID == id {
			return i
		}
	}
	return -1
}
func indexOfEdge(id string, graph Graph) int {
	for i, edge := range graph.Edges {
		if edge.EdgeID == id {
			return i
		}
	}
	return -1
}
func blockedFrontier(index int, graph Graph, states []string) ([]string, []string) {
	var claims, edges []string
	for _, edge := range incomingEdges(index, graph) {
		from := indexOfClaim(edge.FromClaimID, graph)
		if from >= 0 && (edge.Kind == Supports || edge.Kind == Requires) && (states[from] == "OPEN" || states[from] == "REFUTED") {
			claims, edges = append(claims, edge.FromClaimID), append(edges, edge.EdgeID)
		}
	}
	return claims, edges
}

func deriveMetrics(graph Graph, states []string, resolutions []Resolution, outcomes []Transition, evidence EvidenceReceipt, recovered bool) Metrics {
	metrics := Metrics{FixedClaimTotal: ClaimTotal, DistinctPropositionTotal: distinctPropositions(graph), FixedEdgeTotal: EdgeTotal, EligibleEdgeTotal: len(graph.Edges), ClassifiedClaimTotal: len(states), ClassificationBasisPoints: 10000, TransitionTotal: InitialTransitionTotal, TruthTableCaseTotal: len(TruthTableCases()), AuthorityCaseTotal: len(AuthorityCases())}
	if recovered {
		metrics.TransitionTotal += ClaimTotal
	}
	for _, ec := range evidence.Claims {
		if ec.Status == CurrentEvidence {
			metrics.CurrentEvidenceTotal++
		}
		if ec.Status == HistoricalFixture {
			metrics.HistoricalEvidenceTotal++
		}
		if ec.ObservedPredicate == ObservationUnknown {
			metrics.UnknownEvidenceTotal++
		}
	}
	observed, shortestUnion := map[string]bool{}, map[string]bool{}
	for _, state := range states {
		switch state {
		case "OPEN":
			metrics.OpenClaimTotal++
		case "DISCHARGED":
			metrics.DischargedClaimTotal++
		case "REFUTED":
			metrics.RefutedClaimTotal++
		}
	}
	for i, resolution := range resolutions {
		switch resolution.Kind {
		case "DIRECT_UNKNOWN":
			metrics.DirectUnknownClaimTotal++
		case "DEPENDENCY_BLOCKED":
			metrics.DependencyBlockedClaimTotal++
		case "DIRECT_REFUTED":
			metrics.DirectRefutedClaimTotal++
		case "DEPENDENCY_REFUTED":
			metrics.DependencyRefutedClaimTotal++
		case "DIRECT_DISCHARGED":
			metrics.DirectDischargedClaimTotal++
		case "DEPENDENCY_DISCHARGED":
			metrics.DependencyDischargedTotal++
		}
		for _, edge := range outcomes[i].UpstreamEdgeIDs {
			observed[edge] = true
		}
		for _, edge := range resolution.CauseEdgeIDs {
			shortestUnion[edge] = true
		}
		if len(resolution.CauseEdgeIDs) > metrics.MaximumCausePathDepth {
			metrics.MaximumCausePathDepth = len(resolution.CauseEdgeIDs)
		}
	}
	metrics.ObservedCausalEdgeTotal, metrics.ShortestPathEdgeUnionTotal = len(observed), len(shortestUnion)
	for _, kind := range EdgeKinds() {
		metric := EdgeMetric{Kind: kind}
		for _, edge := range graph.Edges {
			if edge.Kind != kind {
				continue
			}
			metric.Eligible++
			if observed[edge.EdgeID] {
				metric.ObservedCausal++
			}
			for _, outcome := range outcomes {
				if contains(outcome.UpstreamEdgeIDs, edge.EdgeID) {
					if outcome.After == "OPEN" {
						metric.Blocking++
					}
					if outcome.After == "REFUTED" {
						metric.Refuting++
					}
					if recovered && outcome.After == "DISCHARGED" {
						metric.Discharge++
					}
				}
			}
		}
		metrics.ObservedBlockingEdgeTotal += metric.Blocking
		metrics.ObservedRefutingEdgeTotal += metric.Refuting
		metrics.ObservedRecoveryEdgeTotal += metric.Discharge
		metrics.EdgeMetrics = append(metrics.EdgeMetrics, metric)
	}
	return metrics
}
func distinctPropositions(graph Graph) int {
	seen := map[string]bool{}
	for _, claim := range graph.Nodes {
		seen[claim.PropositionDigest] = true
	}
	return len(seen)
}
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func decisionFor(states []string, evidence EvidenceReceipt, recovered bool) Decision {
	switch authorityResolution(evidence) {
	case "NET_REPOSITORY_STATE_CHANGED":
		return Decision{Value: "FAIL_CLOSED", Resolution: "AUTHORITY_CHANGED", Reason: "AUTHORITY/REPOSITORY_SNAPSHOT/NET_REPOSITORY_STATE_CHANGED", SemanticPromotionAuthorized: false}
	case "TRANSIENT_WRITE_AUTHORITY_UNKNOWN":
		return Decision{Value: "FAIL_CLOSED", Resolution: "AUTHORITY_UNKNOWN", Reason: evidence.Capability.Coordinate.Stage + "/" + evidence.Capability.Coordinate.Step + "/TRANSIENT_WRITE_AUTHORITY_UNKNOWN", SemanticPromotionAuthorized: false}
	}
	if allState(states, "DISCHARGED") {
		if recovered {
			return Decision{Value: "PASS", Resolution: "CAUSAL_RECOVERY_DISCHARGED", Reason: "APPEND_ONLY_EVIDENCE_RECOVERY", SemanticPromotionAuthorized: false}
		}
		return Decision{Value: "PASS", Resolution: "DIRECT_EVIDENCE_DISCHARGED", Reason: "CURRENT_EVIDENCE_PREDICATES_SATISFIED", SemanticPromotionAuthorized: false}
	}
	if anyState(states, "REFUTED") {
		if countState(states, "REFUTED") == 1 {
			return Decision{Value: "FAIL_CLOSED", Resolution: "DIRECT_REFUTATION", Reason: "ONLY_DIRECT_EXPLICIT_CONTRADICTION", SemanticPromotionAuthorized: false}
		}
		return Decision{Value: "FAIL_CLOSED", Resolution: "CAUSAL_REFUTATION", Reason: "EXPLICIT_CONTRADICTION_OR_FAILURE_ENTAILMENT", SemanticPromotionAuthorized: false}
	}
	return Decision{Value: "FAIL_CLOSED", Resolution: "UNRESOLVED_CLAIM", Reason: "UNKNOWN_REMAINS_OPEN", SemanticPromotionAuthorized: false}
}
func authorityResolution(evidence EvidenceReceipt) string {
	if evidence.Snapshot.RepositoryWrites != 0 || evidence.Snapshot.BeforeDigest != evidence.Snapshot.AfterDigest {
		return "NET_REPOSITORY_STATE_CHANGED"
	}
	if evidence.Capability.Status != CurrentEvidence || evidence.Capability.Provider == "" {
		return "TRANSIENT_WRITE_AUTHORITY_UNKNOWN"
	}
	return "NET_REPOSITORY_STATE_UNCHANGED"
}

func AuthorityCases() []AuthorityCase {
	cases := []AuthorityCase{
		{CaseID: "NET-SAME-CURRENT", NetworkState: "NET_SAME", CapabilityStatus: CurrentEvidence, ExpectedResolution: "NET_REPOSITORY_STATE_UNCHANGED"},
		{CaseID: "NET-CHANGED-CURRENT", NetworkState: "NET_CHANGED", CapabilityStatus: CurrentEvidence, ExpectedResolution: "NET_REPOSITORY_STATE_CHANGED"},
		{CaseID: "TRANSIENT-UNKNOWN", NetworkState: "TRANSIENT_UNKNOWN", CapabilityStatus: UnknownEvidence, ExpectedResolution: "TRANSIENT_WRITE_AUTHORITY_UNKNOWN"},
	}
	for i := range cases {
		cases[i].ObservedResolution = authorityResolutionForCase(cases[i].NetworkState, cases[i].CapabilityStatus)
	}
	return cases
}

func validateAuthorityCases(cases []AuthorityCase) error {
	if len(cases) != 3 {
		return fmt.Errorf("authority cases have %d cases, want 3", len(cases))
	}
	for _, value := range cases {
		if value.ExpectedResolution == "" || value.ExpectedResolution != value.ObservedResolution {
			return fmt.Errorf("authority case %q did not execute its expected resolution", value.CaseID)
		}
	}
	return nil
}

func authorityResolutionForCase(networkState string, capabilityStatus EvidenceStatus) string {
	if networkState == "NET_CHANGED" {
		return "NET_REPOSITORY_STATE_CHANGED"
	}
	if networkState == "TRANSIENT_UNKNOWN" || capabilityStatus != CurrentEvidence {
		return "TRANSIENT_WRITE_AUTHORITY_UNKNOWN"
	}
	return "NET_REPOSITORY_STATE_UNCHANGED"
}
func allState(states []string, target string) bool {
	for _, value := range states {
		if value != target {
			return false
		}
	}
	return true
}
func anyState(states []string, target string) bool {
	for _, value := range states {
		if value == target {
			return true
		}
	}
	return false
}
func countState(states []string, target string) int {
	total := 0
	for _, state := range states {
		if state == target {
			total++
		}
	}
	return total
}
func resolutionStates(values []Resolution) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.State
	}
	return result
}
func validateTransitionChain(transitions []Transition, head string) error {
	if len(transitions) == 0 || transitions[len(transitions)-1].TransitionDigest != head {
		return fmt.Errorf("transition head does not match chain")
	}
	previous := ""
	for i, transition := range transitions {
		if transition.Sequence != i+1 || transition.PreviousTransitionDigest != previous {
			return fmt.Errorf("transition %d predecessor mismatch", i+1)
		}
		digest, _ := transitionDigest(transition)
		if digest != transition.TransitionDigest {
			return fmt.Errorf("transition %d digest mismatch", i+1)
		}
		previous = transition.TransitionDigest
	}
	return nil
}
func sameStrings(left, right []string) bool {
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
