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
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
	Kind       string `json:"kind"`
	Package    struct {
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
	report := Report{Schema: Schema, Label: label, EvidenceClass: input.EvidenceClass, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", WriteSet: input.Producer.WriteSet}
	meaning, err := reconstructSource(input)
	semanticDecision, semanticResolution, semanticReason := "PASS", "EXACT", "SEMANTIC_SOURCE_LOWERING_AND_ARTIFACT_REPLAY_STABLE"
	artifactMeaning := ArtifactMeaning{}
	if err != nil {
		semanticDecision, semanticReason = "FAIL_CLOSED", err.Error()
	} else if input.Producer.SourceDigest != meaning.SourceDigest {
		semanticDecision, semanticReason = "FAIL_CLOSED", "SOURCE_DIGEST_MISMATCH"
	} else {
		report.Source = meaning
		artifactMeaning, err = verifyOutputs(input, meaning)
		if err != nil {
			semanticDecision, semanticReason = "FAIL_CLOSED", err.Error()
		}
	}
	report.SemanticDecision, report.SemanticResolution = semanticDecision, semanticResolution
	report.SemanticReason = semanticReason
	report.Artifact = artifactMeaning
	report.Resource = resourceEnvelope(input)
	report.Imports = ImportBoundary{Independent: true, Numerator: 2, Denominator: 2}
	report.Provenance = Provenance{RawSourceFiles: len(input.Producer.SourceFiles), RawOperationOutputs: 3, RawResourceSamples: len(input.Observations), ConsumerPackage: "internal/meta/languageresourcebudgetconsumer"}
	report.ClaimTransitions = transitions(semanticDecision == "PASS", report.Resource, input.Producer.WriteSet, input.Producer.Effects)
	report.Decision, report.Resolution, report.Reason = overallDecision(semanticDecision, report.Resource, report.ClaimTransitions)
	report.FactsDigest = digestValue(struct {
		Source      SourceMeaning
		Artifact    ArtifactMeaning
		Resource    ResourceEnvelope
		WriteSet    WriteSetObservation
		Transitions []ClaimTransition
	}{report.Source, report.Artifact, report.Resource, report.WriteSet, report.ClaimTransitions})
	report.Provenance.EvidenceDigest = report.FactsDigest
	for index := range report.ClaimTransitions {
		report.ClaimTransitions[index].Evidence = report.FactsDigest
		report.ClaimTransitions[index].PreviousDigest = ""
		if index > 0 {
			report.ClaimTransitions[index].PreviousDigest = report.ClaimTransitions[index-1].Digest
		}
		report.ClaimTransitions[index] = sealTransition(report.ClaimTransitions[index])
	}
	report.Digest = digestValue(report)
	return report
}

func reconstructSource(input Input) (SourceMeaning, error) {
	if len(input.Producer.SourceFiles) != input.Producer.SourceFileCount || len(input.Producer.SourceFiles) != len(input.Contract.SourcePaths) {
		return SourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
	}
	sources := append([]RawSource(nil), input.Producer.SourceFiles...)
	sort.Slice(sources, func(left, right int) bool { return sources[left].Filename < sources[right].Filename })
	var declarations []syntax.Declaration
	var packageName, namespace string
	for index, raw := range sources {
		if raw.Filename == "" {
			return SourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
		}
		content, err := base64.StdEncoding.DecodeString(raw.ContentBase64)
		if err != nil || len(content) == 0 {
			return SourceMeaning{}, fmt.Errorf("SOURCE_PAYLOAD_INVALID")
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
	return SourceMeaning{Package: packageName, Namespace: namespace, Activity: activity.Name, Inputs: inputs, Output: output,
		SourceDigest: sourceSetDigest(sources), SemanticDigest: "sha256:" + ir.StableHash()}, nil
}

func verifyOutputs(input Input, meaning SourceMeaning) (ArtifactMeaning, error) {
	sourcePayload, err := decode(input.Producer.SourceReceiptBase64)
	if err != nil {
		return ArtifactMeaning{}, fmt.Errorf("SOURCE_RECEIPT_INVALID")
	}
	var source sourceReceipt
	if err := json.Unmarshal(sourcePayload, &source); err != nil || source.SchemaVersion != "gooo/diagnostics/v1" || source.Command != "check" || source.Status != "ok" || source.File != firstSourceFilename(input.Producer.SourceFiles) || len(source.Diagnostics) != 0 {
		return ArtifactMeaning{}, fmt.Errorf("SOURCE_RECEIPT_NOT_EXACT")
	}
	artifactPayload, err := decode(input.Producer.ArtifactBase64)
	if err != nil {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_INVALID")
	}
	replayPayload, err := decode(input.Producer.ReplayBase64)
	if err != nil {
		return ArtifactMeaning{}, fmt.Errorf("REPLAY_INVALID")
	}
	var first, replay artifact
	if err := json.Unmarshal(artifactPayload, &first); err != nil {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_INVALID")
	}
	if err := json.Unmarshal(replayPayload, &replay); err != nil {
		return ArtifactMeaning{}, fmt.Errorf("REPLAY_INVALID")
	}
	valid := func(value artifact) bool {
		return value.Schema == "gooo/operation-manifest/v1" && value.Decision == "PASS" && value.Resolution == "EXACT" && value.Reason == "OPERATION_MANIFEST_EMITTED" && value.Kind == "operation-manifest" && value.Operation.Activity == meaning.Activity && bindingsEqual(value.Operation.Inputs, meaning.Inputs) && value.Operation.Output == meaning.Output && value.Package.Name == meaning.Package && value.Package.Namespace == meaning.Namespace && value.Effects.RepositoryWrites == 0 && !value.Effects.MutationAuthority && contentDigest(value.Digest)
	}
	if !valid(first) || !valid(replay) {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_SOURCE_MEANING_MISMATCH")
	}
	if !bytes.Equal(artifactPayload, replayPayload) {
		return ArtifactMeaning{}, fmt.Errorf("ARTIFACT_REPLAY_MISMATCH")
	}
	if !bound(input, "source-check", sourcePayload) || !bound(input, "project-manifest", artifactPayload) || !bound(input, "replay-manifest", replayPayload) {
		return ArtifactMeaning{}, fmt.Errorf("PRODUCER_OUTPUT_DIGEST_MISMATCH")
	}
	return ArtifactMeaning{Activity: first.Operation.Activity, Inputs: first.Operation.Inputs, Output: first.Operation.Output, Decision: first.Decision, Reason: first.Reason}, nil
}

func resourceEnvelope(input Input) ResourceEnvelope {
	result := ResourceEnvelope{Decision: "PASS", Resolution: "EXACT", Operations: len(input.Contract.Operations), Samples: len(input.Observations), ExpectedSamples: len(input.Contract.Operations) * input.Contract.SamplesPerOp, Runner: input.Producer.Runner, PerOperation: []ResourceOperation{}}
	violations := 0
	if !validRunner(input.Producer.Runner) || len(input.Observations) != result.ExpectedSamples {
		result.Decision, result.Resolution = "FAIL_CLOSED", "LOWER_RESOLUTION"
	}
	for _, spec := range input.Contract.Operations {
		values := make([]Observation, 0, input.Contract.SamplesPerOp)
		for _, value := range input.Observations {
			if value.Operation == spec.ID {
				values = append(values, value)
			}
		}
		summary := ResourceOperation{Operation: spec.ID, Samples: len(values)}
		walls := make([]int64, 0, len(values))
		for index, value := range values {
			if index >= input.Contract.SamplesPerOp || value.Sequence != index+1 || !validObservation(value, spec, input) {
				result.Decision, result.Resolution = "FAIL_CLOSED", "LOWER_RESOLUTION"
			}
			walls = append(walls, value.WallTimeNS)
			if value.WallTimeNS > input.Contract.Limits.WallTimeMS*1000000 || value.PeakRSSKiB > input.Contract.Limits.PeakRSSKiB || value.ReceiptBytes > input.Contract.Limits.ReceiptBytes || value.GeneratedBytes > input.Contract.Limits.GeneratedBytes {
				violations++
				summary.BudgetViolations++
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
		}
		if len(values) != input.Contract.SamplesPerOp {
			result.Decision, result.Resolution = "FAIL_CLOSED", "LOWER_RESOLUTION"
		}
		result.PerOperation = append(result.PerOperation, summary)
	}
	if violations > 0 {
		result.Decision, result.Resolution = "FAIL_CLOSED", "EXACT"
	}
	return result
}

func transitions(semanticPass bool, resource ResourceEnvelope, writeSet WriteSetObservation, effects Effects) []ClaimTransition {
	semanticTo, semanticReason := "DISCHARGED", "SEMANTIC_SOURCE_LOWERING_AND_ARTIFACT_REPLAY_STABLE"
	if !semanticPass {
		semanticTo, semanticReason = "REFUTED", "SEMANTIC_EVIDENCE_INVALID"
	}
	resourceTo, resourceReason := "DISCHARGED", "RESOURCE_ENVELOPE_OBSERVED"
	if resource.Resolution != "EXACT" {
		resourceTo, resourceReason = "OPEN", "RESOURCE_SAMPLE_MISSING_OR_INVALID"
	} else if resource.Decision != "PASS" {
		resourceTo, resourceReason = "REFUTED", "RESOURCE_BUDGET_EXCEEDED"
	}
	writeSetTo, writeSetReason := writeSetState(writeSet, effects)
	return []ClaimTransition{{Sequence: 1, ClaimID: "semantic-meaning", From: "OPEN", To: semanticTo, Stage: "REDUCE", Step: "semantic-verdict", Reason: semanticReason}, {Sequence: 2, ClaimID: "runner-resource-envelope", From: "OPEN", To: resourceTo, Stage: "REDUCE", Step: "resource-verdict", Reason: resourceReason}, {Sequence: 3, ClaimID: "read-only-observation", From: "OPEN", To: writeSetTo, Stage: "REDUCE", Step: "effect-verdict", Reason: writeSetReason}}
}

func writeSetState(value WriteSetObservation, effects Effects) (string, string) {
	if effects.RepositoryWrites != 0 || effects.MutationAuthority || value.RepositoryWrites != 0 || value.MutationAuthority || value.DiffExitCode != 0 || len(value.ChangedPaths) != 0 || value.UntrackedFileCount != 0 {
		return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
	}
	if value.Schema != "gooo/meta-resource-budget-write-set/v1" || value.Producer != Producer || value.Consumer != Consumer || !gitDigest(value.BeforeTreeDigest) || !gitDigest(value.AfterTreeDigest) || !contentDigest(value.WriteSetDigest) || value.BeforeTreeDigest != value.AfterTreeDigest || value.Reason != "GIT_DIFF_EXIT_0_AND_WRITE_SET_EMPTY" {
		return "OPEN", "EFFECT_OBSERVATION_MISSING"
	}
	return "DISCHARGED", "EFFECT_BOUNDARY_VERIFIED"
}

func overallDecision(semantic string, resource ResourceEnvelope, transitions []ClaimTransition) (string, string, string) {
	if transitions[2].To == "REFUTED" {
		return "FAIL_CLOSED", "EXACT", "EFFECT_BOUNDARY_VIOLATED"
	}
	if transitions[2].To == "OPEN" {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "EFFECT_OBSERVATION_MISSING"
	}
	if semantic != "PASS" {
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

func validObservation(value Observation, spec Operation, input Input) bool {
	return value.Schema == "gooo/meta-resource-budget-observation/v1" && value.SubjectSHA == input.ExpectedHead && value.Producer == Producer && value.Consumer == Consumer && value.Stage == spec.Stage && value.Step == spec.Step && value.MetaOperation == spec.MetaOperation && value.ProofChoice == spec.ProofChoice && value.Reason == "RUNNER_RESOURCE_OBSERVED" && value.ExitCode == 0 && value.WallTimeNS > 0 && value.PeakRSSKiB > 0 && value.ReceiptBytes >= 0 && value.GeneratedBytes >= 0 && contentDigest(value.OutputDigest)
}

func validRunner(value Runner) bool {
	return value.OS != "" && value.Architecture != "" && value.Image != "" && value.ImageVersion != "" && value.GoVersion == "go1.27.0"
}
func gitDigest(value string) bool { return len(value) == 40 && isHex(value) }
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
func bound(input Input, operation string, output []byte) bool {
	for _, value := range input.Observations {
		if value.Operation == operation && value.Sequence == 1 {
			return value.OutputDigest == digestBytes(output)
		}
	}
	return false
}

func sourceSetDigest(sources []RawSource) string {
	values := append([]RawSource(nil), sources...)
	sort.Slice(values, func(left, right int) bool { return values[left].Filename < values[right].Filename })
	var payload bytes.Buffer
	for _, raw := range values {
		content, _ := decode(raw.ContentBase64)
		digest := sha256.Sum256(content)
		payload.WriteString(hex.EncodeToString(digest[:]))
		payload.WriteByte('\n')
	}
	return digestBytes(payload.Bytes())
}

func firstSourceFilename(sources []RawSource) string {
	values := append([]RawSource(nil), sources...)
	sort.Slice(values, func(left, right int) bool { return values[left].Filename < values[right].Filename })
	if len(values) == 0 {
		return ""
	}
	return values[0].Filename
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}

func sealTransition(value ClaimTransition) ClaimTransition {
	value.Digest = ""
	value.Digest = digestValue(value)
	return value
}
