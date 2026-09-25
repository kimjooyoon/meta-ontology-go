package extractor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

const closurePreservationSchema = "gooo/closure-source-preservation/v1"

type ClosureCaptureReference struct {
	ObjectIdentity string `json:"object_identity"`
	References     int    `json:"references"`
}

// ClosurePreservationProof proves a bounded source relation, not arbitrary Go
// semantic equivalence. Consumers must replay VerifyCallbackFactoryPreservation
// against the source snapshot; validating a serialized summary is not replay.
type ClosurePreservationProof struct {
	Schema              string                    `json:"schema"`
	State               string                    `json:"state"`
	Scope               string                    `json:"scope"`
	SourceDigest        string                    `json:"source_digest"`
	CandidateDigest     string                    `json:"candidate_digest"`
	HelperName          string                    `json:"helper_name"`
	SourceBodyDigest    string                    `json:"source_body_digest"`
	CandidateBodyDigest string                    `json:"candidate_body_digest"`
	SourceContextDigest string                    `json:"source_context_digest"`
	OutputContextDigest string                    `json:"output_context_digest"`
	Captures            []ClosureCaptureReference `json:"captures"`
	CaptureReferences   int                       `json:"capture_references"`
	CallExpressions     int                       `json:"call_expressions"`
	ProofDigest         string                    `json:"proof_digest"`
	SemanticAdmission   string                    `json:"semantic_admission"`
	Stage               string                    `json:"stage"`
	Step                string                    `json:"step"`
	Reason              string                    `json:"reason"`
	UnknownClass        string                    `json:"unknown_class"`
	NextOperation       string                    `json:"next_operation"`
	BlockedBy           []string                  `json:"blocked_by"`
}

// VerifyCallbackFactoryPreservation independently reloads the original input.
// It does not trust the candidate's source digest or any claimed proof result.
func VerifyCallbackFactoryPreservation(root, logical string, candidate CallbackPreviewCandidate) (ClosurePreservationProof, error) {
	source, fset, file, err := readParsedSource(root, logical)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	target := callbackPreviewFunction(file, callbackPreviewTarget)
	if target == nil {
		return ClosurePreservationProof{}, closureProofFailure("original callback target is missing")
	}
	evidence, err := checkTypes(root, logical, fset, file, target)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	callback, err := callbackPreviewFuncLit(target, evidence.info)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	captures, err := callbackPreviewCaptures(callback, evidence, fset)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	return proveClosurePreservation(root, logical, source, fset, file, target, callback, evidence, captures, candidate)
}

func proveClosurePreservation(root, logical string, source []byte, fset *token.FileSet, file *ast.File, target *ast.FuncDecl, callback *ast.FuncLit, evidence typeEvidence, captures []CallbackPreviewCapture, candidate CallbackPreviewCandidate) (ClosurePreservationProof, error) {
	if candidate.SourceDigest != callbackPreviewDigest(source) || candidate.CandidateDigest != callbackPreviewDigest([]byte(candidate.CandidateSource)) {
		return ClosurePreservationProof{}, closureProofFailure("source or candidate snapshot digest differs")
	}
	output, err := closureProofOutput(root, logical, file, target, candidate)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	environment, err := closureProofBindings(fset, file, callback, evidence, captures, output)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	beforeBody, afterBody, references, err := closureProofBodies(source, fset, callback, evidence, output, environment)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	beforeContext, afterContext, err := closureProofContexts(source, fset, callback, output)
	if err != nil {
		return ClosurePreservationProof{}, err
	}
	if !bytes.Equal(beforeBody, afterBody) || !bytes.Equal(beforeContext, afterContext) {
		return ClosurePreservationProof{}, closureProofFailure("normalized callback body or surrounding source differs")
	}
	proof := ClosurePreservationProof{
		Schema: closurePreservationSchema, State: "SOURCE_STRUCTURE_PRESERVED", Scope: "AST_AND_TYPED_CAPTURE_IDENTITIES_ONLY",
		SourceDigest: candidate.SourceDigest, CandidateDigest: candidate.CandidateDigest, HelperName: candidate.HelperName,
		SourceBodyDigest: callbackPreviewDigest(beforeBody), CandidateBodyDigest: callbackPreviewDigest(afterBody),
		SourceContextDigest: callbackPreviewDigest(beforeContext), OutputContextDigest: callbackPreviewDigest(afterContext),
		Captures: references, SemanticAdmission: "UNKNOWN", Stage: "CALLBACK_PRESERVATION", Step: "SEMANTIC_ADMISSION",
		Reason: "CALLER_OBSERVABILITY_NOT_PROVEN", UnknownClass: "DIRECT_MISSING", NextOperation: "PROVE_CALLER_OBSERVABILITY", BlockedBy: []string{},
	}
	for _, reference := range references {
		proof.CaptureReferences += reference.References
	}
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.CallExpr); ok {
			proof.CallExpressions++
		}
		return true
	})
	proof.ProofDigest = closurePreservationDigest(proof)
	return proof, nil
}

