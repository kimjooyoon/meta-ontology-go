package languageresourcebudgetconsumer

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema   = "gooo/meta-resource-budget-consumer-report/v1"
	Producer = "scripts/meta-resource-budget"
	Consumer = "cmd/meta-resource-budget-consumer"
)

type sourceReceipt struct {
	SchemaVersion string            `json:"schema_version"`
	Command       string            `json:"command"`
	Status        string            `json:"status"`
	File          string            `json:"file"`
	Diagnostics   []json.RawMessage `json:"diagnostics"`
}

type artifact struct {
	Schema        string `json:"schema"`
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	Reason        string `json:"reason"`
	Kind          string `json:"kind"`
	SubjectDigest string `json:"subject_digest"`
	Package       struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"package"`
	Operation struct {
		Activity string    `json:"activity"`
		Inputs   []Binding `json:"inputs"`
		Output   Binding   `json:"output"`
	} `json:"operation"`
	Effects Effects `json:"effects"`
	Digest  string  `json:"digest"`
}

func Consume(input Input, label string) Report {
	report := Report{Schema: Schema, Label: label, EvidenceClass: input.EvidenceClass, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", WriteSets: input.Producer.WriteSets}
	contractOK := validInputContract(input)
	meaning, sourceErr := reconstructSource(input)
	semanticDecision, semanticResolution, semanticState, semanticReason := "PASS", "EXACT", "DISCHARGED", "SEMANTIC_SOURCE_LOWERING_AND_ARTIFACT_REPLAY_STABLE"
	artifactMeaning := ArtifactMeaning{}
	if !contractOK {
		semanticDecision, semanticResolution, semanticState, semanticReason = "FAIL_CLOSED", "EXACT", "REFUTED", "CONTRACT_OR_SUBJECT_INVALID"
	} else if sourceErr != nil {
		semanticDecision, semanticResolution, semanticState, semanticReason = semanticFailure(sourceErr.Error())
	} else if input.Producer.SourceDigest != meaning.SourceDigest {
		semanticDecision, semanticResolution, semanticState, semanticReason = "FAIL_CLOSED", "EXACT", "REFUTED", "SOURCE_DIGEST_MISMATCH"
	} else {
		report.Source = meaning
		artifactMeaning, sourceErr = verifyOutputs(input, meaning)
		if sourceErr != nil {
			semanticDecision, semanticResolution, semanticState, semanticReason = semanticFailure(sourceErr.Error())
		}
	}
	report.SemanticDecision, report.SemanticResolution, report.SemanticClaimState, report.SemanticReason = semanticDecision, semanticResolution, semanticState, semanticReason
	report.Artifact = artifactMeaning
	report.Resource = resourceEnvelope(input, meaning)
	report.Imports = input.Producer.ImportScan
	report.Provenance = Provenance{RawSourceFiles: len(input.Producer.SourceFiles), RawOperationOutputs: len(input.Producer.RawOutputs), RawResourceSamples: len(input.Observations), RawEvidenceDigest: rawEvidenceDigest(input), ConsumerPackage: "internal/meta/languageresourcebudgetconsumer"}
	writeTo, writeReason := writeSetState(input.Producer.WriteSets, input.Producer.Effects, input.Contract)
	report.ClaimTransitions = transitions(semanticState, report.Resource, writeTo, writeReason)
	report.Decision, report.Resolution, report.Reason = overallDecision(semanticState, report.Resource, report.ClaimTransitions)
	if !validImportScan(input.Producer.ImportScan) {
		report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "LOWER_RESOLUTION", "IMPORT_SCAN_INVALID"
	}
	report.FactsDigest = digestValue(struct {
		Source      SourceMeaning
		Artifact    ArtifactMeaning
		Resource    ResourceEnvelope
		WriteSets   []WriteSetObservation
		Imports     ImportScan
		RawEvidence string
		Transitions []ClaimTransition
	}{report.Source, report.Artifact, report.Resource, report.WriteSets, report.Imports, report.Provenance.RawEvidenceDigest, report.ClaimTransitions})
	report.Provenance.EvidenceDigest = report.FactsDigest
	previous := ""
	for index := range report.ClaimTransitions {
		report.ClaimTransitions[index].Evidence = report.FactsDigest
		report.ClaimTransitions[index].PreviousDigest = previous
		report.ClaimTransitions[index] = sealTransition(report.ClaimTransitions[index])
		previous = report.ClaimTransitions[index].Digest
	}
	report.Digest = digestValue(report)
	return report
}

func semanticFailure(reason string) (string, string, string, string) {
	if strings.Contains(reason, "MISSING") {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "OPEN", reason
	}
	return "FAIL_CLOSED", "EXACT", "REFUTED", reason
}

func validInputContract(input Input) bool {
	if input.Schema != "gooo/meta-resource-budget-input/v1" || !positiveSHA(input.ExpectedHead) || input.ContractDigest != digestValue(input.Contract) {
		return false
	}
	contract := input.Contract
	if contract.Schema != "gooo/meta-resource-budget-contract/v1" || contract.ID != "meta-resource-budget-v1" || contract.SamplesPerOp != 3 || contract.Indicators != 19 || len(contract.SourcePaths) != 2 || contract.SourcePaths[0] != "examples/meta-resource-budget/activity.gooo" || contract.SourcePaths[1] != "examples/meta-resource-budget/entities.gooo" || len(contract.Operations) != 3 || contract.Limits != (Limits{WallTimeMS: 2000, PeakRSSKiB: 131072, ReceiptBytes: 8192, GeneratedBytes: 16384}) || len(contract.NotClaimed) != 5 || len(contract.References) != 2 || contract.References[0] != (Reference{ID: "github-hosted-runners", URL: "https://docs.github.com/en/actions/reference/runners/github-hosted-runners"}) || contract.References[1] != (Reference{ID: "bazel-hermeticity", URL: "https://bazel.build/basics/hermeticity"}) {
		return false
	}
	expectedNotClaimed := []string{"cross-run performance improvement", "machine-independent resource bounds", "business correctness", "production readiness", "hermetic build guarantee"}
	for index := range expectedNotClaimed {
		if contract.NotClaimed[index] != expectedNotClaimed[index] {
			return false
		}
	}
	expected := map[string]Operation{
		"source-check":     {ID: "source-check", Stage: "LOWER", Step: "parse-source", MetaOperation: "observe-source-receipt", ProofChoice: "FOUNDATION", Output: "RECEIPT"},
		"project-manifest": {ID: "project-manifest", Stage: "PROJECT", Step: "project-operation-manifest", MetaOperation: "project-generated-operation", ProofChoice: "COHERENCE", Output: "GENERATED"},
		"replay-manifest":  {ID: "replay-manifest", Stage: "REPLAY", Step: "replay-generated-artifact", MetaOperation: "prove-generated-replay", ProofChoice: "REGRESSION", Output: "GENERATED"},
	}
	seen := map[string]bool{}
	for _, operation := range contract.Operations {
		want, ok := expected[operation.ID]
		if !ok || want != operation || seen[operation.ID] {
			return false
		}
		seen[operation.ID] = true
	}
	return len(seen) == len(expected)
}

func reconstructSource(input Input) (SourceMeaning, error) {
	if len(input.Producer.SourceFiles) < input.Producer.SourceFileCount || len(input.Producer.SourceFiles) < len(input.Contract.SourcePaths) {
		return SourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_MISSING")
	}
	if len(input.Producer.SourceFiles) != input.Producer.SourceFileCount || len(input.Producer.SourceFiles) != len(input.Contract.SourcePaths) {
		return SourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
	}
	sources := append([]RawSource(nil), input.Producer.SourceFiles...)
	sort.Slice(sources, func(left, right int) bool { return sources[left].Filename < sources[right].Filename })
	var declarations []syntax.Declaration
	var packageName, namespace string
	for index, raw := range sources {
		if raw.Filename == "" || !strings.HasSuffix(raw.Filename, ".gooo") {
			return SourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
		}
		content, err := decode(raw.ContentBase64)
		if err != nil {
			return SourceMeaning{}, fmt.Errorf("SOURCE_PAYLOAD_INVALID")
		}
		if len(content) == 0 {
			return SourceMeaning{}, fmt.Errorf("SOURCE_PAYLOAD_MISSING")
		}
		file, diagnostics := syntax.ParseFile(raw.Filename, string(content))
		if file == nil || diagnostics.HasErrors() || file.Package == nil || file.Namespace == nil {
			return SourceMeaning{}, fmt.Errorf("SOURCE_SYNTAX_INVALID")
		}
		if index == 0 {
			packageName, namespace = file.Package.Name, file.Namespace.Name
		} else if file.Package.Name != packageName || file.Namespace.Name != namespace {
			return SourceMeaning{}, fmt.Errorf("SOURCE_HEADER_MISMATCH")
		}
		declarations = append(declarations, file.Decls...)
	}
	combined := &syntax.File{Package: &syntax.PackageDecl{Name: packageName}, Namespace: &syntax.NamespaceDecl{Name: namespace}, Decls: declarations, Declarations: declarations}
	if _, err := syntax.Format(combined); err != nil {
		return SourceMeaning{}, fmt.Errorf("SOURCE_CANONICAL_FORMAT_INVALID")
	}
	ir, err := bidir.Lower(combined)
	if err != nil {
		return SourceMeaning{}, fmt.Errorf("SOURCE_SEMANTIC_LOWERING_INVALID")
	}
	if err := ir.Validate(); err != nil {
		return SourceMeaning{}, fmt.Errorf("SOURCE_SEMANTIC_VALIDATION_INVALID")
	}
	activities := make([]*syntax.ActivityDecl, 0, 1)
	entities := make(map[string]Binding)
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.ActivityDecl:
			activities = append(activities, value)
		case *syntax.EntityDecl:
			entities[value.Name] = Binding{Name: value.Name, ID: value.ID}
		}
	}
	if len(activities) != 1 {
		return SourceMeaning{}, fmt.Errorf("SOURCE_ENTRY_CARDINALITY_INVALID")
	}
	activity := activities[0]
	parameters := activity.Inputs
	if parameters == nil {
		parameters = activity.Parameters
	}
	inputs := make([]Binding, 0, len(parameters))
	for _, parameter := range parameters {
		value, ok := entities[parameter.Name]
		if !ok {
			return SourceMeaning{}, fmt.Errorf("SOURCE_INPUT_ENTITY_UNKNOWN")
		}
		inputs = append(inputs, value)
	}
	outputName := activity.Output
	if outputName == "" {
		outputName = activity.Result.Name
	}
	output, ok := entities[outputName]
	if !ok {
		return SourceMeaning{}, fmt.Errorf("SOURCE_OUTPUT_ENTITY_UNKNOWN")
	}
	meaning := SourceMeaning{Package: packageName, Namespace: namespace, Activity: activity.Name, Inputs: inputs, Output: Binding{Name: outputName, ID: output.ID}, SourceDigest: sourceSetDigest(sources), SemanticDigest: "sha256:" + ir.StableHash()}
	meaning.TargetDigest = targetDigest(meaning)
	return meaning, nil
}

func verifyOutputs(input Input, meaning SourceMeaning) (ArtifactMeaning, error) {
	if input.Producer.SourceReceiptBase64 == "" {
		return ArtifactMeaning{}, fmt.Errorf("SOURCE_RECEIPT_PAYLOAD_MISSING")
	}
	sourcePayload, err := decode(input.Producer.SourceReceiptBase64)
	if err != nil {
		return ArtifactMeaning{}, fmt.Errorf("SOURCE_RECEIPT_PAYLOAD_MISSING")
	}
	var source sourceReceipt
	if err := json.Unmarshal(sourcePayload, &source); err != nil {
		return ArtifactMeaning{}, fmt.Errorf("SOURCE_RECEIPT_INVALID")
	}
	if source.SchemaVersion != "gooo/diagnostics/v1" || source.Command != "check" || source.Status != "ok" || source.File != firstSourceFilename(input.Producer.SourceFiles) || len(source.Diagnostics) != 0 {
		return ArtifactMeaning{}, fmt.Errorf("SOURCE_RECEIPT_NOT_EXACT")
	}
	if input.Producer.ArtifactBase64 == "" {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_PAYLOAD_MISSING")
	}
	artifactPayload, err := decode(input.Producer.ArtifactBase64)
	if err != nil {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_PAYLOAD_MISSING")
	}
	if input.Producer.ReplayBase64 == "" {
		return ArtifactMeaning{}, fmt.Errorf("REPLAY_PAYLOAD_MISSING")
	}
	replayPayload, err := decode(input.Producer.ReplayBase64)
	if err != nil {
		return ArtifactMeaning{}, fmt.Errorf("REPLAY_PAYLOAD_MISSING")
	}
	var first, replay artifact
	if err := json.Unmarshal(artifactPayload, &first); err != nil {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_INVALID")
	}
	if err := json.Unmarshal(replayPayload, &replay); err != nil {
		return ArtifactMeaning{}, fmt.Errorf("REPLAY_INVALID")
	}
	valid := func(value artifact) bool {
		return value.Schema == "gooo/operation-manifest/v1" && value.Decision == "PASS" && value.Resolution == "EXACT" && value.Reason == "OPERATION_MANIFEST_EMITTED" && value.Kind == "operation-manifest" && value.Operation.Activity == meaning.Activity && bindingsEqual(value.Operation.Inputs, meaning.Inputs) && value.Operation.Output == meaning.Output && value.Package.Name == meaning.Package && value.Package.Namespace == meaning.Namespace && value.Effects.RepositoryWrites == 0 && !value.Effects.MutationAuthority && contentDigest(value.SubjectDigest) && contentDigest(value.Digest)
	}
	if !valid(first) || !valid(replay) {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_SOURCE_MEANING_MISMATCH")
	}
	if !bytes.Equal(artifactPayload, replayPayload) {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_REPLAY_MISMATCH")
	}
	if err := verifyRawOutputs(input, meaning, sourcePayload, artifactPayload, replayPayload); err != nil {
		return ArtifactMeaning{}, err
	}
	return ArtifactMeaning{Activity: first.Operation.Activity, Inputs: first.Operation.Inputs, Output: first.Operation.Output, Decision: first.Decision, Reason: first.Reason, SourceReceiptDigest: digestBytes(sourcePayload), ArtifactDigest: digestBytes(artifactPayload), ReplayDigest: digestBytes(replayPayload)}, nil
}

func verifyRawOutputs(input Input, meaning SourceMeaning, sourcePayload, artifactPayload, replayPayload []byte) error {
	expected := len(input.Contract.Operations)
	if len(input.Producer.RawOutputs) < expected {
		return fmt.Errorf("RAW_OPERATION_OUTPUTS_MISSING")
	}
	byKey := make(map[string]RawOutput, len(input.Producer.RawOutputs))
	for _, raw := range input.Producer.RawOutputs {
		if raw.Operation == "" || raw.Sequence != 1 || raw.PayloadBase64 == "" {
			return fmt.Errorf("RAW_OPERATION_OUTPUT_MISSING")
		}
		key := outputKey(raw.Operation, raw.Sequence)
		if _, exists := byKey[key]; exists {
			return fmt.Errorf("RAW_OPERATION_OUTPUT_DUPLICATE")
		}
		payload, err := decode(raw.PayloadBase64)
		if err != nil {
			return fmt.Errorf("RAW_OPERATION_OUTPUT_INVALID")
		}
		spec, ok := operation(raw.Operation, input.Contract)
		if !ok || raw.Kind != spec.Output {
			return fmt.Errorf("RAW_OPERATION_OUTPUT_TARGET_MISMATCH")
		}
		byKey[key] = raw
		if !outputObservationsMatch(input.Observations, raw.Operation, digestBytes(payload), meaning) {
			return fmt.Errorf("PRODUCER_OUTPUT_BINDING_MISMATCH")
		}
	}
	for _, spec := range input.Contract.Operations {
		if _, ok := byKey[outputKey(spec.ID, 1)]; !ok {
			return fmt.Errorf("RAW_OPERATION_OUTPUT_MISSING")
		}
	}
	if !bytes.Equal(firstOutput(byKey, "source-check"), sourcePayload) || !bytes.Equal(firstOutput(byKey, "project-manifest"), artifactPayload) || !bytes.Equal(firstOutput(byKey, "replay-manifest"), replayPayload) {
		return fmt.Errorf("PRODUCER_OUTPUT_PAYLOAD_MISMATCH")
	}
	return nil
}

func resourceEnvelope(input Input, meaning SourceMeaning) ResourceEnvelope {
	result := ResourceEnvelope{Decision: "PASS", Resolution: "EXACT", Operations: len(input.Contract.Operations), Samples: len(input.Observations), ExpectedSamples: len(input.Contract.Operations) * input.Contract.SamplesPerOp, Runner: input.Producer.Runner, PerOperation: []ResourceOperation{}}
	byOperation := make(map[string][]Observation, len(input.Contract.Operations))
	for _, value := range input.Observations {
		byOperation[value.Operation] = append(byOperation[value.Operation], value)
	}
	for _, spec := range input.Contract.Operations {
		values := append([]Observation(nil), byOperation[spec.ID]...)
		sort.SliceStable(values, func(left, right int) bool { return values[left].Sequence < values[right].Sequence })
		summary := ResourceOperation{Operation: spec.ID, Samples: len(values), MetricStatus: map[string]string{"wall-time": "SATISFIED", "peak-rss": "SATISFIED", "receipt-bytes": "NOT_APPLICABLE", "generated-bytes": "NOT_APPLICABLE"}}
		if len(values) < input.Contract.SamplesPerOp {
			summary.MissingSamples = input.Contract.SamplesPerOp - len(values)
			result.Decision, result.Resolution = "FAIL_CLOSED", "LOWER_RESOLUTION"
		}
		if len(values) > input.Contract.SamplesPerOp {
			summary.InvalidSamples = len(values) - input.Contract.SamplesPerOp
			result.Decision, result.Resolution = "FAIL_CLOSED", "LOWER_RESOLUTION"
		}
		if spec.Output == "RECEIPT" {
			summary.MetricStatus["receipt-bytes"] = "SATISFIED"
		} else {
			summary.MetricStatus["generated-bytes"] = "SATISFIED"
		}
		for _, value := range values {
			valid := validObservation(value, spec, input, meaning)
			if value.Sequence < 1 || value.Sequence > input.Contract.SamplesPerOp || !valid {
				summary.InvalidSamples++
				result.Decision, result.Resolution = "FAIL_CLOSED", "LOWER_RESOLUTION"
				continue
			}
			if value.WallTimeNS > summary.WallMaxNS {
				summary.WallMaxNS = value.WallTimeNS
			}
			if value.PeakRSSKiB > summary.PeakRSSMaxKiB {
				summary.PeakRSSMaxKiB = value.PeakRSSKiB
			}
			if value.ReceiptBytes > summary.ReceiptMaxBytes {
				summary.ReceiptMaxBytes = value.ReceiptBytes
			}
			if value.GeneratedBytes > summary.GeneratedMaxBytes {
				summary.GeneratedMaxBytes = value.GeneratedBytes
			}
			if value.WallTimeNS > input.Contract.Limits.WallTimeMS*1000000 {
				summary.BudgetViolations++
				summary.MetricStatus["wall-time"] = "REFUTED"
			}
			if value.PeakRSSKiB > input.Contract.Limits.PeakRSSKiB {
				summary.BudgetViolations++
				summary.MetricStatus["peak-rss"] = "REFUTED"
			}
			if value.ReceiptBytes > input.Contract.Limits.ReceiptBytes {
				summary.BudgetViolations++
				summary.MetricStatus["receipt-bytes"] = "REFUTED"
			}
			if value.GeneratedBytes > input.Contract.Limits.GeneratedBytes {
				summary.BudgetViolations++
				summary.MetricStatus["generated-bytes"] = "REFUTED"
			}
		}
		if summary.MissingSamples > 0 || summary.InvalidSamples > 0 {
			for metric, status := range summary.MetricStatus {
				if status != "NOT_APPLICABLE" {
					summary.MetricStatus[metric] = "UNKNOWN"
				}
			}
		}
		result.PerOperation = append(result.PerOperation, summary)
	}
	if result.Resolution == "EXACT" {
		for _, summary := range result.PerOperation {
			if summary.BudgetViolations > 0 {
				result.Decision = "FAIL_CLOSED"
			}
		}
	}
	return result
}

func validObservation(value Observation, spec Operation, input Input, meaning SourceMeaning) bool {
	return value.Schema == "gooo/meta-resource-budget-observation/v1" && value.SubjectSHA == input.ExpectedHead && value.Producer == Producer && value.Consumer == Consumer && value.Operation == spec.ID && value.Stage == spec.Stage && value.Step == spec.Step && value.MetaOperation == spec.MetaOperation && value.ProofChoice == spec.ProofChoice && value.Reason == "RUNNER_RESOURCE_OBSERVED" && value.ExitCode == 0 && value.WallTimeNS > 0 && value.PeakRSSKiB > 0 && value.ReceiptBytes >= 0 && value.GeneratedBytes >= 0 && contentDigest(value.OutputDigest) && value.SourceRawDigest == meaning.SourceDigest && value.SourceSemanticDigest == meaning.SemanticDigest && value.EntryDigest == meaning.TargetDigest && value.TargetDigest == meaning.TargetDigest
}

func transitions(semanticState string, resource ResourceEnvelope, writeTo, writeReason string) []ClaimTransition {
	semanticReason := "SEMANTIC_SOURCE_LOWERING_AND_ARTIFACT_REPLAY_STABLE"
	if semanticState != "DISCHARGED" {
		semanticReason = "SEMANTIC_EVIDENCE_" + semanticState
	}
	resourceTo, resourceReason := "DISCHARGED", "RESOURCE_ENVELOPE_OBSERVED"
	if resource.Resolution != "EXACT" {
		resourceTo, resourceReason = "OPEN", resourceReasonFor(resource)
	} else if resource.Decision != "PASS" {
		resourceTo, resourceReason = "REFUTED", "RESOURCE_BUDGET_EXCEEDED"
	}
	return []ClaimTransition{{Sequence: 1, ClaimID: "semantic-meaning", From: "OPEN", To: semanticState, Stage: "CONSUME", Step: "semantic-verdict", Reason: semanticReason}, {Sequence: 2, ClaimID: "runner-resource-envelope", From: "OPEN", To: resourceTo, Stage: "CONSUME", Step: "resource-verdict", Reason: resourceReason}, {Sequence: 3, ClaimID: "net-repository-state", From: "OPEN", To: writeTo, Stage: "CONSUME", Step: "effect-verdict", Reason: writeReason}}
}

func resourceReasonFor(resource ResourceEnvelope) string {
	for _, value := range resource.PerOperation {
		if value.MissingSamples > 0 {
			return "RESOURCE_SAMPLE_MISSING"
		}
	}
	return "RESOURCE_SAMPLE_INVALID"
}

func writeSetState(values []WriteSetObservation, effects Effects, contract Contract) (string, string) {
	if effects.RepositoryWrites != 0 || effects.MutationAuthority {
		return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
	}
	if len(values) != len(contract.Operations) {
		return "OPEN", "EFFECT_OBSERVATION_MISSING"
	}
	seen := map[string]bool{}
	open := false
	for _, value := range values {
		if seen[value.Operation] {
			return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
		}
		seen[value.Operation] = true
		if value.RepositoryWrites != 0 || value.MutationAuthority || value.DiffExitCode != 0 || len(value.ChangedPaths) != 0 || value.UntrackedFileCount != 0 || value.BeforeTreeDigest != value.AfterTreeDigest {
			return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
		}
		if value.Schema != "gooo/meta-resource-budget-write-set/v1" || value.Producer != Producer || value.Consumer != Consumer || !gitDigest(value.BeforeTreeDigest) || !gitDigest(value.AfterTreeDigest) || !contentDigest(value.WriteSetDigest) || !value.AuthorityObserved || !value.BeforeStatusObserved || !value.AfterStatusObserved || value.SampleStart != 1 || value.SampleEnd != contract.SamplesPerOp || value.Reason != "NET_REPOSITORY_STATE_UNCHANGED_ACROSS_OPERATION_WINDOW" {
			open = true
			continue
		}
		before, beforeErr := decode(value.BeforeStatusBase64)
		after, afterErr := decode(value.AfterStatusBase64)
		if beforeErr != nil || afterErr != nil || value.WriteSetDigest != statusDigest(before, after) {
			open = true
			continue
		}
		if !bytes.Equal(before, after) {
			return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
		}
	}
	for _, spec := range contract.Operations {
		if !seen[spec.ID] {
			open = true
		}
	}
	if open {
		return "OPEN", "EFFECT_OBSERVATION_MISSING"
	}
	return "DISCHARGED", "NET_REPOSITORY_STATE_UNCHANGED"
}

func overallDecision(semantic string, resource ResourceEnvelope, transitions []ClaimTransition) (string, string, string) {
	if transitions[2].To == "REFUTED" {
		return "FAIL_CLOSED", "EXACT", "EFFECT_BOUNDARY_VIOLATED"
	}
	if transitions[2].To == "OPEN" {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "EFFECT_OBSERVATION_MISSING"
	}
	if semantic == "OPEN" {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "SEMANTIC_EVIDENCE_MISSING"
	}
	if semantic != "DISCHARGED" {
		return "FAIL_CLOSED", "EXACT", "SEMANTIC_EVIDENCE_INVALID"
	}
	if resource.Resolution != "EXACT" {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "RESOURCE_SAMPLE_MISSING_OR_INVALID"
	}
	if resource.Decision != "PASS" {
		return "FAIL_CLOSED", "EXACT", "RESOURCE_BUDGET_EXCEEDED"
	}
	return "PASS", "EXACT", "RESOURCE_ENVELOPE_OBSERVED"
}

func outputKey(operation string, sequence int) string {
	return fmt.Sprintf("%s#%d", operation, sequence)
}

func firstOutput(outputs map[string]RawOutput, operation string) []byte {
	value, ok := outputs[outputKey(operation, 1)]
	if !ok {
		return nil
	}
	payload, _ := decode(value.PayloadBase64)
	return payload
}
func outputDigestForObservation(observations []Observation, operation string, sequence int, digest string) bool {
	count := 0
	for _, value := range observations {
		if value.Operation == operation && value.Sequence == sequence {
			count++
			if value.OutputDigest != digest {
				return false
			}
		}
	}
	return count == 1
}
func outputObservationsMatch(observations []Observation, operation, digest string, meaning SourceMeaning) bool {
	count := 0
	for _, value := range observations {
		if value.Operation == operation {
			count++
			if value.OutputDigest != digest || value.SourceRawDigest != meaning.SourceDigest || value.SourceSemanticDigest != meaning.SemanticDigest || value.EntryDigest != meaning.TargetDigest || value.TargetDigest != meaning.TargetDigest {
				return false
			}
		}
	}
	return count > 0
}

func sourceSetDigest(sources []RawSource) string {
	values := append([]RawSource(nil), sources...)
	sort.Slice(values, func(left, right int) bool { return values[left].Filename < values[right].Filename })
	var payload bytes.Buffer
	for _, raw := range values {
		content, _ := decode(raw.ContentBase64)
		payload.WriteString(raw.Filename)
		payload.WriteByte(0)
		payload.Write(content)
		payload.WriteByte(0)
	}
	return digestBytes(payload.Bytes())
}
func targetDigest(meaning SourceMeaning) string {
	return digestValue(struct {
		Package   string
		Namespace string
		Activity  string
		Inputs    []Binding
		Output    Binding
	}{meaning.Package, meaning.Namespace, meaning.Activity, meaning.Inputs, meaning.Output})
}
func firstSourceFilename(sources []RawSource) string {
	values := append([]RawSource(nil), sources...)
	sort.Slice(values, func(left, right int) bool { return values[left].Filename < values[right].Filename })
	if len(values) == 0 {
		return ""
	}
	return values[0].Filename
}
func bindingsEqual(left, right []Binding) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func decode(value string) ([]byte, error) { return base64.StdEncoding.DecodeString(value) }
func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func digestValue(value any) string { data, _ := json.Marshal(value); return digestBytes(data) }
func sealTransition(value ClaimTransition) ClaimTransition {
	value.Digest = ""
	value.Digest = digestValue(value)
	return value
}
func positiveSHA(value string) bool { return len(value) == 40 && isHex(value) }
func gitDigest(value string) bool   { return len(value) == 40 && isHex(value) }
func contentDigest(value string) bool {
	return len(value) == 71 && strings.HasPrefix(value, "sha256:") && isHex(value[7:])
}
func isHex(value string) bool {
	for _, char := range value {
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') && !(char >= 'A' && char <= 'F') {
			return false
		}
	}
	return true
}
func validRunner(value Runner) bool {
	return value.OS != "" && value.Architecture != "" && value.Image != "" && value.ImageVersion != "" && value.GoVersion == "go1.27.0"
}
func statusDigest(before, after []byte) string {
	value := append(append([]byte(nil), before...), 0)
	value = append(value, after...)
	return digestBytes(value)
}
func validImportScan(value ImportScan) bool {
	return value.Schema == "gooo/meta-resource-budget-import-scan/v1" && value.Denominator == 2 && value.Numerator >= 0 && value.Numerator <= value.Denominator && value.Numerator == boolInt(!value.ConsumerPackageReducerImported)+boolInt(!value.ConsumerCommandReducerImported) && value.ConsumerPackageFilesScanned > 0 && value.ConsumerCommandFilesScanned > 0 && value.Digest == digestValue(struct {
		Schema          string `json:"schema"`
		PackageImported bool   `json:"consumer_package_reducer_imported"`
		CommandImported bool   `json:"consumer_command_reducer_imported"`
		PackageFiles    int    `json:"consumer_package_files_scanned"`
		CommandFiles    int    `json:"consumer_command_files_scanned"`
		Numerator       int    `json:"numerator"`
		Denominator     int    `json:"denominator"`
	}{value.Schema, value.ConsumerPackageReducerImported, value.ConsumerCommandReducerImported, value.ConsumerPackageFilesScanned, value.ConsumerCommandFilesScanned, value.Numerator, value.Denominator})
}

func rawEvidenceDigest(input Input) string {
	return digestValue(struct {
		Sources      []RawSource
		Outputs      []RawOutput
		Observations []Observation
		WriteSets    []WriteSetObservation
	}{input.Producer.SourceFiles, input.Producer.RawOutputs, input.Observations, input.Producer.WriteSets})
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
