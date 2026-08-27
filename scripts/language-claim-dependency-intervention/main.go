package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

type intervention struct {
	Name                            string   `json:"name"`
	Kind                            string   `json:"kind"`
	BaselineSourceDigest            string   `json:"baseline_source_digest"`
	InterventionSourceDigest        string   `json:"intervention_source_digest"`
	BaselineSemanticDigest          string   `json:"baseline_semantic_digest"`
	InterventionSemanticDigest      string   `json:"intervention_semantic_digest"`
	BaselineGraphDigest             string   `json:"baseline_graph_digest"`
	InterventionGraphDigest         string   `json:"intervention_graph_digest"`
	BaselineEvidenceDigest          string   `json:"baseline_evidence_digest"`
	InterventionEvidenceDigest      string   `json:"intervention_evidence_digest"`
	BaselineDecision                string   `json:"baseline_decision"`
	InterventionDecision            string   `json:"intervention_decision"`
	BaselineStates                  []string `json:"baseline_states"`
	InterventionStates              []string `json:"intervention_states"`
	BaselineTransitionDigests       []string `json:"baseline_transition_digests"`
	InterventionTransitionDigests   []string `json:"intervention_transition_digests"`
	BaselineCausePath               []string `json:"baseline_cause_path"`
	InterventionCausePath           []string `json:"intervention_cause_path"`
	SemanticDigestChanged           bool     `json:"semantic_digest_changed"`
	GraphDigestChanged              bool     `json:"graph_digest_changed"`
	EvidenceDigestChanged           bool     `json:"evidence_digest_changed"`
	StateTransitionChanged          bool     `json:"state_transition_changed"`
	CausePathChanged                bool     `json:"cause_path_changed"`
	DecisionChanged                 bool     `json:"decision_changed"`
	EdgeTypeChanged                 bool     `json:"edge_type_changed"`
	BaselineReadOnly                bool     `json:"baseline_read_only"`
	InterventionReadOnly            bool     `json:"intervention_read_only"`
	BaselineRepositoryWrites        int      `json:"baseline_repository_writes"`
	InterventionRepositoryWrites    int      `json:"intervention_repository_writes"`
	BaselineAuthorityResolution     string   `json:"baseline_authority_resolution"`
	InterventionAuthorityResolution string   `json:"intervention_authority_resolution"`
}

type report struct {
	Schema                                     string                   `json:"schema"`
	Interventions                              []intervention           `json:"interventions"`
	EvidenceProvenanceCases                    []evidenceProvenanceCase `json:"evidence_provenance_cases"`
	CommentOnlySemanticPreservationNumerator   int                      `json:"comment_only_semantic_preservation_numerator"`
	CommentOnlySemanticPreservationDenominator int                      `json:"comment_only_semantic_preservation_denominator"`
}

type evidenceProvenanceCase struct {
	Name                            string   `json:"name"`
	BaselineRequest                 string   `json:"baseline_request"`
	InterventionRequest             string   `json:"intervention_request"`
	BaselineArtifactPath            string   `json:"baseline_artifact_path"`
	InterventionArtifactPath        string   `json:"intervention_artifact_path"`
	BaselineArtifactBytesDigest     string   `json:"baseline_artifact_bytes_digest"`
	InterventionArtifactBytesDigest string   `json:"intervention_artifact_bytes_digest"`
	BaselineEvidenceDigest          string   `json:"baseline_evidence_digest"`
	InterventionEvidenceDigest      string   `json:"intervention_evidence_digest"`
	BaselineDecision                string   `json:"baseline_decision"`
	InterventionDecision            string   `json:"intervention_decision"`
	BaselineStates                  []string `json:"baseline_states"`
	InterventionStates              []string `json:"intervention_states"`
	BaselineTransitionDigests       []string `json:"baseline_transition_digests"`
	InterventionTransitionDigests   []string `json:"intervention_transition_digests"`
	EvidenceDigestChanged           bool     `json:"evidence_digest_changed"`
	StateTransitionChanged          bool     `json:"state_transition_changed"`
	DecisionChanged                 bool     `json:"decision_changed"`
}

