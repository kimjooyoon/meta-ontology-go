package bodyexecution

import (
	"context"
	"fmt"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

type CompositionOptions struct {
	EntryActivity string
	ModelPath     string
	FillModelPath string
	RecordPolicy  *bodycodegen.RecordAssemblyPolicy
}

// Setup is observed once per composition. Per-activity receipts retain the
// model digests, proposed assignment, finite scores and inference duration.
type CompositionFillModelInfo struct {
	Loaded  bool    `json:"loaded"`
	Loads   int     `json:"loads"`
	SetupMS float64 `json:"setup_ms"`
}

func (g *compositionAssemblyGenerator) generateFill(ctx context.Context, filename string, source []byte,
	activity string, spec *assemblyspec.Spec) (bodycodegen.Result, error) {
	options := bodycodegen.IRBodyFillOptions{}
	if g.options.FillModelPath != "" {
		if g.fillProvider == nil {
			start := time.Now()
			provider, err := decisionroute.LoadTinyGoProvider(g.options.FillModelPath)
			if err != nil {
				return bodycodegen.Result{}, fmt.Errorf("load retained body-fill model: %w", err)
			}
			g.fillProvider = provider
			g.fillInfo = &CompositionFillModelInfo{Loaded: true, Loads: 1, SetupMS: float64(time.Since(start)) / float64(time.Millisecond)}
			options.TinyModelLoadMS = &g.fillInfo.SetupMS
		}
		options.TinyGoProvider = g.fillProvider
	}
	return bodycodegen.GenerateWithSourceIRBodyFill(ctx, filename, source, activity, spec, "", "", options)
}
