package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

const jointSchema = "gooo/joint-construction/v1"
const jointMixedSchema = "gooo/joint-construction/v2"
const jointRejectionSchema = "gooo/joint-construction/v3"
const jointFillSchema = "gooo/joint-construction/v4"
const jointFillRejectionSchema = "gooo/joint-construction/v5"
const jointFaultSchema = "gooo/joint-construction/v6"
const jointPathSchema = "gooo/joint-construction/v7"
const jointPathHoldoutSchema = "gooo/joint-construction/v8"
const jointRecordHoldoutSchema = "gooo/joint-construction/v9"
const jointCriteriaSchema = "gooo/joint-construction/v10"

// JointConstruction keeps local preparation, caller feedback and subsequent
// evaluation distinct. Initial is also the source-bound historical model order.
type JointConstruction struct {
	FinalCriteriaHoldouts []JointCriteriaHoldout `json:"final_criteria_holdouts,omitempty"`
	FinalRecordHoldouts   []JointRecordHoldout   `json:"final_record_holdouts,omitempty"`
	Schema                string                 `json:"schema"`
	InputObservation      string                 `json:"input_observation,omitempty"`
	Stage                 string                 `json:"stage"`
	Failure               string                 `json:"failure,omitempty"`
	OriginalSourceSHA256  string                 `json:"original_source_sha256"`
	ConstructionCases     CompositionCases       `json:"construction_cases"`
	ConstructionSHA256    string                 `json:"construction_sha256"`
	Initial               Composition            `json:"initial"`
	ProgramBudget         int                    `json:"program_budget"`
	CandidateSpace        string                 `json:"candidate_space"`
	CandidateKinds        []string               `json:"candidate_kinds,omitempty"`
	Attempts              []JointAttempt         `json:"attempts"`
	SelectedAttempt       int                    `json:"selected_attempt"`
	SelectedSource        string                 `json:"selected_source"`
	Selected              Composition            `json:"selected"`
	FinalPathHoldouts     []JointPathHoldout     `json:"final_path_holdouts,omitempty"`
	Decision              string                 `json:"decision"`
	StopReason            string                 `json:"stop_reason"`
	ElapsedNS             int64                  `json:"elapsed_ns"`
	Scope                 string                 `json:"scope"`
}

type JointAttempt struct {
	CriteriaPassed   int                           `json:"criteria_passed,omitempty"`
	CriteriaFailed   int                           `json:"criteria_failed,omitempty"`
	CriteriaUnknown  int                           `json:"criteria_unknown,omitempty"`
	CriteriaTotal    int                           `json:"criteria_total,omitempty"`
	Masks            []uint16                      `json:"masks"`
	Rejection        *JointCandidateRejection      `json:"rejection,omitempty"`
	Candidates       []bodycodegen.RecordCandidate `json:"candidates"`
	SearchCandidates []bodycodegen.SearchCandidate `json:"search_candidates,omitempty"`
	FillCandidates   []bodycodegen.FillCandidate   `json:"fill_candidates,omitempty"`
	PathCandidates   []bodycodegen.PathCandidate   `json:"path_candidates,omitempty"`
	LocalPassed      int                           `json:"local_passed"`
	LocalTotal       int                           `json:"local_total"`
	Runtime          CompositionRuntime            `json:"runtime"`
}

// A rejected expression consumed a program attempt but did not reach native
// caller execution. Local counts cover only the scored prefix before this slot.
type JointCandidateRejection struct {
	Stage       string `json:"stage"`
	Slot        int    `json:"slot"`
	Activity    string `json:"activity"`
	CandidateID string `json:"candidate_id"`
	Reason      string `json:"reason"`
}

type JointOptions struct {
	EntryActivity string
	ModelPath     string
	FillModelPath string
	ProgramBudget int
	GoBinary      string
}

type JointEvaluation struct {
	Runtime              CompositionRuntime   `json:"runtime"`
	ReplayFailure        *JointReplayFailure  `json:"replay_failure,omitempty"`
	ConstructionReplayed bool                 `json:"construction_replayed"`
	NewModelCalls        int                  `json:"new_model_calls"`
	InputSeparation      JointInputSeparation `json:"input_separation"`
}

// JointReplayFailure is a current observation of a saved attempt. It never
// replaces the saved history or claims that the new evaluation suite ran.
type JointReplayFailure struct {
	AttemptIndex int                `json:"attempt_index"`
	Stage        string             `json:"stage"`
	Runtime      CompositionRuntime `json:"runtime"`
}

// This metric describes caller root tuples, including duplicates. It makes no
// claim about model-training exposure or all indirect local-example inputs.
type JointInputSeparation struct {
	UniqueInputs       int    `json:"unique_inputs"`
	DuplicateRows      int    `json:"duplicate_rows"`
	ConstructionInputs int    `json:"construction_inputs"`
	OtherInputs        int    `json:"other_inputs"`
	Scope              string `json:"scope"`
}
