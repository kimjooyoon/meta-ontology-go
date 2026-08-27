package claimdependency

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	evidenceProcedure       = "RAW_ARTIFACT_OBSERVATION_BINDING_V3"
	observationSchema       = "gooo.meta.claim-dependency-observation/v3"
	observationBundleSchema = "gooo.meta.claim-dependency-observation-bundle/v2"
	observationProcedure    = "CI_TARGET_SPECIFIC_VALIDATOR_COMPARATOR_V4"
)

// BuildCurrentEvidence keeps the original provider API for callers that do
// not have an external target observation. Parsing/lowering a source is only a
// declared recipe in that case, so the resulting evidence is UNKNOWN.
func BuildCurrentEvidence(artifactPath, operation, capabilityPath, repositoryRoot, outputPath string) (EvidenceReceipt, error) {
	return BuildCurrentEvidenceForSource(artifactPath, artifactPath, operation, capabilityPath, repositoryRoot, outputPath, "")
}

// BuildCurrentEvidenceWithObservation binds a current evidence receipt to a
// separately produced observation of the target artifact. Operation is a
// CLAIMED_INPUT/REQUEST and never selects the observed predicate.
func BuildCurrentEvidenceWithObservation(artifactPath, operation, capabilityPath, repositoryRoot, outputPath, observationPath string) (EvidenceReceipt, error) {
	return BuildCurrentEvidenceForSource(artifactPath, artifactPath, operation, capabilityPath, repositoryRoot, outputPath, observationPath)
}

