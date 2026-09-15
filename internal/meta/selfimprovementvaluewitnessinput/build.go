package selfimprovementvaluewitnessinput

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func CanonicalCorpus() []ValueCase {
	return []ValueCase{
		{ID: "negative", Input: -2, ExpectedOutput: -1},
		{ID: "negative-to-zero", Input: -1, ExpectedOutput: 0},
		{ID: "zero", Input: 0, ExpectedOutput: 1},
		{ID: "positive", Input: 41, ExpectedOutput: 42},
		{ID: "maximum-boundary", Input: math.MaxInt64 - 1, ExpectedOutput: math.MaxInt64},
	}
}

func KnownRegistry() RegistryIdentity {
	registry := RegistryIdentity{
		Schema:  "gooo/self-improvement-value-witness-evaluator-registry/v1",
		ID:      "gooo://self-improvement/value-witness-evaluator-registry/v1",
		Version: "v1", EvaluatorID: EvaluatorID, EvaluatorVersion: EvaluatorVersion,
		Operations: valueexecution.CanonicalOperationSpecs(),
	}
	registry.Digest = digestJSON(registry)
	return registry
}

func Build(repository fs.FS, sourcePath, activity, candidateStableID, candidateDigest, subjectSHA, observationDigest string) (ExecutionInput, error) {
	if sourcePath != SourcePath || activity != ActivityName {
		return ExecutionInput{}, errors.New("value-witness execution input source or activity is not the registered pair")
	}
	raw, err := fs.ReadFile(repository, sourcePath)
	if err != nil {
		return ExecutionInput{}, fmt.Errorf("read value-witness source: %w", err)
	}
	return BuildFromBytes(sourcePath, raw, activity, candidateStableID, candidateDigest, subjectSHA, observationDigest)
}

func BuildFromBytes(sourcePath string, raw []byte, activity, candidateStableID, candidateDigest, subjectSHA, observationDigest string) (ExecutionInput, error) {
	if sourcePath != SourcePath || activity != ActivityName {
		return ExecutionInput{}, errors.New("value-witness execution input source or activity is not the registered pair")
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(raw))
	if diagnostics.HasErrors() {
		return ExecutionInput{}, errors.New("value-witness source has syntax errors")
	}
	program, err := valueexecution.Compile(sourcePath, raw, activity)
	if err != nil {
		return ExecutionInput{}, fmt.Errorf("compile value-witness source for binding: %w", err)
	}
	declaration, ok := activityDeclaration(file, activity)
	if !ok {
		return ExecutionInput{}, errors.New("registered value-witness activity declaration is absent")
	}
	if file.Namespace == nil {
		return ExecutionInput{}, errors.New("registered value-witness namespace is absent")
	}
	if declaration.ValueProgram != ValueProgram || program.Operation.Program != ValueProgram {
		return ExecutionInput{}, errors.New("registered value-witness program is not the exact operation")
	}
	registry := KnownRegistry()
	if len(registry.Operations) != 1 || !reflect.DeepEqual(registry.Operations[0], program.Operation.Spec) {
		return ExecutionInput{}, errors.New("compiled operation is not the registered evaluator operation")
	}
	span := sourceSpan(declaration.Span)
	corpus := CanonicalCorpus()
	input := ExecutionInput{
		Schema: Schema, ContractID: ContractID, CandidateStableID: candidateStableID,
		CandidateDigest: candidateDigest, SubjectSHA: subjectSHA, ObservationDigest: observationDigest,
		OperationID: OperationID, BoundedTarget: BoundedTarget, Phase: Phase,
		Source: SourceSnapshot{Path: sourcePath, Bytes: string(raw), Digest: digestBytes(raw)},
		Activity: ActivityIdentity{
			DeclarationKind: "Activity", Name: declaration.Name,
			QualifiedName: file.Namespace.Name + "." + declaration.Name,
			InputEntities: append([]string(nil), program.Operation.InputEntities...),
			OutputEntity:  program.Operation.OutputEntity, ValueProgram: declaration.ValueProgram,
			ValueProgramDigest:  digestBytes([]byte(declaration.ValueProgram)),
			SemanticFingerprint: program.SemanticFingerprint, ASTSpan: span,
		},
		Corpus: corpus, CorpusDigest: CorpusDigest(corpus), AllowedEffects: []string{},
		EvaluatorRegistry: registry, ToolchainTestContractID: ToolchainTestID,
		OutputSchema: OutputSchema, InputAuthority: InputAuthority, OutputAuthority: OutputAuthority,
		MaxExecutions: MaxExecutions, RepositoryWritesAllowed: false,
	}
	input.Digest = executionInputDigest(input)
	return input, nil
}

