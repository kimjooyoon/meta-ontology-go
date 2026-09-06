package extractor

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type CallbackExtractionArtifact struct {
	Path   string `json:"path"`
	Source string `json:"source"`
	Digest string `json:"digest"`
	Lines  int    `json:"lines"`
}

type CallbackExtractionProposal struct {
	Schema             string                                `json:"schema"`
	LogicalPath        string                                `json:"logical_path"`
	Subject            string                                `json:"subject"`
	SourceDigest       string                                `json:"source_digest"`
	IntermediateDigest string                                `json:"intermediate_digest"`
	PackageDigest      string                                `json:"package_digest"`
	PartitionDigest    string                                `json:"partition_digest"`
	LineLimit          int                                   `json:"line_limit"`
	MaximumLines       int                                   `json:"maximum_lines"`
	Contract           generation.CallbackExtractionContract `json:"contract"`
	StructureProof     ClosurePreservationProof              `json:"structure_proof"`
	Artifacts          []CallbackExtractionArtifact          `json:"artifacts"`
	Claims             []CallbackExtractionClaim             `json:"claims"`
	Closed             int                                   `json:"closed"`
	Unknown            int                                   `json:"unknown"`
	Refuted            int                                   `json:"refuted"`
	OperationAdmission string                                `json:"operation_admission"`
	ApplyPermission    string                                `json:"apply_permission"`
}

// PlanCallbackExtraction uses the normal package renderer, without constructing
// an admitted Result. It accepts a source-bound function subject, not a preview.
func PlanCallbackExtraction(root, logical, subject string) (CallbackExtractionProposal, error) {
	contract, err := generation.LoadCallbackExtractionContract()
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	source, fset, file, err := readParsedSource(root, logical)
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	name, valid := strings.CutPrefix(subject, "func:")
	target := callbackPreviewFunction(file, name)
	if !valid || name == "" || target == nil || target.Body == nil {
		return CallbackExtractionProposal{}, fmt.Errorf("callback extraction subject is not a source function: %s", subject)
	}
	evidence, err := checkTypes(root, logical, fset, file, target)
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	callback, err := callbackPreviewFuncLit(target, evidence.info)
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	captures, err := callbackPreviewCaptures(callback, evidence, fset)
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	effects := callbackPreviewEffects(callback, evidence, fset)
	candidate, err := callbackPreviewFactoryCandidate(root, source, logical, fset, file, target, callback, evidence, captures, effects)
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	structure, err := proveClosurePreservation(root, logical, source, fset, file, target, callback, evidence, captures, candidate)
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	generated, paths, err := renderCallbackExtractionPackage(root, logical, []byte(candidate.CandidateSource))
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	partition, err := callbackExtractionPartitionDigest([]byte(candidate.CandidateSource), generated)
	if err != nil {
		return CallbackExtractionProposal{}, err
	}
	if err := projectedFinalConformance(root, logical, generated); err != nil {
		return CallbackExtractionProposal{}, err
	}
	proposal := CallbackExtractionProposal{
		Schema: "gooo/callback-extraction-proposal/v1", LogicalPath: logical, Subject: subject,
		SourceDigest: callbackPreviewDigest(source), IntermediateDigest: candidate.CandidateDigest,
		PackageDigest: proofDigest(generatedPackagePayload(generated)), PartitionDigest: partition,
		LineLimit: functionLineLimit, Contract: contract, StructureProof: structure,
		OperationAdmission: "UNKNOWN", ApplyPermission: "FORBIDDEN",
	}
	for _, path := range paths {
		raw := generated[path]
		lines := physicalLines(raw)
		proposal.MaximumLines = max(proposal.MaximumLines, lines)
		proposal.Artifacts = append(proposal.Artifacts, CallbackExtractionArtifact{Path: path, Source: string(raw), Digest: proofDigest(raw), Lines: lines})
	}
	if err := bindCallbackExtractionClaims(&proposal); err != nil {
		return CallbackExtractionProposal{}, err
	}
	return proposal, nil
}

func renderCallbackExtractionPackage(root, logical string, source []byte) (map[string][]byte, []string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, source, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	list, declarations, err := extractionInputs(root, source, fset, file)
	if err != nil {
		return nil, nil, err
	}
	output, partitions, err := capacityRender(fset, file, source, declarations, list, functionLineLimit)
	if err != nil {
		return nil, nil, err
	}
	helpers, paths, err := renderHelperFiles(root, logical, source, fset, file, list, partitions)
	if err != nil {
		return nil, nil, err
	}
	generated := map[string][]byte{logical: output.source}
	maps.Copy(generated, helpers)
	for path, raw := range generated {
		if physicalLines(raw) > functionLineLimit {
			return nil, nil, failWithDiagnostics("plan-callback-extraction", "render-package", "CALLBACK_PACKAGE_CAPACITY_REFUTED", "KNOWN_CONTRADICTION", "report-counterexample", []string{"path=" + path})
		}
	}
	return generated, append([]string{logical}, paths...), nil
}

// The rejected normal recipe remains rejected. Its alternative proposal is
// carried as diagnostic data, never copied into Result.Generated.
func withCallbackExtractionProposal(root, logical string, original []byte, cause error) error {
	var failure Failure
	if !errors.As(cause, &failure) || (failure.Reason != "NO_SAFE_DECLARATION_CAPACITY" && failure.Reason != "CALLBACK_ENCLOSING_IDENTITY_UNPROVEN") {
		return cause
	}
	file, err := parser.ParseFile(token.NewFileSet(), logical, original, 0)
	if err != nil {
		return cause
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil || !callbackExtractionHasLiteral(function) {
			continue
		}
		proposal, err := PlanCallbackExtraction(root, logical, "func:"+function.Name.Name)
		if err != nil {
			failure.Diagnostics = append(failure.Diagnostics, "callback_extraction_planning_error="+function.Name.Name+":"+err.Error())
			continue
		}
		raw, err := json.Marshal(proposal)
		if err != nil {
			failure.Diagnostics = append(failure.Diagnostics, "callback_extraction_encoding_error="+err.Error())
			return failure
		}
		failure.Diagnostics = append(failure.Diagnostics, "callback_extraction_proposal="+string(raw))
		return failure
	}
	return failure
}

func callbackExtractionHasLiteral(function *ast.FuncDecl) bool {
	found := false
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			found = true
			return false
		}
		return !found
	})
	return found
}