// BuildCurrentEvidenceForSource keeps the source graph and observed target
// separate. The source supplies claim identity; the artifact is only raw
// target input re-observed by the provider.
func BuildCurrentEvidenceForSource(sourcePath, artifactPath, operation, capabilityPath, repositoryRoot, outputPath, observationPath string) (EvidenceReceipt, error) {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return EvidenceReceipt{}, fmt.Errorf("provider source reconstruction: %w", err)
	}
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		return EvidenceReceipt{}, err
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
		if err := decodeStrictJSON(data, &bundle); err != nil {
			return EvidenceReceipt{}, fmt.Errorf("target observation decode: %w", err)
		}
		if err := validateObservationBundle(bundle, sourcePath, source, artifactPath, artifact, parsed.Graph); err != nil {
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
	claims := make([]EvidenceClaim, len(parsed.Graph.Nodes))
	for i, claim := range parsed.Graph.Nodes {
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
	var observationRaw []byte
	if observationPath != "" {
		observationRaw, err = os.ReadFile(observationPath)
		if err != nil {
			return EvidenceReceipt{}, fmt.Errorf("target observation raw bytes: %w", err)
		}
	}
	receipt := EvidenceReceipt{Schema: EvidenceSchema, Provider: "github-actions-current-evidence-provider/v4", SourcePath: sourcePath, SourceBytesDigest: digestBytes(source), SourceGraphDigest: parsed.Graph.Digest, ArtifactPath: artifactPath, ArtifactBytesDigest: digestBytes(artifact), Operation: operation, RequestStatus: "CLAIMED_INPUT", Procedure: evidenceProcedure, ObservationPath: observationPath, ObservationBundleDigest: observationBundleDigest, ObservationBundleRaw: observationRaw, Observations: observations, ObservedPredicate: predicate, ObservedValue: observedValue, Status: status, Coordinate: Coordinate{Stage: "OBSERVE", Step: "current-evidence-provider", Reason: observationReason(status, predicate)}, Claims: claims, Capability: capability, Snapshot: snapshot}
	receipt.Digest, err = evidenceReceiptDigest(receipt)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	return receipt, nil
}

func BuildObservationBundle(sourcePath string, source []byte, artifactPath, output, profile, contractPath, failureReceiptPath string) (ObservationBundle, error) {
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		return ObservationBundle{}, err
	}
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return ObservationBundle{}, fmt.Errorf("observer source reconstruction: %w", err)
	}
	contract, err := readValidatorContract(contractPath)
	if err != nil {
		return ObservationBundle{}, err
	}
	contractBytes, err := os.ReadFile(contractPath)
	if err != nil {
		return ObservationBundle{}, err
	}
	for _, claim := range parsed.Graph.Nodes {
		material, ok := contractClaim(contract, claim.ActivityName)
		if !ok || !claimIdentityMatchesContract(claim, material) {
			return ObservationBundle{}, fmt.Errorf("validator contract claim inventory does not match source claim %q", claim.ActivityName)
		}
	}
	actualDigest := digestBytes(artifact)
	failure := FailureReceipt{}
	if failureReceiptPath != "" {
		failureBytes, err := os.ReadFile(failureReceiptPath)
		if err != nil {
			return ObservationBundle{}, fmt.Errorf("failure receipt: %w", err)
		}
		if err := decodeStrictJSON(failureBytes, &failure); err != nil {
			return ObservationBundle{}, fmt.Errorf("failure receipt decode: %w", err)
		}
		if err := validateFailureReceipt(failure, sourcePath, source, artifactPath, artifact, parsed.Graph); err != nil {
			return ObservationBundle{}, err
		}
	}
	observations := []ObservationReceipt{}
	structural := []StructuralContradiction{}
	for _, claim := range parsed.Graph.Nodes {
		material, ok := contractClaim(contract, claim.ActivityName)
		if !ok {
			return ObservationBundle{}, fmt.Errorf("validator contract has no claim %q", claim.ActivityName)
		}
		row, rowDigest, ok := targetRow(artifact, claim.ActivityName)
		if !ok {
			continue
		}
		if claim.ValueProgram != material.ExpectedValueProgram && len(structural) == 0 {
			structural = append(structural, StructuralContradiction{ClaimID: claim.ClaimID, PropositionDigest: claim.PropositionDigest, ExpectedValue: material.ExpectedValueProgram, DeclaredValue: claim.ValueProgram, ProcedureID: material.ProcedureID})
		}
		if claimIdentityMatchesContract(claim, material) && claim.ValueProgram == material.ExpectedValueProgram && artifactPath == contract.ExpectedArtifactPath && actualDigest == contract.ExpectedArtifactDigest && rowDigest == material.TargetRowDigest {
			observed := claimObservationMaterial(claim, material, artifactPath, actualDigest, rowDigest)
			observations = append(observations, makeObservation("CLAIM", claim.ClaimID, claim.PropositionDigest, "", "", "", "", claim.Target, artifactPath, actualDigest, observed, observed, ObservationEvidence, ObservationEvidence, material.ProcedureID, "CURRENT_TARGET_PREDICATE_MATCH", string(row)))
		}
	}
	for _, edge := range parsed.Graph.Edges {
		from := indexOfClaim(edge.FromClaimID, parsed.Graph)
		to := indexOfClaim(edge.ToClaimID, parsed.Graph)
		if from < 0 || to < 0 {
			continue
		}
		fromContract, fromOK := contractClaim(contract, parsed.Graph.Nodes[from].ActivityName)
		toContract, toOK := contractClaim(contract, parsed.Graph.Nodes[to].ActivityName)
		if !fromOK || !toOK || artifactPath != contract.ExpectedArtifactPath || actualDigest != contract.ExpectedArtifactDigest {
			continue
		}
		_, fromRowDigest, fromRowOK := targetRow(artifact, parsed.Graph.Nodes[from].ActivityName)
		_, toRowDigest, toRowOK := targetRow(artifact, parsed.Graph.Nodes[to].ActivityName)
		fromExpected := fromRowOK && fromRowDigest == fromContract.TargetRowDigest
		toAlternate := toRowOK && toContract.AlternateRowDigest != "" && toRowDigest == toContract.AlternateRowDigest
		if edge.Kind == Contradicts && fromExpected && toAlternate {
			expected := edgeMaterial("contract", edge, parsed.Graph, contract, actualDigest, artifactPath, false)
			observed := edgeTargetMaterial("observed", edge, parsed.Graph, fromRowDigest, toRowDigest, artifactPath, actualDigest)
			observations = append(observations, makeObservation("EDGE", parsed.Graph.Nodes[to].ClaimID, parsed.Graph.Nodes[to].PropositionDigest, edge.EdgeID, edge.FromClaimID, edge.ToClaimID, edge.Kind, parsed.Graph.Nodes[to].Target, artifactPath, actualDigest, expected, observed, ObservationContradiction, ObservationContradiction, "CI_EDGE_TARGET_CONTRADICTION_COMPARATOR", "CONTRADICTS_TARGET_VALUE_OPPOSITE"))
		}
		fromAlternate := fromRowOK && fromContract.AlternateRowDigest != "" && fromRowDigest == fromContract.AlternateRowDigest
		if edge.Kind == FailureEntailment && fromAlternate && toAlternate && failureReceiptPath != "" && failure.EdgeID == edge.EdgeID {
			expected := edgeMaterial("contract", edge, parsed.Graph, contract, actualDigest, artifactPath, false)
			observed := edgeTargetMaterial("observed", edge, parsed.Graph, fromRowDigest, toRowDigest, artifactPath, actualDigest) + "|exit=" + strconv.Itoa(failure.ObservedExitCode) + "|result=" + failure.Result
			observations = append(observations, makeObservation("EDGE", parsed.Graph.Nodes[to].ClaimID, parsed.Graph.Nodes[to].PropositionDigest, edge.EdgeID, edge.FromClaimID, edge.ToClaimID, edge.Kind, parsed.Graph.Nodes[to].Target, artifactPath, actualDigest, expected, observed, ObservationFailure, ObservationFailure, "CI_EDGE_FAILURE_ANTECEDENT_PROCESS", digestBytes(append(failure.Stdout, failure.Stderr...))))
		}
	}
	if len(observations) == 0 {
		observations = append(observations, makeObservation("UNRELATED_ARTIFACT", "", "", "", "", "", "", TargetAddress{Artifact: artifactPath}, artifactPath, actualDigest, "validator:bound-target", fmt.Sprintf("observed:artifact_path=%s|artifact_bytes_digest=%s|output=%s", artifactPath, actualDigest, output), ObservationUnknown, ObservationUnknown, "NO_CLAIM_SCOPED_PREDICATE_MATCH", output))
	}
	for i := range observations {
		observations[i].Digest, err = observationReceiptDigest(observations[i])
		if err != nil {
			return ObservationBundle{}, err
		}
	}
	for i := range structural {
		structural[i].Digest = digestBytes([]byte(fmt.Sprintf("%s|%s|%s|%s", structural[i].ClaimID, structural[i].ExpectedValue, structural[i].DeclaredValue, structural[i].ProcedureID)))
	}
	bundle := ObservationBundle{Schema: observationBundleSchema, Provider: "github-actions-target-observer/v4", SourcePath: sourcePath, SourceDigest: digestBytes(source), ArtifactPath: artifactPath, ArtifactBytesDigest: actualDigest, ContractPath: contractPath, ContractDigest: digestBytes(contractBytes), ContractRaw: append([]byte(nil), contractBytes...), FailureReceiptPath: failureReceiptPath, Profile: profile, Observations: observations, StructuralContradictions: structural}
	if failureReceiptPath != "" {
		failureBytes, err := os.ReadFile(failureReceiptPath)
		if err != nil {
			return ObservationBundle{}, err
		}
		bundle.FailureReceiptDigest = digestBytes(failureBytes)
		bundle.FailureReceiptRaw = append([]byte(nil), failureBytes...)
	}
	bundle.Digest, err = observationBundleDigest(bundle)
	if err != nil {
		return ObservationBundle{}, err
	}
	return bundle, nil
}

