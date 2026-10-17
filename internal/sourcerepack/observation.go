package sourcerepack

import (
	"errors"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

// Observation keeps a blocked plan distinct from the diagnostic used by write mode.
func (blocked repackBlocked) observation() string {
	return fmt.Sprintf("source-repacker: blocked subject=%q reason=%q", blocked.Subject, blocked.Reason)
}

func checkPlansTo(output io.Writer, cfg config, report metricevidence.Report, indicators []metricevidence.Indicator) error {
	planned, blocked, matched := 0, 0, 0
	for _, indicator := range indicators {
		if cfg.subject != "" && cfg.subject != indicator.Subject {
			continue
		}
		matched++
		_, err := planRepack(cfg.root, indicator.Subject, report.Meta.Policy.MaxFileLines)
		if err == nil {
			planned++
			continue
		}
		if blockedPlan, ok := errors.AsType[repackBlocked](err); ok {
			blocked++
			fmt.Fprintln(output, blockedPlan.observation())
			continue
		}
		return fmt.Errorf("plan %s: %w", indicator.Subject, err)
	}
	if cfg.subject != "" && matched == 0 {
		return fmt.Errorf("subject %q is not an actionable split indicator", cfg.subject)
	}
	fmt.Fprintf(output, "source-repacker: actionable=%d planned=%d blocked=%d write=false\n", matched, planned, blocked)
	return nil
}
