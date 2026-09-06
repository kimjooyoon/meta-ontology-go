package extractor

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const (
	callbackPreviewSchema            = "gooo.callback-preview/v1"
	callbackPreviewTarget            = "TestPaginationFixturesExecuteParserAndHTTPClient"
	callbackPreviewStateUnknown      = "UNKNOWN"
	callbackPreviewPromotionNone     = "NONE"
	callbackPreviewStage             = "CALLBACK_PREVIEW"
	callbackPreviewUnknownDirect     = "DIRECT_MISSING"
	callbackPreviewUnknownDependency = "DEPENDENCY_BLOCKED"
	callbackPreviewUnknownAmbiguous  = "AMBIGUOUS"
	callbackPreviewUnknownUnbounded  = "UNBOUNDED"
	callbackPreviewWrapperLowering   = "wrapper"
	callbackPreviewFactoryLowering   = "closure-factory"
)

// CallbackPreviewResult is deliberately not Result. Its candidate is caller-
// owned preview data and cannot enter the normal OperationResult or staging
// path.
type CallbackPreviewResult struct {
	StructureProof           *ClosurePreservationProof          `json:"structure_proof,omitempty"`
	LoweringStrategy         string                             `json:"lowering_strategy"`
	Schema                   string                             `json:"schema"`
	LogicalPath              string                             `json:"logical_path"`
	Subject                  string                             `json:"subject"`
	SourceDigest             string                             `json:"source_digest"`
	ContractSourceDigest     string                             `json:"contract_source_digest"`
	ContractSemanticDigest   string                             `json:"contract_semantic_digest"`
	State                    string                             `json:"state"`
	Reason                   string                             `json:"reason"`
	Stage                    string                             `json:"stage"`
	Step                     string                             `json:"step"`
	UnknownClass             string                             `json:"unknown_class"`
	NextOperation            string                             `json:"next_operation"`
	BlockedBy                []string                           `json:"blocked_by"`
	Candidate                *CallbackPreviewCandidate          `json:"candidate,omitempty"`
	Captures                 []CallbackPreviewCapture           `json:"captures,omitempty"`
	PendingEffects           []CallbackPreviewEffect            `json:"pending_effects,omitempty"`
	Evidence                 CallbackPreviewEvidence            `json:"evidence"`
	ContractRecords          []generation.CallbackPreviewRecord `json:"contract_records,omitempty"`
	OperationResultAdmission string                             `json:"operation_result_admission"`
	ApplyPermission          string                             `json:"apply_permission"`
}

type CallbackPreviewCandidate struct {
	CandidateIdentity   string `json:"candidate_identity"`
	SourceDigest        string `json:"source_digest"`
	CandidateDigest     string `json:"candidate_digest"`
	HelperName          string `json:"helper_name"`
	WrapperSource       string `json:"wrapper_source"`
	HelperSource        string `json:"helper_source"`
	CandidateSource     string `json:"candidate_source"`
	HelperBytes         int    `json:"helper_bytes"`
	HelperLines         int    `json:"helper_lines"`
	ParentFunctionLines int    `json:"parent_function_lines"`
	CaptureCount        int    `json:"capture_count"`
	PendingEffectCount  int    `json:"pending_effect_count"`
	State               string `json:"state"`
	Promotion           string `json:"promotion"`
}

type CallbackPreviewCapture struct {
	Name           string `json:"name"`
	ObjectIdentity string `json:"object_identity"`
	ObjectType     string `json:"object_type"`
	BindingMode    string `json:"binding_mode"`
}

