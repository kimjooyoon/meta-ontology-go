package metricintervention

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metrictransition"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func LoadBaseline(metricsPath, repository, subjectSHA string) (Baseline, error) {
	report, err := artifact.ReadJSON[linecaps.LineMetricsReport](metricsPath)
	if err != nil {
		return Baseline{}, err
	}
	if report.Repository != repository || report.CommitSHA != subjectSHA {
		return Baseline{}, fmt.Errorf("source metrics exact-subject binding is invalid")
	}
	if report.Meta.Schema != sourcepolicy.IndicatorSchema || report.Meta.Policy.Schema != sourcepolicy.Schema {
		return Baseline{}, fmt.Errorf("source metric policy schema is invalid")
	}
	root, err := projectRootCounts(report.Directories)
	if err != nil {
		return Baseline{}, err
	}
	if !report.Meta.Policy.ExemptProjectRootTopology || !rootEvidenceComplete(report.Meta, root) {
		return Baseline{}, fmt.Errorf("project root topology exemption is not evidenced")
	}
	report.Root, report.StorageRoot = "", ""
	sourceDigest, err := artifact.Digest(report)
	if err != nil {
		return Baseline{}, err
	}
	policy := RootPolicy{CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE", TopologyReason: "ROOT_TOPOLOGY_EXEMPT", READMERequirement: "NOT_APPLICABLE"}
	return Baseline{Schema: BaselineSchema, RepositoryStateSchema: metrictransition.StateSchema, SourceIndicatorSchema: report.Meta.Schema, SourcePolicySchema: report.Meta.Policy.Schema, Repository: repository, SubjectSHA: subjectSHA, Root: root, RootPolicy: policy, SourceMetricsDigest: sourceDigest}, nil
}

func projectRootCounts(metrics []linecaps.DirectoryMetric) (metrictransition.Counts, error) {
	for _, metric := range metrics {
		if metric.Path != "." {
			continue
		}
		if metric.SubjectKind != sourcepolicy.SubjectKindProjectRoot {
			return metrictransition.Counts{}, fmt.Errorf("project root subject kind is invalid")
		}
		data, err := json.Marshal(metric)
		var counts metrictransition.Counts
		if err == nil {
			err = json.Unmarshal(data, &counts)
		}
		return counts, err
	}
	return metrictransition.Counts{}, fmt.Errorf("logical source metrics have no project root")
}
