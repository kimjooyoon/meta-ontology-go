package languageresourcebudget

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
		Inputs   []binding `json:"inputs"`
		Output   binding   `json:"output"`
	} `json:"operation"`
	Effects Effects `json:"effects"`
	Digest  string  `json:"digest"`
}

type binding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type sourceMeaning struct {
	SourceDigest   string
	SemanticDigest string
	TargetDigest   string
	Package        string
	Namespace      string
	Activity       string
	Inputs         []binding
	Output         binding
}

func verifyProducer(input Input) (Semantic, error) {
	meaning, err := reconstructSourceMeaning(input)
	if err != nil {
		return semanticFailure(err.Error(), sourceMeaning{})
	}
	if input.Producer.SourceDigest != meaning.SourceDigest {
		return semanticFailure("SOURCE_DIGEST_MISMATCH", meaning)
	}

	semantic := Semantic{SourceDigest: meaning.SourceDigest, SemanticDigest: meaning.SemanticDigest, TargetDigest: meaning.TargetDigest}
	if input.Producer.SourceReceiptBase64 == "" {
		return semanticFailure("SOURCE_RECEIPT_PAYLOAD_MISSING", meaning)
	}
	sourcePayload, err := decodePayload(input.Producer.SourceReceiptBase64)
	if err != nil {
		return semanticFailure("SOURCE_RECEIPT_PAYLOAD_MISSING", meaning)
	}
	var source sourceReceipt
	if err := json.Unmarshal(sourcePayload, &source); err != nil {
		return semanticFailure("SOURCE_RECEIPT_INVALID", meaning)
	}
	if source.SchemaVersion != "gooo/diagnostics/v1" || source.Command != "check" || source.Status != "ok" ||
		source.File != firstSourceFilename(input.Producer.SourceFiles) || len(source.Diagnostics) != 0 {
		return semanticFailure("SOURCE_RECEIPT_NOT_EXACT", meaning)
	}

	if input.Producer.ArtifactBase64 == "" {
		return semanticFailure("ARTIFACT_PAYLOAD_MISSING", meaning)
	}
	artifactPayload, err := decodePayload(input.Producer.ArtifactBase64)
	if err != nil {
		return semanticFailure("ARTIFACT_PAYLOAD_MISSING", meaning)
	}
	if input.Producer.ReplayBase64 == "" {
		return semanticFailure("REPLAY_PAYLOAD_MISSING", meaning)
	}
	replayPayload, err := decodePayload(input.Producer.ReplayBase64)
	if err != nil {
		return semanticFailure("REPLAY_PAYLOAD_MISSING", meaning)
	}
	var first, replay artifact
	if err := json.Unmarshal(artifactPayload, &first); err != nil {
		return semanticFailure("ARTIFACT_INVALID", meaning)
	}
	if err := json.Unmarshal(replayPayload, &replay); err != nil {
		return semanticFailure("REPLAY_INVALID", meaning)
	}
	valid := func(value artifact) bool {
		return value.Schema == "gooo/operation-manifest/v1" && value.Decision == "PASS" && value.Resolution == "EXACT" &&
			value.Reason == "OPERATION_MANIFEST_EMITTED" && value.Kind == "operation-manifest" &&
			value.Operation.Activity == meaning.Activity && value.Operation.Inputs != nil &&
			bindingsEqual(value.Operation.Inputs, meaning.Inputs) && value.Operation.Output == meaning.Output &&
			value.Package.Name == meaning.Package && value.Package.Namespace == meaning.Namespace &&
			value.Effects.RepositoryWrites == 0 && !value.Effects.MutationAuthority &&
			contentDigest(value.SubjectDigest) && contentDigest(value.Digest)
	}
	firstDigest, replayDigest := digestBytes(artifactPayload), digestBytes(replayPayload)
	semantic.ArtifactDigest, semantic.ReplayDigest = firstDigest, replayDigest
	if !valid(first) || !valid(replay) {
		return semanticFailureWithValue("ARTIFACT_SEMANTICS_INVALID", meaning, semantic)
	}
	if !bytes.Equal(artifactPayload, replayPayload) {
		return semanticFailureWithValue("ARTIFACT_REPLAY_MISMATCH", meaning, semantic)
	}
	if err := verifyRawOutputs(input, meaning, sourcePayload, artifactPayload, replayPayload); err != nil {
		return semanticFailureWithValue(err.Error(), meaning, semantic)
	}
	semantic.Decision, semantic.Resolution, semantic.ClaimState = "PASS", "EXACT", "DISCHARGED"
	semantic.Reason = "SEMANTIC_SOURCE_LOWERING_AND_ARTIFACT_REPLAY_STABLE"
	return semantic, nil
}

