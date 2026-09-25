package integrationprogress

import (
	"fmt"
	"strings"
)

func RenderProgram() []byte {
	var source strings.Builder
	source.WriteString("package integrationprogress\nnamespace integrationprogress\n\n")
	source.WriteString("entity PullRequestCohort id \"gooo://meta/integration-progress/entity/pull-request-cohort\"\n")
	source.WriteString("entity PortfolioObservation id \"gooo://meta/integration-progress/entity/portfolio-observation\"\n")
	source.WriteString("entity ProgressCell id \"gooo://meta/integration-progress/entity/progress-cell\"\n")
	source.WriteString("entity ProgressReport id \"gooo://meta/integration-progress/entity/progress-report\"\n\n")
	for _, stage := range StageSpecs() {
		fmt.Fprintf(&source, "activity %s(PortfolioObservation) -> ProgressCell computes %q\n",
			stage.Activity, stage.Computes)
	}
	source.WriteString("activity MeasureIntegrationProgress(ProgressCell) -> ProgressReport computes \"gooo.metric.integration-progress:v1\"\n")
	source.WriteString("activity ReplayIntegrationProgress(ProgressReport) -> ProgressReport computes \"gooo.replay.integration-progress:v1\"\n")
	return []byte(source.String())
}