type CallbackPreviewEffect struct {
	CallIdentity  string   `json:"call_identity"`
	Symbol        string   `json:"symbol"`
	Signature     string   `json:"signature"`
	ReceiverType  string   `json:"receiver_type"`
	EffectKind    string   `json:"effect_kind"`
	State         string   `json:"state"`
	StartOffset   int      `json:"start_offset"`
	EndOffset     int      `json:"end_offset"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type CallbackPreviewEvidence struct {
	CandidateIdentity        string   `json:"candidate_identity"`
	SourceDigest             string   `json:"source_digest"`
	CandidateDigest          string   `json:"candidate_digest"`
	State                    string   `json:"state"`
	CaptureCount             int      `json:"capture_count"`
	PendingEffectCount       int      `json:"pending_effect_count"`
	ResolvedEffectCount      int      `json:"resolved_effect_count"`
	HelperLines              int      `json:"helper_lines"`
	ParentFunctionLines      int      `json:"parent_function_lines"`
	OperationResultAdmission string   `json:"operation_result_admission"`
	ApplyPermission          string   `json:"apply_permission"`
	Stage                    string   `json:"stage"`
	Step                     string   `json:"step"`
	Reason                   string   `json:"reason"`
	UnknownClass             string   `json:"unknown_class"`
	NextOperation            string   `json:"next_operation"`
	BlockedBy                []string `json:"blocked_by"`
}

// PreviewBoundedPaginationCallback observes and renders the exact callback
// shape used by the pagination fixture. It never writes the repository and
// never returns an accepted extraction Result.
func PreviewBoundedPaginationCallback(root, logical string) (CallbackPreviewResult, error) {
	return PreviewBoundedPaginationCallbackWithStrategy(root, logical, callbackPreviewWrapperLowering)
}

// PreviewBoundedPaginationCallbackWithStrategy selects a typed, recorded preview
// lowering. Neither lowering can authorize an OperationResult or repository write.
func PreviewBoundedPaginationCallbackWithStrategy(root, logical, lowering string) (CallbackPreviewResult, error) {
	if lowering != callbackPreviewWrapperLowering && lowering != callbackPreviewFactoryLowering {
		return CallbackPreviewResult{}, fmt.Errorf("unsupported callback preview lowering %q", lowering)
	}
	contract, err := generation.LoadCallbackPreviewContract()
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	source, fset, file, err := readParsedSource(root, logical)
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	base := CallbackPreviewResult{Schema: callbackPreviewSchema, LogicalPath: logical, Subject: "func:" + callbackPreviewTarget, SourceDigest: callbackPreviewDigest(source), ContractSourceDigest: contract.SourceDigest, ContractSemanticDigest: contract.SemanticDigest, State: callbackPreviewStateUnknown, Stage: callbackPreviewStage, BlockedBy: []string{}, OperationResultAdmission: "FORBIDDEN", ApplyPermission: "FORBIDDEN",
		LoweringStrategy: lowering}
	inputRecord, err := contract.BuildCallbackPreviewRecord(contract.InputEntity, map[string]string{"LogicalPath": logical, "LoweringStrategy": lowering, "Subject": base.Subject, "SourceDigest": base.SourceDigest, "State": base.State})
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	base.ContractRecords = []generation.CallbackPreviewRecord{inputRecord}
	target := callbackPreviewFunction(file, callbackPreviewTarget)
	if target == nil {
		callbackPreviewSetLifecycle(&base, "CALLBACK_TARGET_MISSING")
		base.Evidence = callbackPreviewEvidence(base, nil, 0, 0)
		base.ContractRecords = append(base.ContractRecords, callbackPreviewEvidenceRecord(contract, base, base.Evidence))
		return base, nil
	}
	typeEvidence, err := checkTypes(root, logical, fset, file, target)
	if err != nil {
		callbackPreviewSetLifecycle(&base, "TYPE_EVIDENCE_MISSING")
		base.Evidence = callbackPreviewEvidence(base, nil, 0, 0)
		base.ContractRecords = append(base.ContractRecords, callbackPreviewEvidenceRecord(contract, base, base.Evidence))
		return base, nil
	}
	callback, err := callbackPreviewFuncLit(target, typeEvidence.info)
	if err != nil {
		callbackPreviewSetLifecycle(&base, err.Error())
		base.Evidence = callbackPreviewEvidence(base, nil, 0, 0)
		base.ContractRecords = append(base.ContractRecords, callbackPreviewEvidenceRecord(contract, base, base.Evidence))
		return base, nil
	}
	captures, err := callbackPreviewCaptures(callback, typeEvidence, fset)
	if err != nil {
		callbackPreviewSetLifecycle(&base, "CALLBACK_CAPTURE_UNSUPPORTED")
		base.Evidence = callbackPreviewEvidence(base, nil, 0, 0)
		base.ContractRecords = append(base.ContractRecords, callbackPreviewEvidenceRecord(contract, base, base.Evidence))
		return base, nil
	}
	effects := callbackPreviewEffects(callback, typeEvidence, fset)
	var candidate CallbackPreviewCandidate
	if callbackPreviewRecordFieldMap(inputRecord)["LoweringStrategy"] == callbackPreviewFactoryLowering {
		candidate, err = callbackPreviewFactoryCandidate(root, source, logical, fset, file, target, callback, typeEvidence, captures, effects)
	} else {
		candidate, err = callbackPreviewCandidate(source, logical, fset, file, target, callback, captures, effects)
	}
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	base.Candidate = &candidate
	if lowering == callbackPreviewFactoryLowering {
		proof, proofErr := proveClosurePreservation(root, logical, source, fset, file, target, callback, typeEvidence, captures, candidate)
		if proofErr != nil {
			return CallbackPreviewResult{}, proofErr
		}
		base.StructureProof = &proof
	}
	base.Captures = captures
	base.PendingEffects = effects
	callbackPreviewSetLifecycle(&base, "PENDING_TYPED_CALLBACK_EFFECTS")
	base.BlockedBy = callbackPreviewEffectIdentities(effects)
	if candidate.HelperLines > functionLineLimit || candidate.ParentFunctionLines > functionLineLimit {
		callbackPreviewSetLifecycle(&base, "CALLBACK_CANDIDATE_OVER_CAPACITY")
		base.BlockedBy = []string{candidate.CandidateIdentity}
	}
	base.Evidence = callbackPreviewEvidence(base, &candidate, len(captures), len(effects))
	records, err := callbackPreviewRecords(contract, base, candidate, captures, effects)
	if err != nil {
		return CallbackPreviewResult{}, err
	}
	base.ContractRecords = records
	return base, nil
}

func callbackPreviewSetLifecycle(result *CallbackPreviewResult, reason string) {
	result.Reason = reason
	result.Stage = callbackPreviewStage
	result.UnknownClass = callbackPreviewUnknownDirect
	result.Step = "TARGET_DISCOVERY"
	result.NextOperation = "REVIEW_CALLBACK_TARGET"
	result.BlockedBy = []string{}
	switch reason {
	case "TYPE_EVIDENCE_MISSING":
		result.Step = "TYPE_EVIDENCE"
		result.NextOperation = "RECHECK_TYPE_EVIDENCE"
	case "CALLBACK_TARGET_SHAPE_UNSUPPORTED":
		result.UnknownClass = callbackPreviewUnknownAmbiguous
		result.Step = "CALLBACK_SHAPE"
		result.NextOperation = "RESELECT_CALLBACK_SHAPE"
	case "CALLBACK_CAPTURE_UNSUPPORTED":
		result.UnknownClass = callbackPreviewUnknownAmbiguous
		result.Step = "CAPTURE_BINDING"
		result.NextOperation = "REVIEW_CAPTURE_BINDING"
	case "PENDING_TYPED_CALLBACK_EFFECTS":
		result.UnknownClass = callbackPreviewUnknownDependency
		result.Step = "PENDING_EFFECT_REVIEW"
		result.NextOperation = "RESOLVE_TYPED_CALLBACK_EFFECTS"
	case "CALLBACK_CANDIDATE_OVER_CAPACITY":
		result.UnknownClass = callbackPreviewUnknownUnbounded
		result.Step = "CAPACITY_GATE"
		result.NextOperation = "REMEASURE_BOUNDED_CANDIDATE"
	}
}

func callbackPreviewEvidence(result CallbackPreviewResult, candidate *CallbackPreviewCandidate, captures, effects int) CallbackPreviewEvidence {
	evidence := CallbackPreviewEvidence{State: result.State, CaptureCount: captures, PendingEffectCount: effects, OperationResultAdmission: result.OperationResultAdmission, ApplyPermission: result.ApplyPermission, Stage: result.Stage, Step: result.Step, Reason: result.Reason, UnknownClass: result.UnknownClass, NextOperation: result.NextOperation, BlockedBy: append([]string{}, result.BlockedBy...)}
	if candidate != nil {
		evidence.CandidateIdentity = candidate.CandidateIdentity
		evidence.SourceDigest = candidate.SourceDigest
		evidence.CandidateDigest = candidate.CandidateDigest
		evidence.HelperLines = candidate.HelperLines
		evidence.ParentFunctionLines = candidate.ParentFunctionLines
	} else {
		evidence.SourceDigest = result.SourceDigest
	}
	return evidence
}

func callbackPreviewRecords(contract generation.CallbackPreviewContractEvidence, result CallbackPreviewResult, candidate CallbackPreviewCandidate, captures []CallbackPreviewCapture, effects []CallbackPreviewEffect) ([]generation.CallbackPreviewRecord, error) {
	captureNames := make([]string, 0, len(captures))
	captureIdentities := make([]string, 0, len(captures))
	captureTypes := make([]string, 0, len(captures))
	captureModes := make([]string, 0, len(captures))
	for _, capture := range captures {
		captureNames = append(captureNames, capture.Name)
		captureIdentities = append(captureIdentities, capture.ObjectIdentity)
		captureTypes = append(captureTypes, capture.ObjectType)
		captureModes = append(captureModes, capture.BindingMode)
	}
	callIdentities := make([]string, 0, len(effects))
	symbols := make([]string, 0, len(effects))
	signatures := make([]string, 0, len(effects))
	receiverTypes := make([]string, 0, len(effects))
	effectKinds := make([]string, 0, len(effects))
	states := make([]string, 0, len(effects))
	effectStages := make([]string, 0, len(effects))
	effectSteps := make([]string, 0, len(effects))
	effectReasons := make([]string, 0, len(effects))
	effectUnknownClasses := make([]string, 0, len(effects))
	effectNextOperations := make([]string, 0, len(effects))
	effectBlockedBy := make([][]string, 0, len(effects))
	for _, effect := range effects {
		callIdentities = append(callIdentities, effect.CallIdentity)
		symbols = append(symbols, effect.Symbol)
		signatures = append(signatures, effect.Signature)
		receiverTypes = append(receiverTypes, effect.ReceiverType)
		effectKinds = append(effectKinds, effect.EffectKind)
		states = append(states, effect.State)
		effectStages = append(effectStages, effect.Stage)
		effectSteps = append(effectSteps, effect.Step)
		effectReasons = append(effectReasons, effect.Reason)
		effectUnknownClasses = append(effectUnknownClasses, effect.UnknownClass)
		effectNextOperations = append(effectNextOperations, effect.NextOperation)
		effectBlockedBy = append(effectBlockedBy, append([]string{}, effect.BlockedBy...))
	}
	encode := func(values []string) (string, error) { return generation.EncodeCallbackPreviewList(values) }
	encodedCaptureNames, err := encode(captureNames)
	if err != nil {
		return nil, err
	}
	encodedCaptureIdentities, err := encode(captureIdentities)
	if err != nil {
		return nil, err
	}
	encodedCaptureTypes, err := encode(captureTypes)
	if err != nil {
		return nil, err
	}
	encodedCaptureModes, err := encode(captureModes)
	if err != nil {
		return nil, err
	}
	encodedCallIdentities, err := encode(callIdentities)
	if err != nil {
		return nil, err
	}
	encodedSymbols, err := encode(symbols)
	if err != nil {
		return nil, err
	}
	encodedSignatures, err := encode(signatures)
	if err != nil {
		return nil, err
	}
	encodedReceiverTypes, err := encode(receiverTypes)
	if err != nil {
		return nil, err
	}
	encodedEffectKinds, err := encode(effectKinds)
	if err != nil {
		return nil, err
	}
	encodedStates, err := encode(states)
	if err != nil {
		return nil, err
	}
	candidateValues := map[string]string{
		"CandidateIdentity": candidate.CandidateIdentity, "SourceDigest": candidate.SourceDigest, "CandidateDigest": candidate.CandidateDigest,
		"HelperName": candidate.HelperName, "HelperBytes": strconv.Itoa(candidate.HelperBytes), "HelperLines": strconv.Itoa(candidate.HelperLines),
		"ParentFunctionLines": strconv.Itoa(candidate.ParentFunctionLines), "CaptureCount": strconv.Itoa(candidate.CaptureCount),
		"PendingEffectCount": strconv.Itoa(candidate.PendingEffectCount), "State": candidate.State, "Promotion": candidate.Promotion,
	}
	for name, value := range closurePreservationRecordValues(result.StructureProof) {
		candidateValues[name] = value
	}
	candidateRecord, err := contract.BuildCallbackPreviewRecord(contract.CandidateEntity, candidateValues)
	if err != nil {
		return nil, err
	}
	capturesRecord, err := contract.BuildCallbackPreviewRecord(contract.CapturesEntity, map[string]string{
		"CandidateIdentity": candidate.CandidateIdentity, "CaptureNames": encodedCaptureNames, "ObjectIdentities": encodedCaptureIdentities,
		"ObjectTypes": encodedCaptureTypes, "BindingModes": encodedCaptureModes, "Count": strconv.Itoa(len(captures)),
	})
	if err != nil {
		return nil, err
	}
	effectsRecord, err := contract.BuildCallbackPreviewRecord(contract.EffectsEntity, map[string]string{
		"CandidateIdentity": candidate.CandidateIdentity, "CallIdentities": encodedCallIdentities, "Symbols": encodedSymbols,
		"Signatures": encodedSignatures, "ReceiverTypes": encodedReceiverTypes, "EffectKinds": encodedEffectKinds,
		"States": encodedStates, "EffectStages": callbackPreviewEncodeList(effectStages), "EffectSteps": callbackPreviewEncodeList(effectSteps),
		"EffectReasons": callbackPreviewEncodeList(effectReasons), "EffectUnknownClasses": callbackPreviewEncodeList(effectUnknownClasses),
		"EffectNextOperations": callbackPreviewEncodeList(effectNextOperations), "EffectBlockedBy": callbackPreviewEncodeNestedList(effectBlockedBy),
		"Count": strconv.Itoa(len(effects)), "ResolvedCount": "0", "State": result.State,
	})
	if err != nil {
		return nil, err
	}
	evidenceRecord := callbackPreviewEvidenceRecord(contract, result, result.Evidence)
	records := []generation.CallbackPreviewRecord{result.ContractRecords[0], candidateRecord, capturesRecord, effectsRecord, evidenceRecord}
	if err := contract.ValidateCallbackPreviewFlow(records); err != nil {
		return nil, err
	}
	return records, nil
}

func callbackPreviewEvidenceRecord(contract generation.CallbackPreviewContractEvidence, result CallbackPreviewResult, evidence CallbackPreviewEvidence) generation.CallbackPreviewRecord {
	record, err := contract.BuildCallbackPreviewRecord(contract.EvidenceEntity, map[string]string{
		"CandidateIdentity": evidence.CandidateIdentity, "SourceDigest": evidence.SourceDigest, "CandidateDigest": evidence.CandidateDigest,
		"State": evidence.State, "CaptureCount": strconv.Itoa(evidence.CaptureCount), "PendingEffectCount": strconv.Itoa(evidence.PendingEffectCount),
		"ResolvedEffectCount": strconv.Itoa(evidence.ResolvedEffectCount), "HelperLines": strconv.Itoa(evidence.HelperLines),
		"ParentFunctionLines": strconv.Itoa(evidence.ParentFunctionLines), "OperationResultAdmission": evidence.OperationResultAdmission,
		"ApplyPermission": evidence.ApplyPermission, "Stage": evidence.Stage, "Step": evidence.Step, "Reason": evidence.Reason,
		"UnknownClass": evidence.UnknownClass, "NextOperation": evidence.NextOperation, "BlockedBy": callbackPreviewEncodeList(evidence.BlockedBy),
	})
	if err != nil {
		return generation.CallbackPreviewRecord{}
	}
	return record
}

func callbackPreviewEncodeList(values []string) string {
	encoded, err := generation.EncodeCallbackPreviewList(values)
	if err != nil {
		return generation.CallbackPreviewListCodecPrefix + "[]"
	}
	return encoded
}

func callbackPreviewEncodeNestedList(values [][]string) string {
	encoded, err := generation.EncodeCallbackPreviewNestedList(values)
	if err != nil {
		return generation.CallbackPreviewListCodecPrefix + "[]"
	}
	return encoded
}

func ValidateCallbackPreviewResult(result CallbackPreviewResult) error {
	contract, err := generation.LoadCallbackPreviewContract()
	if err != nil {
		return err
	}
	return validateCallbackPreviewResult(contract, result)
}

func validateCallbackPreviewResult(contract generation.CallbackPreviewContractEvidence, result CallbackPreviewResult) error {
	if result.LoweringStrategy != callbackPreviewWrapperLowering && result.LoweringStrategy != callbackPreviewFactoryLowering {
		return fmt.Errorf("callback preview lowering strategy is unknown")
	}
	if result.Schema != callbackPreviewSchema || result.State != callbackPreviewStateUnknown || result.OperationResultAdmission != "FORBIDDEN" || result.ApplyPermission != "FORBIDDEN" || result.Stage != callbackPreviewStage || result.Reason == "" || result.UnknownClass == "" || result.Step == "" || result.NextOperation == "" || result.BlockedBy == nil {
		return fmt.Errorf("callback preview result identity or admission is invalid")
	}
	expectedClass, knownClass := callbackPreviewUnknownClassForReason(result.Reason)
	if !knownClass || result.UnknownClass != expectedClass {
		return fmt.Errorf("callback preview unknown class is not bound to reason")
	}
	expectedStep, expectedNext := callbackPreviewLifecycleForReason(result.Reason)
	if result.Step != expectedStep || result.NextOperation != expectedNext {
		return fmt.Errorf("callback preview lifecycle step is not bound to reason")
	}
	if (expectedClass == callbackPreviewUnknownDirect || expectedClass == callbackPreviewUnknownAmbiguous) && len(result.BlockedBy) != 0 {
		return fmt.Errorf("callback preview direct or ambiguous unknown has a blocker frontier")
	}
	if result.ContractSourceDigest != contract.SourceDigest || result.ContractSemanticDigest != contract.SemanticDigest || result.Evidence.State != result.State || result.Evidence.OperationResultAdmission != result.OperationResultAdmission || result.Evidence.ApplyPermission != result.ApplyPermission || result.Evidence.Stage != result.Stage || result.Evidence.Step != result.Step || result.Evidence.Reason != result.Reason || result.Evidence.UnknownClass != result.UnknownClass || result.Evidence.NextOperation != result.NextOperation || !sameStringSlice(result.Evidence.BlockedBy, result.BlockedBy) {
		return fmt.Errorf("callback preview result lifecycle evidence is not bound")
	}
	if result.Candidate == nil {
		if len(result.Captures) != 0 || len(result.PendingEffects) != 0 {
			return fmt.Errorf("callback preview direct-unknown native collections are not empty")
		}
		if len(result.ContractRecords) != 2 {
			return fmt.Errorf("callback preview direct-unknown record flow is not bounded")
		}
		input, evidence := result.ContractRecords[0], result.ContractRecords[1]
		if input.Entity != contract.InputEntity || evidence.Entity != contract.EvidenceEntity {
			return fmt.Errorf("callback preview direct-unknown record entities are not bound")
		}
		if err := contract.ValidateCallbackPreviewRecord(input); err != nil {
			return err
		}
		if err := contract.ValidateCallbackPreviewRecord(evidence); err != nil {
			return err
		}
		inputValues := callbackPreviewRecordFieldMap(input)
		if inputValues["LoweringStrategy"] != result.LoweringStrategy {
			return fmt.Errorf("callback preview direct lowering input is not bound")
		}
		if inputValues["LogicalPath"] != result.LogicalPath || inputValues["Subject"] != result.Subject || inputValues["SourceDigest"] != result.SourceDigest || inputValues["State"] != result.State {
			return fmt.Errorf("callback preview direct input record is not bound")
		}
		evidenceValues := callbackPreviewRecordFieldMap(evidence)
		expectedEvidence := callbackPreviewEvidence(result, nil, 0, 0)
		if !callbackPreviewEvidenceMatches(result.Evidence, expectedEvidence) {
			return fmt.Errorf("callback preview direct native evidence is not canonical")
		}
		if evidenceValues["CandidateIdentity"] != expectedEvidence.CandidateIdentity || evidenceValues["SourceDigest"] != expectedEvidence.SourceDigest || evidenceValues["CandidateDigest"] != expectedEvidence.CandidateDigest || evidenceValues["State"] != expectedEvidence.State || evidenceValues["CaptureCount"] != "0" || evidenceValues["PendingEffectCount"] != "0" || evidenceValues["ResolvedEffectCount"] != "0" || evidenceValues["HelperLines"] != "0" || evidenceValues["ParentFunctionLines"] != "0" || evidenceValues["OperationResultAdmission"] != expectedEvidence.OperationResultAdmission || evidenceValues["ApplyPermission"] != expectedEvidence.ApplyPermission || evidenceValues["Stage"] != expectedEvidence.Stage || evidenceValues["Step"] != expectedEvidence.Step || evidenceValues["Reason"] != expectedEvidence.Reason || evidenceValues["UnknownClass"] != expectedEvidence.UnknownClass || evidenceValues["NextOperation"] != expectedEvidence.NextOperation {
			return fmt.Errorf("callback preview direct evidence record is not bound")
		}
		if err := validateCallbackPreviewListField(evidenceValues, "BlockedBy", result.BlockedBy); err != nil {
			return err
		}
		return nil
	}
	candidate := result.Candidate
	if err := validateClosurePreservationSummary(result); err != nil {
		return err
	}
	if err := validateCallbackPreviewLowering(result); err != nil {
		return err
	}
	if candidate.CandidateIdentity == "" || candidate.SourceDigest != result.SourceDigest || candidate.CandidateDigest == "" || callbackPreviewDigest([]byte(candidate.CandidateSource)) != candidate.CandidateDigest || candidate.HelperBytes != len([]byte(candidate.HelperSource)) || candidate.HelperLines != physicalLines([]byte(candidate.HelperSource)) || candidate.State != result.State || candidate.Promotion != callbackPreviewPromotionNone || candidate.CaptureCount != len(result.Captures) || candidate.PendingEffectCount != len(result.PendingEffects) || candidate.HelperLines <= 0 || candidate.HelperLines > functionLineLimit || candidate.ParentFunctionLines <= 0 || candidate.ParentFunctionLines > functionLineLimit {
		return fmt.Errorf("callback preview candidate does not match bounded result")
	}
	if result.Reason == "PENDING_TYPED_CALLBACK_EFFECTS" {
		if !sameStringSlice(result.BlockedBy, callbackPreviewEffectIdentities(result.PendingEffects)) {
			return fmt.Errorf("callback preview effect frontier is not bound")
		}
	} else if result.Reason == "CALLBACK_CANDIDATE_OVER_CAPACITY" && !sameStringSlice(result.BlockedBy, []string{candidate.CandidateIdentity}) {
		return fmt.Errorf("callback preview capacity frontier is not bound")
	}
	for _, effect := range result.PendingEffects {
		if effect.CallIdentity == "" || effect.State != callbackPreviewStateUnknown || effect.Stage != callbackPreviewStage || effect.Step != "PENDING_EFFECT_REVIEW" || effect.Reason != "PENDING_TYPED_CALLBACK_EFFECTS" || effect.UnknownClass != callbackPreviewUnknownDirect || effect.NextOperation != "RESTORE_TYPED_CALLBACK_EFFECT" || effect.BlockedBy == nil || len(effect.BlockedBy) != 0 || effect.StartOffset < 0 || effect.EndOffset < effect.StartOffset {
			return fmt.Errorf("callback preview pending effect %q is not bounded", effect.CallIdentity)
		}
	}
	if result.Evidence.CandidateIdentity != candidate.CandidateIdentity || result.Evidence.SourceDigest != candidate.SourceDigest || result.Evidence.CandidateDigest != candidate.CandidateDigest || result.Evidence.CaptureCount != len(result.Captures) || result.Evidence.PendingEffectCount != len(result.PendingEffects) || result.Evidence.ResolvedEffectCount != 0 || result.Evidence.HelperLines != candidate.HelperLines || result.Evidence.ParentFunctionLines != candidate.ParentFunctionLines {
		return fmt.Errorf("callback preview evidence counters or digest is not bound")
	}
	if len(result.ContractRecords) != 5 {
		return fmt.Errorf("callback preview contract record flow is incomplete")
	}
	if err := contract.ValidateCallbackPreviewFlow(result.ContractRecords); err != nil {
		return err
	}
	records := make(map[string]map[string]string, len(result.ContractRecords))
	for _, record := range result.ContractRecords {
		records[record.Entity] = callbackPreviewRecordFieldMap(record)
	}
	inputValues := records[contract.InputEntity]
	if inputValues["LoweringStrategy"] != result.LoweringStrategy {
		return fmt.Errorf("callback preview lowering input is not bound")
	}
	if inputValues["LogicalPath"] != result.LogicalPath || inputValues["Subject"] != result.Subject || inputValues["SourceDigest"] != result.SourceDigest || inputValues["State"] != result.State {
		return fmt.Errorf("callback preview input record is not bound")
	}
	candidateValues := records[contract.CandidateEntity]
	for name, expected := range closurePreservationRecordValues(result.StructureProof) {
		if candidateValues[name] != expected {
			return fmt.Errorf("callback preview source structure field %s is not bound", name)
		}
	}
	for name, expected := range map[string]string{
		"CandidateIdentity": candidate.CandidateIdentity, "SourceDigest": candidate.SourceDigest, "CandidateDigest": candidate.CandidateDigest,
		"HelperName": candidate.HelperName, "HelperBytes": strconv.Itoa(candidate.HelperBytes), "HelperLines": strconv.Itoa(candidate.HelperLines),
		"ParentFunctionLines": strconv.Itoa(candidate.ParentFunctionLines), "CaptureCount": strconv.Itoa(candidate.CaptureCount), "PendingEffectCount": strconv.Itoa(candidate.PendingEffectCount),
		"State": candidate.State, "Promotion": candidate.Promotion,
	} {
		if candidateValues[name] != expected {
			return fmt.Errorf("callback preview candidate record field %s is not bound", name)
		}
	}
	if err := validateCallbackPreviewListField(records[contract.CapturesEntity], "CaptureNames", callbackPreviewCaptureValues(result.Captures, func(c CallbackPreviewCapture) string { return c.Name })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.CapturesEntity], "ObjectIdentities", callbackPreviewCaptureValues(result.Captures, func(c CallbackPreviewCapture) string { return c.ObjectIdentity })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.CapturesEntity], "ObjectTypes", callbackPreviewCaptureValues(result.Captures, func(c CallbackPreviewCapture) string { return c.ObjectType })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.CapturesEntity], "BindingModes", callbackPreviewCaptureValues(result.Captures, func(c CallbackPreviewCapture) string { return c.BindingMode })); err != nil {
		return err
	}
	if records[contract.CapturesEntity]["CandidateIdentity"] != candidate.CandidateIdentity || records[contract.CapturesEntity]["Count"] != strconv.Itoa(len(result.Captures)) {
		return fmt.Errorf("callback preview capture record counters are not bound")
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "CallIdentities", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.CallIdentity })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "Symbols", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.Symbol })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "Signatures", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.Signature })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "ReceiverTypes", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.ReceiverType })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "EffectKinds", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.EffectKind })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "States", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.State })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "EffectStages", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.Stage })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "EffectSteps", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.Step })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "EffectReasons", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.Reason })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "EffectUnknownClasses", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.UnknownClass })); err != nil {
		return err
	}
	if err := validateCallbackPreviewListField(records[contract.EffectsEntity], "EffectNextOperations", callbackPreviewEffectValues(result.PendingEffects, func(e CallbackPreviewEffect) string { return e.NextOperation })); err != nil {
		return err
	}
	effectBlockedBy := make([][]string, 0, len(result.PendingEffects))
	for _, effect := range result.PendingEffects {
		effectBlockedBy = append(effectBlockedBy, effect.BlockedBy)
	}
	if err := validateCallbackPreviewNestedListField(records[contract.EffectsEntity], "EffectBlockedBy", effectBlockedBy); err != nil {
		return err
	}
	if records[contract.EffectsEntity]["CandidateIdentity"] != candidate.CandidateIdentity || records[contract.EffectsEntity]["Count"] != strconv.Itoa(len(result.PendingEffects)) || records[contract.EffectsEntity]["ResolvedCount"] != "0" || records[contract.EffectsEntity]["State"] != result.State {
		return fmt.Errorf("callback preview effect record counters are not bound")
	}
	evidenceValues := records[contract.EvidenceEntity]
	if evidenceValues["CandidateIdentity"] != result.Evidence.CandidateIdentity || evidenceValues["SourceDigest"] != result.Evidence.SourceDigest || evidenceValues["CandidateDigest"] != result.Evidence.CandidateDigest || evidenceValues["State"] != result.Evidence.State || evidenceValues["CaptureCount"] != strconv.Itoa(result.Evidence.CaptureCount) || evidenceValues["PendingEffectCount"] != strconv.Itoa(result.Evidence.PendingEffectCount) || evidenceValues["ResolvedEffectCount"] != strconv.Itoa(result.Evidence.ResolvedEffectCount) || evidenceValues["HelperLines"] != strconv.Itoa(result.Evidence.HelperLines) || evidenceValues["ParentFunctionLines"] != strconv.Itoa(result.Evidence.ParentFunctionLines) || evidenceValues["OperationResultAdmission"] != result.Evidence.OperationResultAdmission || evidenceValues["ApplyPermission"] != result.Evidence.ApplyPermission || evidenceValues["Stage"] != result.Evidence.Stage || evidenceValues["Step"] != result.Evidence.Step || evidenceValues["Reason"] != result.Evidence.Reason || evidenceValues["UnknownClass"] != result.Evidence.UnknownClass || evidenceValues["NextOperation"] != result.Evidence.NextOperation {
		return fmt.Errorf("callback preview evidence record is not bound")
	}
	if err := validateCallbackPreviewListField(evidenceValues, "BlockedBy", result.BlockedBy); err != nil {
		return err
	}
	return nil
}

func callbackPreviewEvidenceMatches(actual, expected CallbackPreviewEvidence) bool {
	return actual.CandidateIdentity == expected.CandidateIdentity && actual.SourceDigest == expected.SourceDigest && actual.CandidateDigest == expected.CandidateDigest && actual.State == expected.State && actual.CaptureCount == expected.CaptureCount && actual.PendingEffectCount == expected.PendingEffectCount && actual.ResolvedEffectCount == expected.ResolvedEffectCount && actual.HelperLines == expected.HelperLines && actual.ParentFunctionLines == expected.ParentFunctionLines && actual.OperationResultAdmission == expected.OperationResultAdmission && actual.ApplyPermission == expected.ApplyPermission && actual.Stage == expected.Stage && actual.Step == expected.Step && actual.Reason == expected.Reason && actual.UnknownClass == expected.UnknownClass && actual.NextOperation == expected.NextOperation && (actual.BlockedBy == nil) == (expected.BlockedBy == nil) && sameStringSlice(actual.BlockedBy, expected.BlockedBy)
}

func callbackPreviewUnknownClassForReason(reason string) (string, bool) {
	switch reason {
	case "CALLBACK_TARGET_MISSING", "TYPE_EVIDENCE_MISSING":
		return callbackPreviewUnknownDirect, true
	case "CALLBACK_TARGET_SHAPE_UNSUPPORTED", "CALLBACK_CAPTURE_UNSUPPORTED":
		return callbackPreviewUnknownAmbiguous, true
	case "PENDING_TYPED_CALLBACK_EFFECTS":
		return callbackPreviewUnknownDependency, true
	case "CALLBACK_CANDIDATE_OVER_CAPACITY":
		return callbackPreviewUnknownUnbounded, true
	default:
		return "", false
	}
}

func callbackPreviewLifecycleForReason(reason string) (string, string) {
	switch reason {
	case "TYPE_EVIDENCE_MISSING":
		return "TYPE_EVIDENCE", "RECHECK_TYPE_EVIDENCE"
	case "CALLBACK_TARGET_SHAPE_UNSUPPORTED":
		return "CALLBACK_SHAPE", "RESELECT_CALLBACK_SHAPE"
	case "CALLBACK_CAPTURE_UNSUPPORTED":
		return "CAPTURE_BINDING", "REVIEW_CAPTURE_BINDING"
	case "PENDING_TYPED_CALLBACK_EFFECTS":
		return "PENDING_EFFECT_REVIEW", "RESOLVE_TYPED_CALLBACK_EFFECTS"
	case "CALLBACK_CANDIDATE_OVER_CAPACITY":
		return "CAPACITY_GATE", "REMEASURE_BOUNDED_CANDIDATE"
	default:
		return "TARGET_DISCOVERY", "REVIEW_CALLBACK_TARGET"
	}
}

func validateCallbackPreviewListField(values map[string]string, name string, expected []string) error {
	encoded := values[name]
	if err := generation.ValidateCallbackPreviewList(encoded); err != nil {
		return fmt.Errorf("callback preview record field %s: %w", name, err)
	}
	actual := []string{}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(encoded, generation.CallbackPreviewListCodecPrefix)), &actual); err != nil {
		return err
	}
	if !sameStringSlice(actual, expected) {
		return fmt.Errorf("callback preview record field %s does not match native values", name)
	}
	return nil
}

func validateCallbackPreviewNestedListField(values map[string]string, name string, expected [][]string) error {
	encoded := values[name]
	if err := generation.ValidateCallbackPreviewNestedList(encoded); err != nil {
		return fmt.Errorf("callback preview record field %s: %w", name, err)
	}
	var actual [][]string
	if err := json.Unmarshal([]byte(strings.TrimPrefix(encoded, generation.CallbackPreviewListCodecPrefix)), &actual); err != nil {
		return err
	}
	if len(actual) != len(expected) {
		return fmt.Errorf("callback preview record field %s does not match native values", name)
	}
	for index := range actual {
		if !sameStringSlice(actual[index], expected[index]) {
			return fmt.Errorf("callback preview record field %s does not match native values", name)
		}
	}
	return nil
}

func callbackPreviewRecordFieldMap(record generation.CallbackPreviewRecord) map[string]string {
	values := make(map[string]string, len(record.Fields))
	for _, field := range record.Fields {
		values[field.Name] = field.Value
	}
	return values
}

func callbackPreviewCaptureValues(captures []CallbackPreviewCapture, value func(CallbackPreviewCapture) string) []string {
	values := make([]string, 0, len(captures))
	for _, capture := range captures {
		values = append(values, value(capture))
	}
	return values
}

func callbackPreviewEffectValues(effects []CallbackPreviewEffect, value func(CallbackPreviewEffect) string) []string {
	values := make([]string, 0, len(effects))
	for _, effect := range effects {
		values = append(values, value(effect))
	}
	return values
}

func sameStringSlice(left, right []string) bool {
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

func callbackPreviewFunction(file *ast.File, name string) *ast.FuncDecl {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name != nil && function.Name.Name == name {
			return function
		}
	}
	return nil
}

func callbackPreviewFuncLit(function *ast.FuncDecl, info *types.Info) (*ast.FuncLit, error) {
	var found *ast.FuncLit
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.X == nil {
			return true
		}
		object, ok := info.Uses[selector.Sel].(*types.TypeName)
		if !ok || object.Pkg() == nil || object.Pkg().Path() != "net/http" || object.Name() != "HandlerFunc" {
			return true
		}
		literal, ok := call.Args[0].(*ast.FuncLit)
		if !ok || literal.Type == nil || literal.Type.Params == nil || len(literal.Type.Params.List) != 2 || literal.Type.Results != nil {
			return true
		}
		if !callbackPreviewHTTPParam(literal.Type.Params.List[0].Type, info, "ResponseWriter") || !callbackPreviewHTTPRequestParam(literal.Type.Params.List[1].Type, info) {
			return true
		}
		found = literal
		return false
	})
	if found == nil {
		return nil, fmt.Errorf("CALLBACK_TARGET_SHAPE_UNSUPPORTED")
	}
	return found, nil
}

func callbackPreviewHTTPParam(expression ast.Expr, info *types.Info, name string) bool {
	typeValue := info.TypeOf(expression)
	named, ok := typeValue.(*types.Named)
	return ok && named.Obj() != nil && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "net/http" && named.Obj().Name() == name
}

func callbackPreviewHTTPRequestParam(expression ast.Expr, info *types.Info) bool {
	pointer, ok := info.TypeOf(expression).(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := pointer.Elem().(*types.Named)
	return ok && named.Obj() != nil && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "net/http" && named.Obj().Name() == "Request"
}

func callbackPreviewCaptures(callback *ast.FuncLit, evidence typeEvidence, fset *token.FileSet) ([]CallbackPreviewCapture, error) {
	local := make(map[types.Object]bool)
	selectorNames := make(map[*ast.Ident]bool)
	ast.Inspect(callback.Type, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok {
			if object := evidence.info.Defs[identifier]; object != nil {
				local[object] = true
			}
		}
		return true
	})
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.Ident:
			if object := evidence.info.Defs[value]; object != nil {
				local[object] = true
			}
		case *ast.SelectorExpr:
			selectorNames[value.Sel] = true
		}
		return true
	})
	objects := make(map[types.Object]bool)
	var unsupported types.Object
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if !ok || selectorNames[identifier] {
			return true
		}
		object := evidence.info.Uses[identifier]
		if object == nil || local[object] {
			return true
		}
		if object.Type() == nil {
			return true
		}
		if basic, ok := object.Type().Underlying().(*types.Basic); ok && basic.Kind() == types.UntypedNil && object.Pkg() == nil {
			return true
		}
		if _, ok := object.(*types.TypeName); ok {
			return true
		}
		if _, ok := object.(*types.PkgName); ok {
			return true
		}
		if variable, ok := object.(*types.Var); ok {
			if variable.Parent() == evidence.pkg.Scope() {
				return true
			}
			objects[object] = true
			return true
		}
		if object.Pkg() == nil {
			return true
		}
		unsupported = object
		objects[object] = true
		return true
	})
	if unsupported != nil {
		return nil, fmt.Errorf("unsupported non-variable callback capture %s", unsupported.Name())
	}
	captures := make([]CallbackPreviewCapture, 0, len(objects))
	for object := range objects {
		mode := "pointer-identity"
		captures = append(captures, CallbackPreviewCapture{Name: object.Name(), ObjectIdentity: callbackPreviewObjectIdentity(object, fset), ObjectType: callbackPreviewTypeString(object.Type(), evidence.pkg), BindingMode: mode})
	}
	sort.Slice(captures, func(left, right int) bool { return captures[left].Name < captures[right].Name })
	return captures, nil
}

func callbackPreviewEffects(callback *ast.FuncLit, evidence typeEvidence, fset *token.FileSet) []CallbackPreviewEffect {
	effects := make([]CallbackPreviewEffect, 0)
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		object, receiverType, kind := callbackPreviewCallIdentity(call, evidence)
		if object == nil && kind == "typed-conversion" {
			return true
		}
		if object == nil && kind == "known-pure-conversion" {
			return true
		}
		symbol, signature := "<unknown>", "<unknown>"
		if object != nil {
			symbol, signature = callbackPreviewObjectIdentity(object, fset), callbackPreviewTypeString(object.Type(), evidence.pkg)
		}
		start := fset.Position(call.Pos()).Offset
		end := fset.Position(call.End()).Offset
		effectKind := kind
		if callbackPreviewFrameSensitive(object) {
			effectKind = "caller-frame-sensitive"
		}
		if effectKind == "" {
			effectKind = "unresolved-callee-effect"
		}
		effects = append(effects, callbackPreviewPendingEffect(fmt.Sprintf("%s@%d:%d", symbol, start, end), symbol, signature, receiverType, effectKind, start, end))
		return true
	})
	ast.Inspect(callback.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.FuncLit:
			effects = append(effects, callbackPreviewPendingEffect(fmt.Sprintf("nested-func-lit@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), "<func-lit>", "<unknown>", "", "nested-function-retention", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset))
		case *ast.GoStmt:
			effects = append(effects, callbackPreviewPendingEffect(fmt.Sprintf("go@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), "<go>", "", "", "async-retention", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset))
		case *ast.DeferStmt:
			effects = append(effects, callbackPreviewPendingEffect(fmt.Sprintf("defer@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), "<defer>", "", "", "deferred-effect", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset))
		case *ast.RangeStmt:
			if rangeType := evidence.info.TypeOf(value.X); rangeType != nil {
				if _, ok := rangeType.Underlying().(*types.Signature); ok {
					effects = append(effects, callbackPreviewPendingEffect(fmt.Sprintf("range-func@%d:%d", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset), "<range-func>", callbackPreviewTypeString(rangeType, evidence.pkg), "", "function-iterator", fset.Position(value.Pos()).Offset, fset.Position(value.End()).Offset))
				}
			}
		}
		return true
	})
	sort.Slice(effects, func(left, right int) bool { return effects[left].StartOffset < effects[right].StartOffset })
	return effects
}

func callbackPreviewPendingEffect(callIdentity, symbol, signature, receiverType, effectKind string, start, end int) CallbackPreviewEffect {
	return CallbackPreviewEffect{CallIdentity: callIdentity, Symbol: symbol, Signature: signature, ReceiverType: receiverType, EffectKind: effectKind, State: callbackPreviewStateUnknown, StartOffset: start, EndOffset: end, Stage: callbackPreviewStage, Step: "PENDING_EFFECT_REVIEW", Reason: "PENDING_TYPED_CALLBACK_EFFECTS", UnknownClass: callbackPreviewUnknownDirect, NextOperation: "RESTORE_TYPED_CALLBACK_EFFECT", BlockedBy: []string{}}
}

func callbackPreviewEffectIdentities(effects []CallbackPreviewEffect) []string {
	identities := make([]string, 0, len(effects))
	for _, effect := range effects {
		identities = append(identities, effect.CallIdentity)
	}
	return identities
}

func callbackPreviewCallIdentity(call *ast.CallExpr, evidence typeEvidence) (types.Object, string, string) {
	if call == nil || evidence.info == nil {
		return nil, "", "unknown-call"
	}
	if tv, ok := evidence.info.Types[call.Fun]; ok && tv.IsType() {
		return nil, "", "typed-conversion"
	}
	switch function := call.Fun.(type) {
	case *ast.Ident:
		object := evidence.info.Uses[function]
		if object == nil {
			if evidence.info.TypeOf(function) != nil {
				return nil, callbackPreviewTypeString(evidence.info.TypeOf(function), evidence.pkg), "function-valued-call"
			}
			return nil, "", "unknown-call"
		}
		return object, "", callbackPreviewObjectKind(object, call, evidence)
	case *ast.SelectorExpr:
		if selection := evidence.info.Selections[function]; selection != nil {
			return selection.Obj(), callbackPreviewTypeString(selection.Recv(), evidence.pkg), callbackPreviewObjectKind(selection.Obj(), call, evidence)
		}
		return evidence.info.Uses[function.Sel], "", callbackPreviewObjectKind(evidence.info.Uses[function.Sel], call, evidence)
	default:
		return nil, "", "function-valued-call"
	}
}

func callbackPreviewObjectKind(object types.Object, call *ast.CallExpr, evidence typeEvidence) string {
	if object == nil {
		return "unknown-call"
	}
	if _, ok := object.(*types.Var); ok {
		if signature, ok := object.Type().Underlying().(*types.Signature); ok && signature != nil {
			return "function-valued-call"
		}
	}
	if function, ok := object.(*types.Func); ok {
		if function.Pkg() != nil && function.Pkg().Path() == "net/http" && (function.Name() == "Error" || function.Name() == "NotFound") {
			return "dynamic-interface-argument"
		}
		if selection := callbackPreviewSelection(call, evidence.info); selection != nil {
			if _, ok := selection.Recv().Underlying().(*types.Interface); ok {
				return "dynamic-interface-method"
			}
			return "typed-method"
		}
		if function.Pkg() != nil && evidence.pkg != nil && function.Pkg() != evidence.pkg {
			return "external-function"
		}
		return "local-function"
	}
	return "unresolved-callee-effect"
}

func callbackPreviewSelection(call *ast.CallExpr, info *types.Info) *types.Selection {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || info == nil {
		return nil
	}
	return info.Selections[selector]
}

func callbackPreviewFrameSensitive(object types.Object) bool {
	function, ok := object.(*types.Func)
	return ok && function.Pkg() != nil && (function.Pkg().Path() == "runtime" || function.Pkg().Path() == "runtime/debug") && (function.Name() == "Caller" || function.Name() == "Callers" || function.Name() == "CallersFrames" || function.Name() == "Stack")
}

func callbackPreviewCandidate(source []byte, logical string, fset *token.FileSet, file *ast.File, target *ast.FuncDecl, callback *ast.FuncLit, captures []CallbackPreviewCapture, effects []CallbackPreviewEffect) (CallbackPreviewCandidate, error) {
	usedNames := make(map[string]bool)
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name != nil {
			usedNames[function.Name.Name] = true
		}
	}
	helperName := "boundedCallbackPreview" + target.Name.Name
	if usedNames[helperName] {
		return CallbackPreviewCandidate{}, fmt.Errorf("callback preview helper name collides: %s", helperName)
	}
	params := make([]*ast.Field, 0, len(callback.Type.Params.List)+len(captures))
	for _, field := range callback.Type.Params.List {
		params = append(params, field)
	}
	for _, capture := range captures {
		expression, err := parser.ParseExpr("*" + capture.ObjectType)
		if err != nil {
			return CallbackPreviewCandidate{}, fmt.Errorf("parse capture type %s: %w", capture.ObjectType, err)
		}
		params = append(params, &ast.Field{Names: []*ast.Ident{ast.NewIdent(capture.Name)}, Type: expression})
	}
	helper := &ast.FuncDecl{Name: ast.NewIdent(helperName), Type: &ast.FuncType{Params: &ast.FieldList{List: params}}, Body: callback.Body}
	var helperBuffer bytes.Buffer
	if err := format.Node(&helperBuffer, fset, helper); err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback helper: %w", err)
	}
	wrapper := &ast.FuncLit{Type: callback.Type, Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent(helperName), Args: callbackPreviewCallArguments(callback, captures)}}}}}
	var wrapperBuffer bytes.Buffer
	if err := format.Node(&wrapperBuffer, fset, wrapper); err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback wrapper: %w", err)
	}
	start := fset.Position(callback.Pos()).Offset
	end := fset.Position(callback.End()).Offset
	if start < 0 || end < start || end > len(source) {
		return CallbackPreviewCandidate{}, fmt.Errorf("callback preview source span is invalid")
	}
	modified := make([]byte, 0, len(source)+len(helperBuffer.Bytes())+len(wrapperBuffer.Bytes()))
	modified = append(modified, source[:start]...)
	modified = append(modified, wrapperBuffer.Bytes()...)
	modified = append(modified, source[end:]...)
	modified = append(modified, '\n', '\n')
	modified = append(modified, helperBuffer.Bytes()...)
	formatted, err := format.Source(modified)
	if err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("format callback candidate: %w", err)
	}
	candidateFset := token.NewFileSet()
	candidateFile, err := parser.ParseFile(candidateFset, "callback-preview.go", formatted, parser.ParseComments)
	if err != nil {
		return CallbackPreviewCandidate{}, fmt.Errorf("parse callback candidate: %w", err)
	}
	parentLines := 0
	if candidateTarget := callbackPreviewFunction(candidateFile, target.Name.Name); candidateTarget != nil {
		parentLines = candidateFset.Position(candidateTarget.End()).Line - candidateFset.Position(candidateTarget.Pos()).Line + 1
	}
	helperLines := physicalLines(helperBuffer.Bytes())
	identity := fmt.Sprintf("%s#%s@%d:%d", filepath.ToSlash(logical), target.Name.Name, start, end)
	candidateDigest := callbackPreviewDigest(formatted)
	return CallbackPreviewCandidate{CandidateIdentity: identity, SourceDigest: callbackPreviewDigest(source), CandidateDigest: candidateDigest, HelperName: helperName, WrapperSource: wrapperBuffer.String(), HelperSource: helperBuffer.String(), CandidateSource: string(formatted), HelperBytes: len(helperBuffer.Bytes()), HelperLines: helperLines, ParentFunctionLines: parentLines, CaptureCount: len(captures), PendingEffectCount: len(effects), State: callbackPreviewStateUnknown, Promotion: callbackPreviewPromotionNone}, nil
}

func callbackPreviewCallArguments(callback *ast.FuncLit, captures []CallbackPreviewCapture) []ast.Expr {
	arguments := make([]ast.Expr, 0, 2+len(captures))
	for _, field := range callback.Type.Params.List {
		for _, name := range field.Names {
			arguments = append(arguments, ast.NewIdent(name.Name))
		}
	}
	for _, capture := range captures {
		arguments = append(arguments, &ast.UnaryExpr{Op: token.AND, X: ast.NewIdent(capture.Name)})
	}
	return arguments
}

func callbackPreviewObjectIdentity(object types.Object, fset *token.FileSet) string {
	if object == nil {
		return ""
	}
	position := fset.Position(object.Pos())
	packagePath := ""
	if object.Pkg() != nil {
		packagePath = object.Pkg().Path()
	}
	return fmt.Sprintf("%s:%s:%d:%d", packagePath, object.Name(), position.Line, position.Column)
}

func callbackPreviewTypeString(typeValue types.Type, current *types.Package) string {
	return types.TypeString(typeValue, func(pkg *types.Package) string {
		if pkg == nil {
			return ""
		}
		if current != nil && pkg.Path() == current.Path() {
			return ""
		}
		return pkg.Name()
	})
}

func callbackPreviewDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + fmt.Sprintf("%x", sum[:])
}