func makeObservation(binding, claimID, propositionDigest, edgeID, fromClaimID, toClaimID string, edgeKind EdgeKind, target TargetAddress, artifactPath, targetBytesDigest, expectedValue, observedValue string, expectedPredicate, observedPredicate ObservationPredicate, procedure, reason, output string) ObservationReceipt {
	comparison := "MISMATCH"
	if expectedValue == observedValue {
		comparison = "MATCH"
	}
	return ObservationReceipt{Schema: observationSchema, Provider: "github-actions-target-observer/v4", Binding: binding, ClaimID: claimID, PropositionDigest: propositionDigest, EdgeID: edgeID, FromClaimID: fromClaimID, ToClaimID: toClaimID, EdgeKind: edgeKind, Target: target, TargetPath: artifactPath, TargetBytesDigest: targetBytesDigest, ExpectedPredicate: expectedPredicate, ExpectedValue: expectedValue, ObservedPredicate: observedPredicate, ObservedValue: observedValue, ComparisonResult: comparison, Procedure: procedure, ProcedureDigest: digestBytes([]byte(procedure)), Output: output, OutputDigest: digestBytes([]byte(output)), Coordinate: Coordinate{Stage: "OBSERVE", Step: "target-observer", Reason: reason}}
}

func targetRow(artifact []byte, activity string) ([]byte, string, bool) {
	prefix := []byte("activity " + activity + "(")
	for _, line := range bytes.Split(artifact, []byte("\n")) {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, prefix) {
			return append([]byte(nil), trimmed...), digestBytes(trimmed), true
		}
	}
	return nil, "", false
}

func claimIdentityMatchesContract(claim Claim, expected ValidatorClaim) bool {
	return claim.ClaimID == expected.ClaimID && claim.PropositionDigest == expected.PropositionDigest && claim.ActivityName == expected.ActivityName && reflect.DeepEqual(claim.Target, expected.ExpectedTarget)
}

func claimObservationMaterial(claim Claim, expected ValidatorClaim, artifactPath, artifactDigest, rowDigest string) string {
	return fmt.Sprintf("claim-observation|claim_id=%s|proposition_digest=%s|procedure_id=%s|target_row_digest=%s|artifact_path=%s|artifact_digest=%s|expected_value_program=%s", claim.ClaimID, claim.PropositionDigest, expected.ProcedureID, rowDigest, artifactPath, artifactDigest, expected.ExpectedValueProgram)
}

func edgeTargetMaterial(prefix string, edge Edge, graph Graph, fromRowDigest, toRowDigest, artifactPath, artifactDigest string) string {
	return fmt.Sprintf("%s|edge=%s|kind=%s|from=%s|to=%s|from_target_row_digest=%s|to_target_row_digest=%s|artifact_path=%s|artifact_bytes_digest=%s", prefix, edge.EdgeID, edge.Kind, edge.FromClaimID, edge.ToClaimID, fromRowDigest, toRowDigest, artifactPath, artifactDigest)
}

