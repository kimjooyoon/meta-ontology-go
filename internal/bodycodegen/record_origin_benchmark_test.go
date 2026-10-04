package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func BenchmarkRecordOriginContext(b *testing.B) {
	source, err := os.ReadFile("../../examples/body-codegen/record-field-updates.gooo.fixture")
	if err != nil {
		b.Fatal(err)
	}
	plan, err := prepareRecordAssembly(context.Background(), "r.gooo", source, "Select")
	if err != nil {
		b.Fatal(err)
	}
	flow := recordValueFlow(plan)
	raw, _ := json.Marshal(flow)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		result := recordPlanModelContext(plan, jointdecision.RecordOriginSharedFeatureVersion)
		if result.Status != "ENCODED" {
			b.Fatal(result.Reason)
		}
	}
	b.ReportMetric(float64(len(flow.Nodes)), "nodes")
	b.ReportMetric(float64(len(raw)), "graph-bytes")
}