func semanticFailure(reason string, meaning sourceMeaning) (Semantic, error) {
	return semanticFailureWithValue(reason, meaning, Semantic{SourceDigest: meaning.SourceDigest, SemanticDigest: meaning.SemanticDigest, TargetDigest: meaning.TargetDigest})
}

func semanticFailureWithValue(reason string, meaning sourceMeaning, semantic Semantic) (Semantic, error) {
	semantic.Decision, semantic.Reason = "FAIL_CLOSED", reason
	if missingSemanticReason(reason) {
		semantic.Resolution, semantic.ClaimState = "LOWER_RESOLUTION", "OPEN"
	} else {
		semantic.Resolution, semantic.ClaimState = "EXACT", "REFUTED"
	}
	return semantic, fmt.Errorf(reason)
}

func missingSemanticReason(reason string) bool {
	return strings.Contains(reason, "MISSING") || reason == "SOURCE_FILE_SET_MISSING" || reason == "RAW_OPERATION_OUTPUTS_MISSING"
}

func verifyRawOutputs(input Input, meaning sourceMeaning, sourcePayload, artifactPayload, replayPayload []byte) error {
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
		payload, err := decodePayload(raw.PayloadBase64)
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
	if !bytes.Equal(firstOutput(byKey, "source-check"), sourcePayload) ||
		!bytes.Equal(firstOutput(byKey, "project-manifest"), artifactPayload) ||
		!bytes.Equal(firstOutput(byKey, "replay-manifest"), replayPayload) {
		return fmt.Errorf("PRODUCER_OUTPUT_PAYLOAD_MISMATCH")
	}
	return nil
}

func firstOutput(outputs map[string]RawOutput, operationID string) []byte {
	value, ok := outputs[outputKey(operationID, 1)]
	if !ok {
		return nil
	}
	payload, _ := decodePayload(value.PayloadBase64)
	return payload
}

func outputKey(operationID string, sequence int) string {
	return fmt.Sprintf("%s#%d", operationID, sequence)
}

func outputDigestForObservation(observations []Observation, operationID string, sequence int, digest string) bool {
	count := 0
	for _, observation := range observations {
		if observation.Operation == operationID && observation.Sequence == sequence {
			count++
			if observation.OutputDigest != digest {
				return false
			}
		}
	}
	return count == 1
}

func outputObservationsMatch(observations []Observation, operationID, digest string, meaning sourceMeaning) bool {
	count := 0
	for _, observation := range observations {
		if observation.Operation != operationID {
			continue
		}
		count++
		if observation.OutputDigest != digest || observation.SourceRawDigest != meaning.SourceDigest || observation.SourceSemanticDigest != meaning.SemanticDigest || observation.EntryDigest != meaning.TargetDigest || observation.TargetDigest != meaning.TargetDigest {
			return false
		}
	}
	return count > 0
}