func validateObservationBundle(bundle ObservationBundle, sourcePath string, source []byte, artifactPath string, artifact []byte, graph Graph) error {
	if bundle.Schema != observationBundleSchema || bundle.Provider == "" || bundle.SourcePath != sourcePath || bundle.SourceDigest != digestBytes(source) || bundle.ArtifactPath != artifactPath || bundle.ArtifactBytesDigest != digestBytes(artifact) || bundle.ContractPath == "" || bundle.ContractDigest == "" || len(bundle.ContractRaw) == 0 || bundle.Profile == "" || bundle.Digest == "" || len(bundle.Observations) == 0 {
		return fmt.Errorf("target observation bundle identity or target binding is invalid")
	}
	if digestBytes(bundle.ContractRaw) != bundle.ContractDigest {
		return fmt.Errorf("validator contract bytes changed")
	}
	var contract ValidatorContract
	if err := decodeStrictJSON(bundle.ContractRaw, &contract); err != nil {
		return fmt.Errorf("embedded validator contract decode: %w", err)
	}
	if err := validateValidatorContract(contract); err != nil {
		return err
	}
	for _, claim := range graph.Nodes {
		material, ok := contractClaim(contract, claim.ActivityName)
		if !ok || !claimIdentityMatchesContract(claim, material) {
			return fmt.Errorf("embedded validator contract claim inventory does not match source graph")
		}
	}
	failure := FailureReceipt{}
	if bundle.FailureReceiptPath != "" {
		if len(bundle.FailureReceiptRaw) == 0 || digestBytes(bundle.FailureReceiptRaw) != bundle.FailureReceiptDigest {
			return fmt.Errorf("failure receipt bytes changed")
		}
		if err := decodeStrictJSON(bundle.FailureReceiptRaw, &failure); err != nil {
			return fmt.Errorf("failure receipt decode: %w", err)
		}
		if err := validateFailureReceipt(failure, sourcePath, source, artifactPath, artifact, graph); err != nil {
			return err
		}
	}
	if digest, err := observationBundleDigest(bundle); err != nil || digest != bundle.Digest {
		return fmt.Errorf("target observation bundle digest is invalid")
	}
	seen := map[string]bool{}
	for _, observation := range bundle.Observations {
		if err := validateObservation(observation, artifactPath, artifact, graph, contract, failure, bundle.FailureReceiptPath != ""); err != nil {
			return err
		}
		key := observation.Binding + "|" + observation.ClaimID + "|" + observation.EdgeID
		if seen[key] {
			return fmt.Errorf("target observation bundle has duplicate binding %q", key)
		}
		seen[key] = true
	}
	for _, finding := range bundle.StructuralContradictions {
		index := indexOfClaim(finding.ClaimID, graph)
		if index < 0 || finding.PropositionDigest != graph.Nodes[index].PropositionDigest || finding.ExpectedValue == finding.DeclaredValue || finding.ProcedureID == "" || finding.Digest != digestBytes([]byte(fmt.Sprintf("%s|%s|%s|%s", finding.ClaimID, finding.ExpectedValue, finding.DeclaredValue, finding.ProcedureID))) {
			return fmt.Errorf("structural contradiction is not source-bound")
		}
	}
	return nil
}

