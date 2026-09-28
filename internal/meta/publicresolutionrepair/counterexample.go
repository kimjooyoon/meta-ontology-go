package publicresolutionrepair

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

type Counterexample struct {
	Valid                     bool   `json:"valid"`
	RecordSchema              string `json:"record_schema"`
	RecordDigest              string `json:"record_digest"`
	AggregateReportDigest     string `json:"aggregate_report_digest"`
	OriginSourceDigest        string `json:"origin_source_digest"`
	Operation                 string `json:"operation"`
	CaseID                    string `json:"case_id"`
	Decision                  string `json:"decision"`
	Reason                    string `json:"reason"`
	ChangedComponent          string `json:"changed_component"`
	GraphVariant              string `json:"graph_variant"`
	OmittedTarget             string `json:"omitted_target"`
	ObservedAffectedPartition string `json:"observed_affected_partition"`
	ObservedAffectedTest      string `json:"observed_affected_test"`
	OriginalClosureEdges      int    `json:"original_closure_edges"`
	OriginalImpactedTests     int    `json:"original_impacted_tests"`
	OriginalUnaffectedTests   int    `json:"original_unaffected_tests"`
	CausalStage               string `json:"causal_stage"`
	CausalStep                string `json:"causal_step"`
}

type v15Case struct {
	CaseID             string `json:"case_id"`
	Decision           string `json:"decision"`
	Reason             string `json:"reason"`
	Operation          string `json:"operation"`
	ImpactedPartitions int    `json:"impacted_partitions"`
	Unaffected         int    `json:"unaffected_partitions"`
	ClosureEdges       int    `json:"closure_edges"`
}

type v15PolicyCase struct {
	ID           string `json:"id"`
	Decision     string `json:"decision"`
	Changed      string `json:"changed"`
	GraphVariant string `json:"graph_variant"`
	Option       string `json:"option"`
}

type v15Policy struct {
	SourceDigest string          `json:"source_digest"`
	Cases        []v15PolicyCase `json:"cases"`
}

type v15Aggregate struct {
	Schema   string    `json:"schema"`
	Policy   v15Policy `json:"policy"`
	Cases    []v15Case `json:"cases"`
	Decision string    `json:"decision"`
}

func LoadCounterexample(hiddenRecord, aggregateReport []byte, policy Policy) (Counterexample, error) {
	var hidden v15Case
	if err := json.Unmarshal(hiddenRecord, &hidden); err != nil {
		return Counterexample{}, fmt.Errorf("decode v15 hidden-dependency record: %w", err)
	}
	var aggregate v15Aggregate
	if err := json.Unmarshal(aggregateReport, &aggregate); err != nil {
		return Counterexample{}, fmt.Errorf("decode v15 partial-reuse report: %w", err)
	}
	if hidden.CaseID != OriginalCounterexampleCaseID || hidden.Decision != DecisionRefuted || hidden.Reason != "FAIL_CLOSED_PARTIAL_REUSE_CONTRADICTION" || aggregate.Decision != DecisionClosed {
		return Counterexample{}, errors.New("v15 hidden-dependency record is not the preserved REFUTED counterexample")
	}
	var sourceCase v15PolicyCase
	for _, item := range aggregate.Policy.Cases {
		if item.ID == OriginalCounterexampleCaseID {
			sourceCase = item
			break
		}
	}
	if sourceCase.ID == "" || sourceCase.Decision != DecisionRefuted || sourceCase.Changed == "" || sourceCase.GraphVariant == "" || sourceCase.Option == "" {
		return Counterexample{}, errors.New("v15 aggregate does not preserve the canonical hidden-dependency case")
	}
	target := strings.TrimPrefix(sourceCase.Option, "omitted-target=")
	if target == sourceCase.Option || target == "" {
		return Counterexample{}, errors.New("v15 hidden-dependency case does not identify an omitted target")
	}
	partition, ok := partitionFor(policy, target)
	if !ok {
		return Counterexample{}, fmt.Errorf("v15 hidden-dependency target %q is not a canonical partition", target)
	}
	return Counterexample{
		Valid: true, RecordSchema: "gooo/public-partial-test-reuse-report/v1", RecordDigest: cache.HashBytes(hiddenRecord).String(),
		AggregateReportDigest: cache.HashBytes(aggregateReport).String(), OriginSourceDigest: aggregate.Policy.SourceDigest,
		Operation: hidden.Operation, CaseID: hidden.CaseID, Decision: hidden.Decision, Reason: hidden.Reason,
		ChangedComponent: sourceCase.Changed, GraphVariant: sourceCase.GraphVariant, OmittedTarget: target,
		ObservedAffectedPartition: target, ObservedAffectedTest: partition.TestName,
		OriginalClosureEdges: hidden.ClosureEdges, OriginalImpactedTests: hidden.ImpactedPartitions, OriginalUnaffectedTests: hidden.Unaffected,
		CausalStage: "GRAPH_AND_IMPACT", CausalStep: "CHANGED_DEPENDENCY_OUTSIDE_COMPUTED_CLOSURE",
	}, nil
}

func InvalidCounterexample(valid Counterexample) Counterexample {
	valid.Valid = false
	return valid
}