func BindCandidateDigest(input *ExecutionInput, candidateDigest string) {
	input.CandidateDigest = candidateDigest
	input.Digest = executionInputDigest(*input)
}

func Validate(input ExecutionInput) error {
	if input.Schema != Schema || input.ContractID != ContractID || input.OperationID != OperationID ||
		input.BoundedTarget != BoundedTarget || input.Phase != Phase || input.CandidateStableID == "" ||
		!validDigest(input.CandidateStableID) || !validDigest(input.CandidateDigest) ||
		!validSHA(input.SubjectSHA) || !validDigest(input.ObservationDigest) {
		return errors.New("execution input identity is incomplete")
	}
	if input.Source.Path != SourcePath || input.Source.Bytes == "" || !validDigest(input.Source.Digest) || input.Source.Digest != digestBytes([]byte(input.Source.Bytes)) {
		return errors.New("execution input source snapshot is not exact")
	}
	if input.Activity.DeclarationKind != "Activity" || input.Activity.Name != ActivityName ||
		input.Activity.QualifiedName != "valuewitness."+ActivityName || input.Activity.ValueProgram != ValueProgram ||
		!validDigest(input.Activity.ValueProgramDigest) || input.Activity.ValueProgramDigest != digestBytes([]byte(ValueProgram)) || input.Activity.SemanticFingerprint == "" {
		return errors.New("execution input activity identity is not exact")
	}
	file, diagnostics := syntax.ParseFile(input.Source.Path, input.Source.Bytes)
	if diagnostics.HasErrors() {
		return errors.New("execution input source snapshot does not parse")
	}
	declaration, ok := activityDeclaration(file, ActivityName)
	if !ok || sourceSpan(declaration.Span) != input.Activity.ASTSpan {
		return errors.New("execution input AST span is not bound to the source snapshot")
	}
	program, err := valueexecution.Compile(input.Source.Path, []byte(input.Source.Bytes), ActivityName)
	if err != nil || program.SemanticFingerprint != input.Activity.SemanticFingerprint ||
		program.Operation.Program != ValueProgram || program.Operation.OutputEntity != input.Activity.OutputEntity ||
		!reflect.DeepEqual(program.Operation.InputEntities, input.Activity.InputEntities) {
		return errors.New("execution input semantic activity identity is not exact")
	}
	if !reflect.DeepEqual(input.Corpus, CanonicalCorpus()) || !validDigest(input.CorpusDigest) || input.CorpusDigest != CorpusDigest(input.Corpus) {
		return errors.New("execution input corpus is not exact")
	}
	registry := KnownRegistry()
	if !reflect.DeepEqual(input.EvaluatorRegistry.Operations, registry.Operations) ||
		input.EvaluatorRegistry.Schema != registry.Schema || input.EvaluatorRegistry.ID != registry.ID ||
		input.EvaluatorRegistry.Version != registry.Version || input.EvaluatorRegistry.EvaluatorID != registry.EvaluatorID ||
		input.EvaluatorRegistry.EvaluatorVersion != registry.EvaluatorVersion || !validDigest(input.EvaluatorRegistry.Digest) || input.EvaluatorRegistry.Digest != registry.Digest {
		return errors.New("execution input evaluator registry is not exact")
	}
	if len(input.AllowedEffects) != 0 || input.ToolchainTestContractID != ToolchainTestID || input.OutputSchema != OutputSchema ||
		input.InputAuthority != InputAuthority || input.OutputAuthority != OutputAuthority || input.MaxExecutions != MaxExecutions ||
		input.RepositoryWritesAllowed || !validDigest(input.Digest) || input.Digest != executionInputDigest(input) {
		return errors.New("execution input safety or digest binding is not exact")
	}
	return nil
}

func activityDeclaration(file *syntax.File, name string) (*syntax.ActivityDecl, bool) {
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	for _, declaration := range declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok && activity != nil && activity.Name == name {
			return activity, true
		}
	}
	return nil, false
}

func sourceSpan(span syntax.Span) SourceSpan {
	return SourceSpan{SourceID: span.Filename, StartByte: span.Start.Offset, EndByte: span.End.Offset,
		StartLine: span.Start.Line, StartColumn: span.Start.Column, EndLine: span.End.Line, EndColumn: span.End.Column}
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
