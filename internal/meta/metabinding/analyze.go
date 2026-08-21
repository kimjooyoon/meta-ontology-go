package metabinding

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func Build(root string, in input) (Report, sourcepolicy.Indicator, error) {
	registry := canonicalRegistry()
	index, err := registryIndex(registry)
	if err != nil {
		return Report{}, sourcepolicy.Indicator{}, err
	}
	ontologyDigest, err := readOntology(root)
	if err != nil {
		return Report{}, sourcepolicy.Indicator{}, err
	}
	source := append([]sourcepolicy.Indicator(nil), in.document.Meta.Indicators...)
	self := selfIndicator(countUnbound(source, index))
	all := append(append([]sourcepolicy.Indicator(nil), source...), self)
	summary, witnesses := summarize(all, index)
	report := Report{
		Schema: Schema, CommitSHA: in.document.CommitSHA, Repository: in.document.Repository,
		SourceMetricsDigest: in.sourceDigest, RegistryDigest: digestJSON(registry),
		OntologyDigest: ontologyDigest,
		Decision:       "PASS", Reason: "META_BINDING_COMPLETE",
		Summary: summary, SelfIndicator: self, Witnesses: witnesses,
	}
	if summary.UnboundIndicators != 0 {
		report.Decision, report.Reason = "FAIL_CLOSED", "META_BINDING_INCOMPLETE"
	}
	report.ReportDigest = digestJSON(report)
	return report, self, nil
}

func countUnbound(indicators []sourcepolicy.Indicator, index map[string]Binding) int {
	count := 0
	for _, indicator := range indicators {
		if len(bindingReasons(indicator, index)) != 0 {
			count++
		}
	}
	return count
}

func bindingReasons(indicator sourcepolicy.Indicator, index map[string]Binding) []string {
	operation := string(indicator.Operation)
	binding, exists := index[operation]
	reasons := make([]string, 0, 4)
	if !exists {
		reasons = append(reasons, "UNREGISTERED_OPERATION")
	}
	if strings.TrimSpace(indicator.Producer) == "" {
		reasons = append(reasons, "MISSING_PRODUCER")
	}
	if strings.TrimSpace(indicator.Consumer) == "" {
		reasons = append(reasons, "MISSING_CONSUMER")
	}
	if exists && normalizeProof(fmt.Sprint(indicator.Proof)) != binding.ProofChoice {
		reasons = append(reasons, "PROOF_CHOICE_MISMATCH")
	}
	sort.Strings(reasons)
	return reasons
}
