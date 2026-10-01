package bodycodegen

import (
	"context"
	"fmt"
	"time"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// RetainedModelInfo records one constructor load, never a per-request load.
type RetainedModelInfo struct {
	Schema              string  `json:"schema"`
	Loaded              bool    `json:"loaded"`
	MetadataSHA256      string  `json:"metadata_sha256,omitempty"`
	WeightsSHA256       string  `json:"weights_sha256,omitempty"`
	ResidentTensorBytes int     `json:"resident_tensor_bytes"`
	SetupMS             float64 `json:"setup_ms"`
	Scope               string  `json:"scope"`
}

type typedPathModel struct {
	path      string
	model     *decision.Model
	retention *RetainedModelInfo
	diagnosis *PathDiagnosisOptions
}

// TypedPathGenerator shares immutable model arrays. Every Generate call owns
// fresh preparation, source binding, session, workspace, feedback and receipts.
// Callers must not mutate their source/document/options during a call.
type TypedPathGenerator struct {
	model *decision.Model
	info  RetainedModelInfo
}

type TypedPathOptions struct {
	StepAttempts    int                   `json:"step_attempts,omitempty"`
	FeedbackRounds  int                   `json:"feedback_rounds,omitempty"`
	FeedbackUnfixed bool                  `json:"feedback_unfixed,omitempty"`
	CI              *pathplan.CIHint      `json:"ci,omitempty"`
	Diagnosis       *PathDiagnosisOptions `json:"diagnosis,omitempty"`
}

// NewTypedPathGenerator loads an optional explicit local path model once.
// An empty path constructs the disconnected deterministic generator.
func NewTypedPathGenerator(modelPath string) (*TypedPathGenerator, error) {
	started := time.Now()
	g := &TypedPathGenerator{info: RetainedModelInfo{Schema: "gooo/retained-path-model/v1",
		Scope: "one constructor load; excluded from request timing; fresh source and plan each request"}}
	if modelPath != "" {
		model, err := decision.LoadPath(modelPath)
		if err != nil {
			return nil, fmt.Errorf("load retained structural model: %w", err)
		}
		if model.Schema() != decision.PathMetadataSchema {
			return nil, fmt.Errorf("retained model must use the structural path ABI")
		}
		g.model = model
		g.info.Loaded = true
		g.info.MetadataSHA256 = model.MetadataSHA256()
		g.info.WeightsSHA256 = model.WeightsSHA256()
		g.info.ResidentTensorBytes = model.ResidentTensorBytes()
	}
	g.info.SetupMS = elapsedMS(started)
	return g, nil
}

func (g *TypedPathGenerator) Info() RetainedModelInfo { return g.info }

func (g *TypedPathGenerator) Generate(ctx context.Context, filename string, source []byte, activityName string,
	document pathplan.Document, options TypedPathOptions) (Result, error) {
	if g == nil || g.info.Schema != "gooo/retained-path-model/v1" {
		return Result{}, fmt.Errorf("generator required")
	}
	feedback, diagnosis, err := resolveTypedPathOptions(options, g.model != nil)
	if err != nil {
		return Result{}, err
	}
	return generateTypedPathRequest(ctx, filename, source, activityName, document,
		typedPathModel{model: g.model, retention: &g.info, diagnosis: diagnosis}, options.StepAttempts, feedback)
}