func closurePreservationDigest(proof ClosurePreservationProof) string {
	proof.ProofDigest = ""
	raw, _ := json.Marshal(proof)
	return callbackPreviewDigest(append([]byte(closurePreservationSchema+"\x00"), raw...))
}

func closurePreservationRecordValues(proof *ClosurePreservationProof) map[string]string {
	if proof == nil {
		return map[string]string{"SourceStructureState": "UNOBSERVED", "SourceStructureDigest": "", "CaptureReferenceCount": "", "CallExpressionCount": ""}
	}
	return map[string]string{
		"SourceStructureState": proof.State, "SourceStructureDigest": proof.ProofDigest,
		"CaptureReferenceCount": strconv.Itoa(proof.CaptureReferences), "CallExpressionCount": strconv.Itoa(proof.CallExpressions),
	}
}

func validateClosurePreservationSummary(result CallbackPreviewResult) error {
	proof := result.StructureProof
	if result.LoweringStrategy != callbackPreviewFactoryLowering {
		if proof != nil {
			return fmt.Errorf("wrapper lowering cannot claim a factory structure proof")
		}
		return nil
	}
	if proof == nil || proof.Schema != closurePreservationSchema || proof.State != "SOURCE_STRUCTURE_PRESERVED" ||
		proof.Scope != "AST_AND_TYPED_CAPTURE_IDENTITIES_ONLY" || proof.SemanticAdmission != "UNKNOWN" ||
		proof.SourceDigest != result.SourceDigest || proof.CandidateDigest != result.Candidate.CandidateDigest || proof.HelperName != result.Candidate.HelperName ||
		proof.SourceBodyDigest == "" || proof.SourceBodyDigest != proof.CandidateBodyDigest ||
		proof.SourceContextDigest == "" || proof.SourceContextDigest != proof.OutputContextDigest ||
		proof.ProofDigest != closurePreservationDigest(*proof) || proof.CallExpressions < 0 || len(proof.Captures) != len(result.Captures) {
		return fmt.Errorf("callback source structure summary is not bound")
	}
	if proof.Stage != "CALLBACK_PRESERVATION" || proof.Step != "SEMANTIC_ADMISSION" || proof.Reason != "CALLER_OBSERVABILITY_NOT_PROVEN" ||
		proof.UnknownClass != "DIRECT_MISSING" || proof.NextOperation != "PROVE_CALLER_OBSERVABILITY" || proof.BlockedBy == nil || len(proof.BlockedBy) != 0 {
		return fmt.Errorf("callback structure proof lost its unknown admission boundary")
	}
	count := 0
	for index, reference := range proof.Captures {
		if reference.ObjectIdentity != result.Captures[index].ObjectIdentity || reference.References <= 0 {
			return fmt.Errorf("callback source structure capture is not bound")
		}
		count += reference.References
	}
	if count != proof.CaptureReferences {
		return fmt.Errorf("callback source structure reference count differs")
	}
	return nil
}

func closureProofFailure(detail string) error {
	return failWithDiagnostics("verify-candidate", "prove-closure-structure", "CLOSURE_STRUCTURE_REFUTED", "KNOWN_CONTRADICTION", "report-counterexample", []string{detail})
}
