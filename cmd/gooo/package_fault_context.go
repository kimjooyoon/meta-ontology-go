package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Display retained facts without evaluating the producer or changing its result.
func writePackageFaultContext(out io.Writer, runtime bodyexecution.CompositionRuntime, fault bodyexecution.CompositionFault) {
	site := fault.Site
	fmt.Fprintf(out, "  activity=%q operator=%q left=%d right=%d\n", site.ActivityID, site.Operator, fault.Left, fault.Right)
	fmt.Fprintf(out, "  Go projection [%d:%d]: %q\n", site.Start, site.End, site.Expression)
	if site.Variable != "" {
		fmt.Fprintf(out, "  local=%q assignment=%q\n", site.Variable, site.Assignment)
	}
	if source := packageFaultLocalRHS(runtime, site); source != nil {
		fmt.Fprintf(out, "  selected Gooo RHS [%d:%d]: %q (source %s)\n", source.Start, source.End, source.Expression, source.SourceSHA256)
	}
}

// A local compound assignment spans the target and RHS in the pure projection.
// Match its retained RHS by activity, variable, operation, digest and byte range.
func packageFaultLocalRHS(runtime bodyexecution.CompositionRuntime, fault bodyexecution.CompositionFaultSite) *bodycodegen.BodyExpressionSource {
	if fault.Variable == "" || fault.Assignment != "/=" || fault.ProjectionSHA256 == "" {
		return nil
	}
	var selected *bodycodegen.BodyExpressionSource
	matches := 0
	for _, local := range runtime.LocalSites {
		if local.RootActivityID != fault.RootActivityID || local.ActivityID != fault.ActivityID || local.Name != fault.Variable ||
			local.Operation != fault.Assignment || local.ProjectionSHA256 != fault.ProjectionSHA256 ||
			local.Start <= fault.Start || local.End != fault.End {
			continue
		}
		matches++
		source := local.Source.Value
		if source != nil && source.Status == "MAPPED" && source.View == "selected_gooo_source" &&
			source.SourceSHA256 != "" && source.SourceSHA256 == runtime.SelectedSourceSHA256 && source.Start >= 0 && source.End > source.Start {
			selected = source
		}
	}
	if matches != 1 {
		return nil
	}
	return selected
}