func main() {
	baselinePath := flag.String("baseline", "", "baseline Gooo source")
	semanticPath := flag.String("semantic", "", "VALUE_PROGRAM intervention source")
	edgePath := flag.String("edge", "", "edge-type intervention source")
	commentPath := flag.String("comment", "", "comment-only intervention source")
	repoRoot := flag.String("repo-root", "", "repository root")
	capability := flag.String("capability", "", "current capability evidence")
	observationDir := flag.String("observation-dir", "", "raw target observation directory")
	outputPath := flag.String("output", "", "intervention artifact")
	flag.Parse()
	if *baselinePath == "" || *semanticPath == "" || *edgePath == "" || *commentPath == "" || *repoRoot == "" || *capability == "" || *observationDir == "" || *outputPath == "" {
		fail("-baseline, -semantic, -edge, -comment, -repo-root, -capability, -observation-dir, and -output are required")
	}
	baseline, semanticSource, edgeSource, commentSource := read(*baselinePath), read(*semanticPath), read(*edgePath), read(*commentPath)
	mainPath := filepath.Join(filepath.Dir(*baselinePath), "main.gooo")
	refutedPath := filepath.Join(filepath.Dir(*baselinePath), "refuted.gooo")
	items := []intervention{
		compare("source-only", "VALUE_PROGRAM", read(mainPath), mainPath, semanticSource, *semanticPath, "acceptance", "same", *repoRoot, *capability, *observationDir),
		compare("observation-only", "OBSERVATION", read(mainPath), mainPath, read(mainPath), mainPath, "acceptance", "availability", *repoRoot, *capability, *observationDir),
		compare("edge-only", "EDGE_KIND", read(refutedPath), refutedPath, edgeSource, *edgePath, "contradiction", "same", *repoRoot, *capability, *observationDir),
		compare("comment-only", "COMMENT_ONLY", baseline, *baselinePath, commentSource, *commentPath, "availability", "same", *repoRoot, *capability, *observationDir),
	}
	provenance := []evidenceProvenanceCase{
		provenanceCompare("caller-flag-only", "acceptance", "availability", read(mainPath), mainPath, read(mainPath), mainPath, *repoRoot, *capability, *observationDir, true),
		provenanceCompare("observation-artifact-changed", "acceptance", "acceptance", read(mainPath), mainPath, semanticSource, *semanticPath, *repoRoot, *capability, *observationDir, false),
		provenanceCompare("observation-absent", "acceptance", "availability", read(mainPath), mainPath, read(mainPath), mainPath, *repoRoot, *capability, *observationDir, false),
	}
	commentPreserved := 0
	for _, item := range items {
		if item.Name == "comment-only" && item.BaselineSourceDigest != item.InterventionSourceDigest && !item.SemanticDigestChanged && !item.StateTransitionChanged && !item.DecisionChanged {
			commentPreserved++
		}
	}
	writeJSON(*outputPath, report{Schema: "gooo.meta.claim-dependency-intervention/v2", Interventions: items, EvidenceProvenanceCases: provenance, CommentOnlySemanticPreservationNumerator: commentPreserved, CommentOnlySemanticPreservationDenominator: 1})
	for _, item := range items {
		fmt.Printf("intervention=%s semantic_digest_changed=%t evidence_digest_changed=%t state_transition_changed=%t decision_changed=%t edge_type_changed=%t authority=%s/%s writes=%d/%d\n", item.Name, item.SemanticDigestChanged, item.EvidenceDigestChanged, item.StateTransitionChanged, item.DecisionChanged, item.EdgeTypeChanged, item.BaselineAuthorityResolution, item.InterventionAuthorityResolution, item.BaselineRepositoryWrites, item.InterventionRepositoryWrites)
	}
}

