package integrationprogress

const (
	ObservationSchema = "gooo/integration-progress-observation/v1"
	ReportSchema      = "gooo/integration-progress-report/v1"
	Repository        = "kimjooyoon/meta-ontology-go"
	CohortID          = "gooo.portfolio.pr-541-570.v1"
	WorkflowName      = "CI [PR authoritative]"
	ArtifactPrefix    = "source-line-metrics-"
	MetaOperation     = "measure-integration-progress"
)

type StageSpec struct {
	ID       string
	Step     string
	Activity string
	Computes string
}

func StageSpecs() []StageSpec {
	return []StageSpec{
		{"pull-observed", "get-pull-request", "ObservePullRequest", "gooo.progress.stage:pull-observed:v1"},
		{"authoritative-run-terminal", "select-authoritative-run", "SelectAuthoritativeRun", "gooo.progress.stage:authoritative-run-terminal:v1"},
		{"evidence-artifact-reachable", "select-head-bound-artifact", "BindEvidenceArtifact", "gooo.progress.stage:evidence-artifact-reachable:v1"},
		{"merge-realized", "observe-merge", "ObserveMerge", "gooo.progress.stage:merge-realized:v1"},
		{"merged-evidence-linked", "bind-evidence-before-merge", "BindMergedEvidence", "gooo.progress.stage:merged-evidence-linked:v1"},
	}
}

func PullNumbers() []int {
	result := make([]int, 0, 30)
	for number := 541; number <= 570; number++ {
		result = append(result, number)
	}
	return result
}

func CellDenominator() int { return len(PullNumbers()) * len(StageSpecs()) }
