package metricprogram

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type entitySpec struct {
	Name string
	ID   string
}

var canonicalEntities = []entitySpec{
	{Name: "SourceMetrics", ID: "gooo://metric-meta-program/entity/source-metrics"},
	{Name: "BoundMetrics", ID: "gooo://metric-meta-program/entity/bound-metrics"},
	{Name: "RootPolicy", ID: "gooo://metric-meta-program/entity/root-policy"},
	{Name: "DimensionRegistry", ID: "gooo://metric-meta-program/entity/dimension-registry"},
	{Name: "ProjectedMetrics", ID: "gooo://metric-meta-program/entity/projected-metrics"},
	{Name: "CounterfactualBoundary", ID: "gooo://metric-meta-program/entity/counterfactual-boundary"},
	{Name: "WorkspaceReceipt", ID: "gooo://metric-meta-program/entity/workspace-receipt"},
	{Name: "VerificationEvidence", ID: "gooo://metric-meta-program/entity/verification-evidence"},
	{Name: "FixedPointReceipt", ID: "gooo://metric-meta-program/entity/fixed-point-receipt"},
}

func CanonicalSource() []byte {
	var source strings.Builder
	source.WriteString("package metricmetaprogram\nnamespace metricmetaprogram\n\n")
	for _, entity := range canonicalEntities {
		fmt.Fprintf(&source, "entity %s id %q\n", entity.Name, entity.ID)
	}
	source.WriteString("\n")
	for _, operation := range canonicalOperations {
		fmt.Fprintf(&source, "activity %s(%s) -> %s\n", operation.Activity, operation.InputEntity, operation.OutputEntity)
	}
	return []byte(source.String())
}

func semanticDigest(source []byte) (string, error) {
	file, diagnostics := syntax.ParseFile(ProgramSourceFilename, string(source))
	if diagnostics.HasErrors() {
		return "", fmt.Errorf("canonical meta program has syntax errors")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", fmt.Errorf("lower canonical meta program: %w", err)
	}
	return normalizeSemanticDigest(ir.StableHash())
}