func validateObservation(value ObservationReceipt, artifactPath string, artifact []byte, graph Graph, contract ValidatorContract, failure FailureReceipt, hasFailure bool) error {
	if value.Schema != observationSchema || value.Provider == "" || value.TargetPath != artifactPath || value.Procedure == "" || value.ProcedureDigest != digestBytes([]byte(value.Procedure)) || value.OutputDigest != digestBytes([]byte(value.Output)) || value.Coordinate.Stage == "" || value.Digest == "" {
		return fmt.Errorf("target observation identity or target binding is invalid")
	}
	if value.TargetBytesDigest != digestBytes(artifact) || (value.ComparisonResult != "MATCH" && value.ComparisonResult != "MISMATCH") {
		return fmt.Errorf("target observation comparison is invalid")
	}
	if value.ComparisonResult == "MATCH" && value.ExpectedValue != value.ObservedValue || value.ComparisonResult == "MISMATCH" && value.ExpectedValue == value.ObservedValue {
		return fmt.Errorf("target observation comparison result does not match values")
	}
	switch value.Binding {
	case "CLAIM":
		claimIndex := indexOfClaim(value.ClaimID, graph)
		if claimIndex < 0 || value.PropositionDigest != graph.Nodes[claimIndex].PropositionDigest || !reflect.DeepEqual(value.Target, graph.Nodes[claimIndex].Target) || value.ExpectedPredicate != ObservationEvidence || value.ObservedPredicate != ObservationEvidence || value.ComparisonResult != "MATCH" {
			return fmt.Errorf("claim-scoped target observation is not bound to its claim")
		}
		material, ok := contractClaim(contract, graph.Nodes[claimIndex].ActivityName)
		row, rowDigest, rowOK := targetRow(artifact, graph.Nodes[claimIndex].ActivityName)
		expected := claimObservationMaterial(graph.Nodes[claimIndex], material, artifactPath, digestBytes(artifact), rowDigest)
		if !ok || !claimIdentityMatchesContract(graph.Nodes[claimIndex], material) || graph.Nodes[claimIndex].ValueProgram != material.ExpectedValueProgram || artifactPath != contract.ExpectedArtifactPath || digestBytes(artifact) != contract.ExpectedArtifactDigest || !rowOK || rowDigest != material.TargetRowDigest || value.ExpectedValue != expected || value.ObservedValue != expected || value.OutputDigest != digestBytes(row) {
			return fmt.Errorf("claim observation does not match external validator material")
		}
	case "EDGE":
		edgeIndex := indexOfEdge(value.EdgeID, graph)
		if edgeIndex < 0 {
			return fmt.Errorf("edge-scoped target observation references an unknown edge")
		}
		edge := graph.Edges[edgeIndex]
		to := indexOfClaim(edge.ToClaimID, graph)
		if to < 0 || value.FromClaimID != edge.FromClaimID || value.ToClaimID != edge.ToClaimID || value.EdgeKind != edge.Kind || value.ClaimID != edge.ToClaimID || value.PropositionDigest != graph.Nodes[to].PropositionDigest || !reflect.DeepEqual(value.Target, graph.Nodes[to].Target) || value.ExpectedPredicate != value.ObservedPredicate || value.ComparisonResult != "MISMATCH" || (edge.Kind == Contradicts && value.ObservedPredicate != ObservationContradiction) || (edge.Kind == FailureEntailment && value.ObservedPredicate != ObservationFailure) {
			return fmt.Errorf("edge-scoped target observation is not bound to its edge")
		}
		from := indexOfClaim(edge.FromClaimID, graph)
		if from < 0 {
			return fmt.Errorf("edge observation references an unknown upstream claim")
		}
		fromContract, fromOK := contractClaim(contract, graph.Nodes[from].ActivityName)
		toContract, toOK := contractClaim(contract, graph.Nodes[to].ActivityName)
		if from < 0 || !fromOK || !toOK || artifactPath != contract.ExpectedArtifactPath || digestBytes(artifact) != contract.ExpectedArtifactDigest {
			return fmt.Errorf("edge observation does not match external validator material")
		}
		_, fromRowDigest, fromRowOK := targetRow(artifact, graph.Nodes[from].ActivityName)
		_, toRowDigest, toRowOK := targetRow(artifact, graph.Nodes[to].ActivityName)
		if artifactPath != contract.ExpectedArtifactPath || digestBytes(artifact) != contract.ExpectedArtifactDigest || !fromRowOK || !toRowOK {
			return fmt.Errorf("edge observation target rows are not externally bound")
		}
		if edge.Kind == Contradicts && (fromRowDigest != fromContract.TargetRowDigest || toContract.AlternateRowDigest == "" || toRowDigest != toContract.AlternateRowDigest) {
			return fmt.Errorf("contradiction edge observation has the wrong direction or value binding")
		}
		if edge.Kind == FailureEntailment && (fromContract.AlternateRowDigest == "" || toContract.AlternateRowDigest == "" || fromRowDigest != fromContract.AlternateRowDigest || toRowDigest != toContract.AlternateRowDigest) {
			return fmt.Errorf("failure edge observation has the wrong failure binding")
		}
		expected := edgeMaterial("contract", edge, graph, contract, digestBytes(artifact), artifactPath, false)
		observed := edgeTargetMaterial("observed", edge, graph, fromRowDigest, toRowDigest, artifactPath, digestBytes(artifact))
		if edge.Kind == Contradicts && (value.ExpectedValue != expected || value.ObservedValue != edgeTargetMaterial("observed", edge, graph, fromRowDigest, toRowDigest, artifactPath, digestBytes(artifact)) || value.Output != "CONTRADICTS_TARGET_VALUE_OPPOSITE") {
			return fmt.Errorf("contradiction edge observation is not a structured opposite-value comparison")
		}
		if edge.Kind == FailureEntailment {
			if !hasFailure || failure.EdgeID != edge.EdgeID || failure.ObservedExitCode == 0 || failure.Result != "NONZERO_EXIT" || value.ExpectedValue != expected || value.ObservedValue != observed+"|exit="+strconv.Itoa(failure.ObservedExitCode)+"|result="+failure.Result || value.Output != digestBytes(append(failure.Stdout, failure.Stderr...)) {
				return fmt.Errorf("failure edge observation lacks an exact non-zero failure receipt")
			}
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

func claimMaterial(prefix string, claim Claim, target TargetAddress, artifactPath, artifactDigest, valueProgram string) string {
	return fmt.Sprintf("%s|activity=%s|target_inputs=%s|target_output=%s|target_artifact=%s|target_path=%s|artifact_bytes_digest=%s|value_program=%s", prefix, claim.ActivityName, strings.Join(target.Inputs, ","), target.Output, target.Artifact, artifactPath, artifactDigest, valueProgram)
}

func claimMatchesExpected(claim Claim, expected ValidatorClaim, contract ValidatorContract, artifactPath, artifactDigest string) bool {
	return reflect.DeepEqual(claim.Target, expected.ExpectedTarget) && artifactPath == contract.ExpectedArtifactPath && artifactDigest == contract.ExpectedArtifactDigest && claim.ValueProgram == expected.ExpectedValueProgram
}

func claimMatchesAlternate(claim Claim, expected ValidatorClaim, contract ValidatorContract, artifactPath, artifactDigest string) bool {
	return expected.AlternateValueProgram != "" && reflect.DeepEqual(claim.Target, expected.ExpectedTarget) && artifactPath == contract.ExpectedArtifactPath && artifactDigest == contract.ExpectedArtifactDigest && claim.ValueProgram == expected.AlternateValueProgram
}

func edgeMaterial(prefix string, edge Edge, graph Graph, contract ValidatorContract, artifactDigest, artifactPath string, observed bool) string {
	from := indexOfClaim(edge.FromClaimID, graph)
	to := indexOfClaim(edge.ToClaimID, graph)
	fromContract, _ := contractClaim(contract, graph.Nodes[from].ActivityName)
	toContract, _ := contractClaim(contract, graph.Nodes[to].ActivityName)
	fromProgram, toProgram := fromContract.ExpectedValueProgram, toContract.ExpectedValueProgram
	if observed {
		fromProgram, toProgram = graph.Nodes[from].ValueProgram, graph.Nodes[to].ValueProgram
	}
	return fmt.Sprintf("%s|edge=%s|kind=%s|from=%s|to=%s|from_value_program=%s|to_value_program=%s|artifact_path=%s|artifact_bytes_digest=%s", prefix, edge.EdgeID, edge.Kind, edge.FromClaimID, edge.ToClaimID, fromProgram, toProgram, artifactPath, artifactDigest)
}

func validateFailureReceipt(value FailureReceipt, sourcePath string, source []byte, artifactPath string, artifact []byte, graph Graph) error {
	if value.Schema != FailureReceiptSchema || value.Provider == "" || value.SourcePath != sourcePath || value.SourceDigest != digestBytes(source) || value.ArtifactPath != artifactPath || value.ArtifactBytesDigest != digestBytes(artifact) || value.EdgeKind != FailureEntailment || value.ObservedExitCode != 1 || value.Result != "NONZERO_EXIT" || value.Procedure != failureProcedure || value.Executable != "CI_EDGE_SPECIFIC_FAILURE_COMPARATOR" || value.ExecutableDigest == "" || digestBytes(value.ExecutableRaw) != value.ExecutableDigest || len(value.Argv) != 5 || value.Argv[0] != "-comparator" || value.Argv[1] != "-input" || value.Argv[3] != "-edge-id" || value.Argv[4] != value.EdgeID || value.StdoutDigest != digestBytes(value.Stdout) || value.StderrDigest != digestBytes(value.Stderr) || value.ProcedureDigest != failureProcedureDigest(value) || value.Coordinate.Stage == "" || value.Digest == "" {
		return fmt.Errorf("failure receipt is not an observed non-zero process result")
	}
	if !bytes.Contains(value.ExecutableRaw, []byte("FAILURE_ANTECEDENT")) || !bytes.Contains(value.ExecutableRaw, []byte("EDGE_SPECIFIC")) || !bytes.Contains(value.Stdout, []byte("FAILURE_ANTECEDENT_OBSERVED")) || !bytes.Contains(value.Stdout, []byte("EDGE_SPECIFIC")) {
		return fmt.Errorf("failure receipt executable is not the fixed edge comparator")
	}
	edgeIndex := indexOfEdge(value.EdgeID, graph)
	if edgeIndex < 0 {
		return fmt.Errorf("failure receipt references an unknown edge")
	}
	edge := graph.Edges[edgeIndex]
	to := indexOfClaim(edge.ToClaimID, graph)
	if to < 0 || edge.Kind != FailureEntailment || value.FromClaimID != edge.FromClaimID || value.ToClaimID != edge.ToClaimID || !reflect.DeepEqual(value.Target, graph.Nodes[to].Target) || len(value.InputTargets) != 2 {
		return fmt.Errorf("failure receipt edge binding is invalid")
	}
	from := indexOfClaim(edge.FromClaimID, graph)
	_, fromTargetDigest, fromTargetOK := targetRow(artifact, graph.Nodes[from].ActivityName)
	_, toTargetDigest, toTargetOK := targetRow(artifact, graph.Nodes[to].ActivityName)
	if from < 0 || !fromTargetOK || !toTargetOK || !failureInputMatches(value.InputTargets[0], graph.Nodes[from], artifactPath, digestBytes(artifact), fromTargetDigest) || !failureInputMatches(value.InputTargets[1], graph.Nodes[to], artifactPath, digestBytes(artifact), toTargetDigest) {
		return fmt.Errorf("failure receipt input targets are not bound to the edge propositions")
	}
	fromRow, _, fromRowOK := targetRow(artifact, graph.Nodes[from].ActivityName)
	toRow, _, toRowOK := targetRow(artifact, graph.Nodes[to].ActivityName)
	if !fromRowOK || !toRowOK || !bytes.Contains(fromRow, []byte(graph.Nodes[from].ValueProgram)) || !bytes.Contains(toRow, []byte(graph.Nodes[to].ValueProgram)) {
		return fmt.Errorf("failure receipt process did not consume the edge target outputs")
	}
	digest, err := failureReceiptDigest(value)
	if err != nil || digest != value.Digest {
		return fmt.Errorf("failure receipt digest is invalid")
	}
	return nil
}

func failureInputMatches(input FailureInput, claim Claim, artifactPath, artifactDigest, targetOutputDigest string) bool {
	return input.ClaimID == claim.ClaimID && input.PropositionDigest == claim.PropositionDigest && reflect.DeepEqual(input.Target, claim.Target) && input.TargetOutputDigest == targetOutputDigest && input.ValueProgram == claim.ValueProgram && input.ArtifactPath == artifactPath && input.ArtifactDigest == artifactDigest
}

func failureProcedureDigest(value FailureReceipt) string {
	parts := []string{value.Procedure, value.ExecutableDigest}
	parts = append(parts, value.Argv...)
	for _, input := range value.InputTargets {
		parts = append(parts, input.ClaimID, input.PropositionDigest, input.ValueProgram, input.ArtifactPath, input.ArtifactDigest, input.Target.Output, input.Target.Artifact, strings.Join(input.Target.Inputs, ","))
	}
	return digestBytes([]byte(strings.Join(parts, "|")))
}

// BuildFailureReceipt is called only after the CI helper has captured the OS
// result of the fixed edge comparator. All source claims and target addresses
// are copied into the receipt so a caller cannot attach an unrelated exit.
func BuildFailureReceipt(sourcePath string, source []byte, artifactPath, edgeID, executable string, executableBytes []byte, argv []string, stdout, stderr []byte, exitCode int) (FailureReceipt, error) {
	if exitCode == 0 {
		return FailureReceipt{}, fmt.Errorf("failure receipt requires a non-zero observed exit")
	}
	if exitCode != 1 {
		return FailureReceipt{}, fmt.Errorf("failure comparator must return the observed edge-specific exit 1")
	}
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return FailureReceipt{}, err
	}
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		return FailureReceipt{}, err
	}
	index := indexOfEdge(edgeID, parsed.Graph)
	if index < 0 || parsed.Graph.Edges[index].Kind != FailureEntailment {
		return FailureReceipt{}, fmt.Errorf("failure receipt requires an exact FAILURE_ENTAILMENT edge")
	}
	edge := parsed.Graph.Edges[index]
	to := indexOfClaim(edge.ToClaimID, parsed.Graph)
	from := indexOfClaim(edge.FromClaimID, parsed.Graph)
	if from < 0 || to < 0 || executable == "" || len(executableBytes) == 0 || len(argv) == 0 {
		return FailureReceipt{}, fmt.Errorf("failure comparator execution binding is incomplete")
	}
	fromRow, fromRowDigest, fromRowOK := targetRow(artifact, parsed.Graph.Nodes[from].ActivityName)
	toRow, toRowDigest, toRowOK := targetRow(artifact, parsed.Graph.Nodes[to].ActivityName)
	if !fromRowOK || !toRowOK || !bytes.Contains(fromRow, []byte(parsed.Graph.Nodes[from].ValueProgram)) || !bytes.Contains(toRow, []byte(parsed.Graph.Nodes[to].ValueProgram)) {
		return FailureReceipt{}, fmt.Errorf("failure comparator input does not match source-derived target outputs")
	}
	receipt := FailureReceipt{Schema: FailureReceiptSchema, Provider: "github-actions-failure-observer/v2", SourcePath: sourcePath, SourceDigest: digestBytes(source), ArtifactPath: artifactPath, ArtifactBytesDigest: digestBytes(artifact), EdgeID: edge.EdgeID, FromClaimID: edge.FromClaimID, ToClaimID: edge.ToClaimID, EdgeKind: edge.Kind, Target: parsed.Graph.Nodes[to].Target, Procedure: failureProcedure, Executable: executable, ExecutableDigest: digestBytes(executableBytes), ExecutableRaw: append([]byte(nil), executableBytes...), Argv: append([]string(nil), argv...), InputTargets: []FailureInput{failureInputFor(parsed.Graph.Nodes[from], artifactPath, digestBytes(artifact), fromRowDigest), failureInputFor(parsed.Graph.Nodes[to], artifactPath, digestBytes(artifact), toRowDigest)}, Stdout: append([]byte(nil), stdout...), StdoutDigest: digestBytes(stdout), Stderr: append([]byte(nil), stderr...), StderrDigest: digestBytes(stderr), ObservedExitCode: exitCode, Result: "NONZERO_EXIT", Coordinate: Coordinate{Stage: "OBSERVE", Step: "failure-antecedent-process", Reason: "OS_OBSERVED_EDGE_SPECIFIC_NONZERO_EXIT"}}
	receipt.ProcedureDigest = failureProcedureDigest(receipt)
	receipt.Digest, err = failureReceiptDigest(receipt)
	if err != nil {
		return FailureReceipt{}, err
	}
	return receipt, nil
}

func failureInputFor(claim Claim, artifactPath, artifactDigest, targetOutputDigest string) FailureInput {
	return FailureInput{ClaimID: claim.ClaimID, PropositionDigest: claim.PropositionDigest, Target: claim.Target, TargetOutputDigest: targetOutputDigest, ValueProgram: claim.ValueProgram, ArtifactPath: artifactPath, ArtifactDigest: artifactDigest}
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
	return fmt.Sprintf("bundle:%s|source_digest:%s|artifact_path:%s|artifact_bytes_digest:%s|contract_digest:%s|observation_total:%d", bundle.Digest, bundle.SourceDigest, bundle.ArtifactPath, bundle.ArtifactBytesDigest, bundle.ContractDigest, len(bundle.Observations))
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
	if err := decodeStrictJSON(data, &capability); err != nil {
		return CapabilityEvidence{}, fmt.Errorf("capability evidence decode: %w", err)
	}
	if capability.Provider == "" || capability.Permission == "" || capability.Coordinate.Stage == "" || capability.Toolchain.Name != "go" || capability.Toolchain.Version != "go1.27.0" {
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
	provenance := fmt.Sprintf("source-semantic:%s|evidence:%s|producer:%s|consumer:%s", semanticDigest, evidence.Digest, ProducerID, ConsumerID)
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
	if evidence.Schema != EvidenceSchema || (evidence.Status != CurrentEvidence && evidence.Status != UnknownEvidence) || evidence.Provider == "" || evidence.SourcePath == "" || evidence.SourceBytesDigest == "" || evidence.SourceGraphDigest == "" || evidence.ArtifactPath == "" || evidence.ArtifactBytesDigest == "" || evidence.Digest == "" || evidence.RequestStatus != "CLAIMED_INPUT" || evidence.Procedure != evidenceProcedure {
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
	if err != nil || capDigest != evidence.Capability.Digest || evidence.Capability.Status != CurrentEvidence || evidence.Capability.Toolchain.Name != "go" || evidence.Capability.Toolchain.Version != "go1.27.0" {
		return fmt.Errorf("capability evidence is invalid")
	}
	evidenceSource, err := os.ReadFile(evidence.SourcePath)
	if err != nil {
		return fmt.Errorf("producer cannot re-observe evidence source: %w", err)
	}
	evidenceGraph, err := graphFromSource(evidenceSource, evidence.SourcePath)
	if err != nil {
		return fmt.Errorf("producer evidence source reconstruction: %w", err)
	}
	if digestBytes(evidenceSource) != evidence.SourceBytesDigest || evidenceGraph.Graph.Digest != evidence.SourceGraphDigest {
		return fmt.Errorf("evidence source bytes or graph digest changed")
	}
	artifact, err := os.ReadFile(evidence.ArtifactPath)
	if err != nil {
		return fmt.Errorf("producer cannot re-observe artifact: %w", err)
	}
	if digestBytes(artifact) != evidence.ArtifactBytesDigest {
		return fmt.Errorf("artifact bytes digest changed")
	}
	observations := []ObservationReceipt(nil)
	var observedBundle ObservationBundle
	if evidence.ObservationPath != "" {
		if len(evidence.ObservationBundleRaw) == 0 || digestBytes(evidence.ObservationBundleRaw) != evidence.ObservationBundleDigest {
			return fmt.Errorf("evidence does not contain durable target observation bytes")
		}
		if err := decodeStrictJSON(evidence.ObservationBundleRaw, &observedBundle); err != nil {
			return fmt.Errorf("target observation bundle decode: %w", err)
		}
		if err := validateObservationBundle(observedBundle, evidence.SourcePath, evidenceSource, evidence.ArtifactPath, artifact, evidenceGraph.Graph); err != nil {
			return err
		}
		if observedBundle.Digest != evidence.ObservationBundleDigest || !reflect.DeepEqual(observedBundle.Observations, evidence.Observations) {
			return fmt.Errorf("embedded target observation bundle differs from raw bundle")
		}
		observations = observedBundle.Observations
	} else if len(evidence.Observations) != 0 || evidence.ObservationBundleDigest != "" || len(evidence.ObservationBundleRaw) != 0 {
		return fmt.Errorf("evidence has observations without a raw observation bundle")
	}

	if len(evidence.Claims) != evidenceGraph.Graph.NodeTotal {
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
		expectedValue = observationBundleValue(observedBundle)
	}
	if evidence.ObservedPredicate != expectedPredicate || evidence.ObservedValue != expectedValue {
		return fmt.Errorf("evidence predicate is not computed by claim-scoped observation procedure")
	}
	for i, claim := range evidenceGraph.Graph.Nodes {
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
	_ = parsed // current source is checked by Evaluate; evidence is independently replayed above.
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
			relation := edgeRelation(edge.Kind, states[from], local[i].Predicate, true, edgeObservationActive(edge, evidence.Observations, ObservationContradiction, graph), edgeObservationActive(edge, evidence.Observations, ObservationFailure, graph))
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

func edgeObservationActive(edge Edge, observations []ObservationReceipt, predicate ObservationPredicate, graph Graph) bool {
	to := indexOfClaim(edge.ToClaimID, graph)
	if to < 0 {
		return false
	}
	for _, observation := range observations {
		if observation.Binding == "EDGE" && observation.EdgeID == edge.EdgeID && observation.FromClaimID == edge.FromClaimID && observation.ToClaimID == edge.ToClaimID && observation.EdgeKind == edge.Kind && observation.ClaimID == edge.ToClaimID && observation.PropositionDigest == graph.Nodes[to].PropositionDigest && reflect.DeepEqual(observation.Target, graph.Nodes[to].Target) && observation.ObservedPredicate == predicate {
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
			if edge.Kind == Requires && from >= 0 && edgeRelation(edge.Kind, states[from], local, true, edgeObservationActive(edge, observations, ObservationContradiction, graph), edgeObservationActive(edge, observations, ObservationFailure, graph)) == relationDischarged {
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
		if from >= 0 && (edge.Kind == Supports || edge.Kind == Requires) && edgeRelation(edge.Kind, states[from], local, true, edgeObservationActive(edge, observations, ObservationContradiction, graph), edgeObservationActive(edge, observations, ObservationFailure, graph)) == relationOpen && (states[from] == "OPEN" || states[from] == "REFUTED") {
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
		responsibility, owner := failureAttribution(i, states[i], causePath, graph)
		resolution := Resolution{ClaimID: claim.ClaimID, Axis: claim.Axis, PropositionDigest: claim.PropositionDigest, State: states[i], Kind: resolutionKind(i, states[i], outcomes[i]), ObservedEvent: outcomes[i].Event, Coordinate: outcomes[i].Coordinate, EvidenceDigest: outcomes[i].EvidenceDigest, Provenance: provenance, FailureResponsibility: responsibility, FailureOwnerClaimID: owner, CausePath: causePath, CauseEdgeIDs: edgeIDs, CauseEdgeKinds: kinds, CauseTransitionDigests: transitionDigests, CauseCoordinate: coordinatePointer(outcomes[i].Coordinate)}
		if states[i] == "OPEN" {
			resolution.MissingEvidenceIDs = []string{"evidence:" + claim.ClaimID}
			resolution.BlockedByClaimIDs, resolution.BlockedByEdgeIDs = blockedFrontier(i, graph, states)
		}
		result[i] = resolution
	}
	return result
}

func failureAttribution(index int, state string, path []string, graph Graph) (string, string) {
	if state == "DISCHARGED" {
		return "N/A", ""
	}
	if index == 0 {
		return "DIRECT_CLAIM", graph.Nodes[index].ClaimID
	}
	if len(path) > 1 {
		return "UPSTREAM_CLAIM", path[0]
	}
	for _, edge := range incomingEdges(index, graph) {
		from := indexOfClaim(edge.FromClaimID, graph)
		if from >= 0 && (graph.Nodes[from].ClaimID != graph.Nodes[index].ClaimID) {
			return "UPSTREAM_CLAIM", graph.Nodes[from].ClaimID
		}
	}
	return "DIRECT_CLAIM", graph.Nodes[index].ClaimID
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
