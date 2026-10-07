// Package bodyrefinement runs source-owned Gooo feedback policies between
// bounded construction attempts and retains every native observation.
package bodyrefinement

import "github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"

type Options struct {
	Activity     string
	MaxAttempts  int
	MaxRounds    int
	ModelPath    string
	GoBinary     string
	PolicySource []byte
	Evaluation   *bodyexecution.CompositionCases
}

type Observation struct {
	Matched         int `json:"matched"`
	Total           int `json:"total"`
	Best            int `json:"best"`
	Attempts        int `json:"attempts"`
	Limit           int `json:"limit"`
	Round           int `json:"round"`
	RoundLimit      int `json:"round_limit"`
	Counterexamples int `json:"counterexamples"`
}

type Decision struct {
	Action       string `json:"action"`
	NextAttempts int    `json:"next_attempts"`
	Retain       bool   `json:"retain"`
	Reason       string `json:"reason"`
}

type Round struct {
	Source        string                           `json:"source"`
	Composition   bodyexecution.Composition        `json:"composition"`
	Runtime       bodyexecution.CompositionRuntime `json:"runtime"`
	Observation   Observation                      `json:"observation"`
	PolicyRuntime bodyexecution.CompositionRuntime `json:"policy_runtime"`
	Decision      Decision                         `json:"decision"`
	FeedbackAdded int                              `json:"feedback_added"`
	RevisedSource string                           `json:"revised_source,omitempty"`
}

type Result struct {
	Schema            string                            `json:"schema"`
	Status            string                            `json:"status"`
	FeedbackStatus    string                            `json:"feedback_status"`
	EvaluationStatus  string                            `json:"evaluation_status"`
	StopReason        string                            `json:"stop_reason"`
	Failure           string                            `json:"failure,omitempty"`
	Activity          string                            `json:"activity"`
	OriginalSource    string                            `json:"original_source"`
	PolicySource      string                            `json:"policy_source"`
	PolicyComposition bodyexecution.Composition         `json:"policy_composition"`
	FeedbackCases     bodyexecution.CompositionCases    `json:"feedback_cases"`
	EvaluationCases   *bodyexecution.CompositionCases   `json:"evaluation_cases,omitempty"`
	Rounds            []Round                           `json:"rounds"`
	SelectedRound     int                               `json:"selected_round"`
	Evaluation        *bodyexecution.CompositionRuntime `json:"evaluation,omitempty"`
	Scope             string                            `json:"scope"`
}