func provenanceCompare(name, baselineRequest, interventionRequest string, baselineSource []byte, baselinePath string, interventionSource []byte, interventionPath, repoRoot, capability, observationDir string, reuseEvidence bool) evidenceProvenanceCase {
	baselineEvidence := evidence(baselinePath, baselineRequest, repoRoot, capability, observationDir)
	interventionEvidence := baselineEvidence
	if reuseEvidence {
		interventionEvidence = evidenceWithObservation(baselinePath, interventionRequest, repoRoot, capability, filepath.Join(observationDir, "accepted.json"))
	} else {
		interventionEvidence = evidence(interventionPath, interventionRequest, repoRoot, capability, observationDir)
	}
	baselineReceipt, err := claimdependency.Evaluate(baselineSource, baselinePath, baselineEvidence, nil)
	if err != nil {
		fail(err.Error())
	}
	interventionReceipt, err := claimdependency.Evaluate(interventionSource, interventionPath, interventionEvidence, nil)
	if err != nil {
		fail(err.Error())
	}
	baseStates, interventionStates := states(baselineReceipt), states(interventionReceipt)
	baseTransitions, interventionTransitions := transitionDigests(baselineReceipt), transitionDigests(interventionReceipt)
	return evidenceProvenanceCase{Name: name, BaselineRequest: baselineRequest, InterventionRequest: interventionRequest, BaselineArtifactPath: baselineEvidence.ArtifactPath, InterventionArtifactPath: interventionEvidence.ArtifactPath, BaselineArtifactBytesDigest: baselineEvidence.ArtifactBytesDigest, InterventionArtifactBytesDigest: interventionEvidence.ArtifactBytesDigest, BaselineEvidenceDigest: baselineEvidence.Digest, InterventionEvidenceDigest: interventionEvidence.Digest, BaselineDecision: baselineReceipt.Decision.Value + ":" + baselineReceipt.Decision.Resolution, InterventionDecision: interventionReceipt.Decision.Value + ":" + interventionReceipt.Decision.Resolution, BaselineStates: baseStates, InterventionStates: interventionStates, BaselineTransitionDigests: baseTransitions, InterventionTransitionDigests: interventionTransitions, EvidenceDigestChanged: baselineEvidence.Digest != interventionEvidence.Digest, StateTransitionChanged: !reflect.DeepEqual(baseStates, interventionStates) || !reflect.DeepEqual(baseTransitions, interventionTransitions), DecisionChanged: baselineReceipt.Decision != interventionReceipt.Decision}
}

func compare(name, kind string, baseline []byte, baselinePath string, changed []byte, changedPath, baseOperation, changedOperation, repoRoot, capability, observationDir string) intervention {
	baseEvidence := evidence(baselinePath, baseOperation, repoRoot, capability, observationDir)
	changedEvidence := baseEvidence
	if changedOperation != "same" && name != "source-only" {
		changedEvidence = evidence(changedPath, changedOperation, repoRoot, capability, observationDir)
	}
	baseReceipt, err := claimdependency.Evaluate(baseline, baselinePath, baseEvidence, nil)
	if err != nil {
		fail(err.Error())
	}
	changedReceipt, err := claimdependency.Evaluate(changed, changedPath, changedEvidence, nil)
	if err != nil {
		fail(err.Error())
	}
	baseStates, changedStates := states(baseReceipt), states(changedReceipt)
	baseTransitions, changedTransitions := transitionDigests(baseReceipt), transitionDigests(changedReceipt)
	baseCause, changedCause := lastCause(baseReceipt), lastCause(changedReceipt)
	return intervention{Name: name, Kind: kind, BaselineSourceDigest: baseReceipt.Subject.SourceDigest, InterventionSourceDigest: changedReceipt.Subject.SourceDigest, BaselineSemanticDigest: baseReceipt.Subject.SemanticDigest, InterventionSemanticDigest: changedReceipt.Subject.SemanticDigest, BaselineGraphDigest: baseReceipt.Graph.Digest, InterventionGraphDigest: changedReceipt.Graph.Digest, BaselineEvidenceDigest: baseReceipt.EvidenceDigest, InterventionEvidenceDigest: changedReceipt.EvidenceDigest, BaselineDecision: baseReceipt.Decision.Value + ":" + baseReceipt.Decision.Resolution, InterventionDecision: changedReceipt.Decision.Value + ":" + changedReceipt.Decision.Resolution, BaselineStates: baseStates, InterventionStates: changedStates, BaselineTransitionDigests: baseTransitions, InterventionTransitionDigests: changedTransitions, BaselineCausePath: baseCause, InterventionCausePath: changedCause, SemanticDigestChanged: baseReceipt.Subject.SemanticDigest != changedReceipt.Subject.SemanticDigest, GraphDigestChanged: baseReceipt.Graph.Digest != changedReceipt.Graph.Digest, EvidenceDigestChanged: baseReceipt.EvidenceDigest != changedReceipt.EvidenceDigest, StateTransitionChanged: !reflect.DeepEqual(baseStates, changedStates) || !reflect.DeepEqual(baseTransitions, changedTransitions), CausePathChanged: !reflect.DeepEqual(baseCause, changedCause), DecisionChanged: baseReceipt.Decision != changedReceipt.Decision, EdgeTypeChanged: !reflect.DeepEqual(baseReceipt.Graph.Edges, changedReceipt.Graph.Edges), BaselineReadOnly: baseReceipt.Subject.ReadOnly, InterventionReadOnly: changedReceipt.Subject.ReadOnly, BaselineRepositoryWrites: baseReceipt.Subject.RepositoryWrites, InterventionRepositoryWrites: changedReceipt.Subject.RepositoryWrites, BaselineAuthorityResolution: baseReceipt.Subject.AuthorityResolution, InterventionAuthorityResolution: changedReceipt.Subject.AuthorityResolution}
}

