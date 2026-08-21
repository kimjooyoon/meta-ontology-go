package verify

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var entities = []struct {
	name string
	id   string
}{
	{name: "SourceMetrics", id: "gooo://metric-meta-program/entity/source-metrics"},
	{name: "BoundMetrics", id: "gooo://metric-meta-program/entity/bound-metrics"},
	{name: "RootPolicy", id: "gooo://metric-meta-program/entity/root-policy"},
	{name: "DimensionRegistry", id: "gooo://metric-meta-program/entity/dimension-registry"},
	{name: "ProjectedMetrics", id: "gooo://metric-meta-program/entity/projected-metrics"},
	{name: "CounterfactualBoundary", id: "gooo://metric-meta-program/entity/counterfactual-boundary"},
	{name: "WorkspaceReceipt", id: "gooo://metric-meta-program/entity/workspace-receipt"},
	{name: "VerificationEvidence", id: "gooo://metric-meta-program/entity/verification-evidence"},
	{name: "FixedPointReceipt", id: "gooo://metric-meta-program/entity/fixed-point-receipt"},
}

func expectedSource() []byte {
	var source strings.Builder
	source.WriteString("package metricmetaprogram\nnamespace metricmetaprogram\n\n")
	for _, entity := range entities {
		fmt.Fprintf(&source, "entity %s id %q\n", entity.name, entity.id)
	}
	source.WriteString("\n")
	for _, operation := range operations {
		fmt.Fprintf(&source, "activity %s(%s) -> %s\n", operation.Activity, operation.InputEntity, operation.OutputEntity)
	}
	return []byte(source.String())
}

func semanticDigest(source []byte) (string, error) {
	file, diagnostics := syntax.ParseFile(programSourceFilename, string(source))
	if diagnostics.HasErrors() {
		return "", fmt.Errorf("meta program source has syntax errors")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", fmt.Errorf("lower meta program source: %w", err)
	}
	return normalizeSemanticDigest(ir.StableHash())
}
