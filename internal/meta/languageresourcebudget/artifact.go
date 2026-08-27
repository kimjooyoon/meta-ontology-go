package languageresourcebudget

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

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
	Package        string
	Namespace      string
	Activity       string
	Inputs         []binding
	Output         binding
}

func verifyProducer(input Input) (Semantic, error) {
	meaning, err := reconstructSourceMeaning(input)
	if err != nil {
		return Semantic{}, err
	}
	if input.Producer.SourceDigest != meaning.SourceDigest {
		return Semantic{}, fmt.Errorf("SOURCE_DIGEST_MISMATCH")
	}
	sourcePayload, err := decodePayload(input.Producer.SourceReceiptBase64)
	if err != nil {
		return Semantic{}, fmt.Errorf("SOURCE_RECEIPT_INVALID")
	}
	var source sourceReceipt
	if err := json.Unmarshal(sourcePayload, &source); err != nil {
		return Semantic{}, fmt.Errorf("SOURCE_RECEIPT_INVALID")
	}
	if source.SchemaVersion != "gooo/diagnostics/v1" || source.Command != "check" || source.Status != "ok" ||
		source.File != firstSourceFilename(input) || len(source.Diagnostics) != 0 {
		return Semantic{}, fmt.Errorf("SOURCE_RECEIPT_NOT_EXACT")
	}
	artifactPayload, err := decodePayload(input.Producer.ArtifactBase64)
	if err != nil {
		return Semantic{}, fmt.Errorf("ARTIFACT_INVALID")
	}
	replayPayload, err := decodePayload(input.Producer.ReplayBase64)
	if err != nil {
		return Semantic{}, fmt.Errorf("REPLAY_INVALID")
	}
	var first, replay artifact
	if err := json.Unmarshal(artifactPayload, &first); err != nil {
		return Semantic{}, fmt.Errorf("ARTIFACT_INVALID")
	}
	if err := json.Unmarshal(replayPayload, &replay); err != nil {
		return Semantic{}, fmt.Errorf("REPLAY_INVALID")
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
	if !valid(first) || !valid(replay) {
		return Semantic{}, fmt.Errorf("ARTIFACT_SEMANTICS_INVALID")
	}
	firstDigest, replayDigest := digestBytes(artifactPayload), digestBytes(replayPayload)
	if !bytes.Equal(artifactPayload, replayPayload) {
		return Semantic{Decision: "FAIL_CLOSED", Resolution: "EXACT", Reason: "ARTIFACT_REPLAY_MISMATCH", SourceDigest: meaning.SourceDigest, SemanticDigest: meaning.SemanticDigest, ArtifactDigest: firstDigest, ReplayDigest: replayDigest}, fmt.Errorf("ARTIFACT_REPLAY_MISMATCH")
	}
	if !boundOutputDigest(input, "source-check", sourcePayload) ||
		!boundOutputDigest(input, "project-manifest", artifactPayload) ||
		!boundOutputDigest(input, "replay-manifest", replayPayload) {
		return Semantic{Decision: "FAIL_CLOSED", Resolution: "EXACT", Reason: "PRODUCER_OUTPUT_DIGEST_MISMATCH", SourceDigest: meaning.SourceDigest, SemanticDigest: meaning.SemanticDigest, ArtifactDigest: firstDigest, ReplayDigest: replayDigest}, fmt.Errorf("PRODUCER_OUTPUT_DIGEST_MISMATCH")
	}
	return Semantic{Decision: "PASS", Resolution: "EXACT", Reason: "SEMANTIC_SOURCE_LOWERING_AND_ARTIFACT_REPLAY_STABLE", SourceDigest: meaning.SourceDigest, SemanticDigest: meaning.SemanticDigest, ArtifactDigest: firstDigest, ReplayDigest: replayDigest}, nil
}

func reconstructSourceMeaning(input Input) (sourceMeaning, error) {
	if len(input.Producer.SourceFiles) != input.Producer.SourceFileCount || len(input.Producer.SourceFiles) != len(input.Contract.SourcePaths) {
		return sourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
	}
	sources := append([]RawSource(nil), input.Producer.SourceFiles...)
	sort.Slice(sources, func(left, right int) bool { return sources[left].Filename < sources[right].Filename })
	var declarations []syntax.Declaration
	var packageName, namespace string
	for index, raw := range sources {
		if raw.Filename == "" {
			return sourceMeaning{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
		}
		content, err := decodePayload(raw.ContentBase64)
		if err != nil || len(content) == 0 {
			return sourceMeaning{}, fmt.Errorf("SOURCE_PAYLOAD_INVALID")
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
	return sourceMeaning{SourceDigest: sourceSetDigest(sources), SemanticDigest: "sha256:" + ir.StableHash(), Package: packageName, Namespace: namespace, Activity: activity.Name, Inputs: inputs, Output: binding{Name: outputName, ID: output.ID}}, nil
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
		digest := sha256.Sum256(content)
		payload.WriteString(hex.EncodeToString(digest[:]))
		payload.WriteByte('\n')
	}
	return digestBytes(payload.Bytes())
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

func boundOutputDigest(input Input, operation string, output []byte) bool {
	for _, value := range input.Observations {
		if value.Operation == operation && value.Sequence == 1 {
			return value.OutputDigest == digestBytes(output)
		}
	}
	return false
}
