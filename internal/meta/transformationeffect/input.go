package transformationeffect

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func loadInputs(opts Options) (inputSet, error) {
	var in inputSet
	items := []struct {
		path string
		into any
		dst  *string
	}{
		{opts.MetricsPath, &in.metrics, &in.digests.SourceMetrics},
		{opts.PlanPath, &in.plan, &in.digests.Plan},
		{opts.ExecutionPath, &in.execution, &in.digests.Execution},
		{opts.ReceiptsPath, &in.receipts, &in.digests.Receipts},
		{opts.ProvenancePath, &in.provenance, &in.digests.Provenance},
	}
	for _, item := range items {
		payload, err := os.ReadFile(item.path)
		if err != nil {
			return in, err
		}
		if err := json.Unmarshal(payload, item.into); err != nil {
			return in, fmt.Errorf("decode %s: %w", item.path, err)
		}
		*item.dst = hashBytes(payload)
	}
	if err := validateInputs(in, opts.ExpectedSHA); err != nil {
		return in, err
	}
	return in, nil
}

func validateInputs(in inputSet, expected string) error {
	if !validSHA(expected) || in.metrics.CommitSHA != expected ||
		in.plan.HeadSHA != expected || in.execution.HeadSHA != expected ||
		in.receipts.HeadSHA != expected || in.provenance.HeadSHA != expected {
		return fmt.Errorf("artifact head is not exact %s", expected)
	}
	if in.metrics.Meta.Schema != sourcepolicy.IndicatorSchema ||
		!in.metrics.Meta.Policy.ExemptProjectRootTopology || !metricLedgerBound(in) {
		return fmt.Errorf("source indicator ledger is not bound")
	}
	wantExecution := generation.BuildExecutionManifest(in.plan)
	wantReceipts := generation.VerifyReceipts(in.plan, in.receipts.Receipts)
	wantProvenance := generation.BindArtifactProvenance(in.plan, in.execution, in.receipts)
	if !reflect.DeepEqual(wantExecution, in.execution) ||
		!reflect.DeepEqual(wantReceipts, in.receipts) ||
		!reflect.DeepEqual(wantProvenance, in.provenance) {
		return fmt.Errorf("meta artifact replay diverged")
	}
	return nil
}

func metricLedgerBound(in inputSet) bool {
	ledger := in.plan.IndicatorDecisionLedger
	if ledger == nil || ledger.IndicatorCount != len(in.metrics.Meta.Indicators) ||
		len(ledger.Entries) != len(in.metrics.Meta.Indicators) {
		return false
	}
	left, right := make([]string, 0, len(ledger.Entries)), make([]string, 0, len(ledger.Entries))
	for _, indicator := range in.metrics.Meta.Indicators {
		left = append(left, hashJSON(indicator))
	}
	for _, entry := range ledger.Entries {
		right = append(right, hashJSON(entry.SourceIndicator))
	}
	sort.Strings(left)
	sort.Strings(right)
	return reflect.DeepEqual(left, right)
}
