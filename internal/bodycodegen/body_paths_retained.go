package bodycodegen

import (
	"context"
	"fmt"
	"time"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
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
	ModelSchema         string  `json:"model_schema,omitempty"`
	FeatureVersion      string  `json:"feature_version,omitempty"`
	ArithmeticVersion   string  `json:"arithmetic_version,omitempty"`
}

type typedPathModel struct {
	path        string
	model       *decision.Model
	joint       *jointdecision.Model
	three       *jointdecision.ThreeModel
	retention   *RetainedModelInfo
	diagnosis   *PathDiagnosisOptions
	observation *PathObservationOptions
}

// TypedPathGenerator shares immutable model arrays. Every Generate call owns
// fresh preparation, source binding, session, workspace, feedback and receipts.
// Callers must not mutate their source/document/options during a call.
type TypedPathGenerator struct {
	model *decision.Model
	joint *jointdecision.Model
	three *jointdecision.ThreeModel
	info  RetainedModelInfo
}

type TypedPathOptions struct {
	StepAttempts    int                     `json:"step_attempts,omitempty"`
	FeedbackRounds  int                     `json:"feedback_rounds,omitempty"`
	FeedbackUnfixed bool                    `json:"feedback_unfixed,omitempty"`
	CI              *pathplan.CIHint        `json:"ci,omitempty"`
	Diagnosis       *PathDiagnosisOptions   `json:"diagnosis,omitempty"`
	Observation     *PathObservationOptions `json:"observation,omitempty"`
}

// NewTypedPathGenerator loads an optional explicit local path model once.
// An empty path constructs the disconnected deterministic generator.
func NewTypedPathGenerator(modelPath string) (*TypedPathGenerator, error) {
	started := time.Now()
	g := &TypedPathGenerator{info: RetainedModelInfo{Schema: "gooo/retained-path-model/v1",
		Scope: "one constructor load; excluded from request timing; fresh source and plan each request"}}
	if modelPath != "" {
		models, err := loadTypedStructuralModel(modelPath)
		if err != nil {
			return nil, fmt.Errorf("load retained structural model: %w", err)
		}
		g.model, g.joint, g.three = models.model, models.joint, models.three
		g.info.Loaded = true
		if g.three != nil {
			g.info.MetadataSHA256, g.info.WeightsSHA256 = g.three.MetadataSHA256(), g.three.WeightsSHA256()
			g.info.ResidentTensorBytes = g.three.ResidentTensorBytes()
			g.info.ModelSchema, g.info.FeatureVersion = g.three.Schema(), g.three.FeatureVersion()
			g.info.ArithmeticVersion = g.three.ArithmeticVersion()
		} else if g.joint != nil {
			g.info.MetadataSHA256, g.info.WeightsSHA256 = g.joint.MetadataSHA256(), g.joint.WeightsSHA256()
			g.info.ResidentTensorBytes = g.joint.ResidentTensorBytes()
			g.info.ModelSchema, g.info.FeatureVersion = g.joint.Schema(), g.joint.FeatureVersion()
		} else {
			g.info.MetadataSHA256 = g.model.MetadataSHA256()
			g.info.WeightsSHA256 = g.model.WeightsSHA256()
			g.info.ResidentTensorBytes = g.model.ResidentTensorBytes()
		}
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
	feedback, diagnosis, err := resolveTypedPathOptions(options, g.model != nil || g.joint != nil || g.three != nil)
	if err != nil {
		return Result{}, err
	}
	observation, err := copyPathObservation(options.Observation)
	if err != nil {
		return Result{}, err
	}
	return generateTypedPathRequest(ctx, filename, source, activityName, document,
		typedPathModel{model: g.model, joint: g.joint, three: g.three, retention: &g.info, diagnosis: diagnosis, observation: observation}, options.StepAttempts, feedback)
}