func evidence(artifact, operation, repoRoot, capability, observationDir string) claimdependency.EvidenceReceipt {
	if operation == "same" {
		fail("internal intervention operation cannot be same")
	}
	observationPath := ""
	if operation != "availability" {
		observationPath = filepath.Join(observationDir, "refuted.json")
		if operation == "acceptance" {
			observationPath = filepath.Join(observationDir, "accepted.json")
			if filepath.Base(artifact) == "value-intervention.gooo" {
				observationPath = filepath.Join(observationDir, "semantic-accepted.json")
			}
		}
	}
	return evidenceWithObservation(artifact, operation, repoRoot, capability, observationPath)
}

func evidenceWithObservation(artifact, operation, repoRoot, capability, observationPath string) claimdependency.EvidenceReceipt {
	output := filepath.Join(os.TempDir(), "gooo-claim-dependency-evidence-"+strings.ReplaceAll(filepath.Base(artifact), ".", "-")+"-"+operation+".json")
	sourcePath, targetPath := artifact, artifact
	if filepath.Base(artifact) == "main.gooo" || filepath.Base(artifact) == "unknown.gooo" {
		targetPath = filepath.Join(filepath.Dir(artifact), "accepted-target.gooo")
	}
	if filepath.Base(artifact) == "refuted.gooo" {
		targetPath = filepath.Join(filepath.Dir(artifact), "refuted-target.gooo")
	}
	if filepath.Base(artifact) == "value-intervention.gooo" {
		sourcePath = filepath.Join(filepath.Dir(artifact), "main.gooo")
	}
	receipt, err := claimdependency.BuildCurrentEvidenceForSource(sourcePath, targetPath, operation, capability, repoRoot, output, observationPath)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(output, receipt)
	return receipt
}
func states(receipt claimdependency.Receipt) []string {
	result := make([]string, len(receipt.Resolutions))
	for i, value := range receipt.Resolutions {
		result[i] = value.State
	}
	return result
}
func transitionDigests(receipt claimdependency.Receipt) []string {
	result := []string{}
	for _, value := range receipt.Transitions {
		if value.Event != "CLAIM_REGISTERED" {
			result = append(result, value.TransitionDigest)
		}
	}
	return result
}
func lastCause(receipt claimdependency.Receipt) []string {
	if len(receipt.Resolutions) == 0 {
		return nil
	}
	return receipt.Resolutions[len(receipt.Resolutions)-1].CausePath
}
func read(path string) []byte {
	value, err := os.ReadFile(path)
	if err != nil {
		fail(err.Error())
	}
	return value
}
func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err.Error())
	}
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(2) }