func reconstructSourceMeaning(input Input) (sourceMeaning, error) {
	if len(input.Producer.SourceFiles) < input.Producer.SourceFileCount || len(input.Producer.SourceFiles) < len(input.Contract.SourcePaths) {
		return sourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_MISSING")
	}
	if len(input.Producer.SourceFiles) != input.Producer.SourceFileCount || len(input.Producer.SourceFiles) != len(input.Contract.SourcePaths) {
		return sourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
	}
	sources := append([]RawSource(nil), input.Producer.SourceFiles...)
	sort.Slice(sources, func(left, right int) bool { return sources[left].Filename < sources[right].Filename })
	var declarations []syntax.Declaration
	var packageName, namespace string
	for index, raw := range sources {
		if raw.Filename == "" || !strings.HasSuffix(raw.Filename, ".gooo") {
			return sourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
		}
		content, err := decodePayload(raw.ContentBase64)
		if err != nil {
			return sourceMeaning{}, fmt.Errorf("SOURCE_PAYLOAD_INVALID")
		}
		if len(content) == 0 {
			return sourceMeaning{}, fmt.Errorf("SOURCE_PAYLOAD_MISSING")
		}
		file, diagnostics := syntax.ParseFile(raw.Filename, string(content))
		if file == nil || diagnostics.HasErrors() || file.Package == nil || file.Namespace == nil {
			return sourceMeaning{}, fmt.Errorf("SOURCE_SYNTAX_INVALID")
		}
		if index == 0 {
			packageName, namespace = file.Package.Name, file.Namespace.Name
		} else if file.Package.Name != packageName || file.Namespace.Name != namespace {
			return sourceMeaning{}, fmt.Errorf("SOURCE_HEADER_MISMATCH")
		}
		declarations = append(declarations, file.Decls...)
	}
	combined := &syntax.File{Package: &syntax.PackageDecl{Name: packageName}, Namespace: &syntax.NamespaceDecl{Name: namespace}, Decls: declarations, Declarations: declarations}
	if _, err := syntax.Format(combined); err != nil {
		return sourceMeaning{}, fmt.Errorf("SOURCE_CANONICAL_FORMAT_INVALID")
	}
	ir, err := bidir.Lower(combined)
	if err != nil {
		return sourceMeaning{}, fmt.Errorf("SOURCE_SEMANTIC_LOWERING_INVALID")
	}
	if err := ir.Validate(); err != nil {
		return sourceMeaning{}, fmt.Errorf("SOURCE_SEMANTIC_VALIDATION_INVALID")
	}
	activities := make([]*syntax.ActivityDecl, 0, 1)
	entities := make(map[string]binding)
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.ActivityDecl:
			activities = append(activities, value)
		case *syntax.EntityDecl:
			entities[value.Name] = binding{Name: value.Name, ID: value.ID}
		}
	}
	if len(activities) != 1 {
		return sourceMeaning{}, fmt.Errorf("SOURCE_ENTRY_CARDINALITY_INVALID")
	}
	activity := activities[0]
	parameters := activity.Inputs
	if parameters == nil {
		parameters = activity.Parameters
	}
	inputs := make([]binding, 0, len(parameters))
	for _, parameter := range parameters {
		value, ok := entities[parameter.Name]
		if !ok {
			return sourceMeaning{}, fmt.Errorf("SOURCE_INPUT_ENTITY_UNKNOWN")
		}
		inputs = append(inputs, value)
	}
	outputName := activity.Output
	if outputName == "" {
		outputName = activity.Result.Name
	}
	output, ok := entities[outputName]
	if !ok {
		return sourceMeaning{}, fmt.Errorf("SOURCE_OUTPUT_ENTITY_UNKNOWN")
	}
	meaning := sourceMeaning{SourceDigest: sourceSetDigest(sources), SemanticDigest: "sha256:" + ir.StableHash(), Package: packageName, Namespace: namespace, Activity: activity.Name, Inputs: inputs, Output: binding{Name: outputName, ID: output.ID}}
	meaning.TargetDigest = targetDigest(meaning)
	return meaning, nil
}

func firstSourceFilename(input Input) string {
	values := append([]RawSource(nil), input.Producer.SourceFiles...)
	sort.Slice(values, func(left, right int) bool { return values[left].Filename < values[right].Filename })
	if len(values) == 0 {
		return ""
	}
	return values[0].Filename
}

func sourceSetDigest(sources []RawSource) string {
	values := append([]RawSource(nil), sources...)
	sort.Slice(values, func(left, right int) bool { return values[left].Filename < values[right].Filename })
	var payload bytes.Buffer
	for _, raw := range values {
		content, _ := decodePayload(raw.ContentBase64)
		payload.WriteString(raw.Filename)
		payload.WriteByte(0)
		payload.Write(content)
		payload.WriteByte(0)
	}
	return digestBytes(payload.Bytes())
}

func targetDigest(meaning sourceMeaning) string {
	return digestValue(struct {
		Package   string
		Namespace string
		Activity  string
		Inputs    []binding
		Output    binding
	}{meaning.Package, meaning.Namespace, meaning.Activity, meaning.Inputs, meaning.Output})
}

func bindingsEqual(left, right []binding) bool {
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

func decodePayload(value string) ([]byte, error) { return base64.StdEncoding.DecodeString(value) }
