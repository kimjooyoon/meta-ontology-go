package roundtripcontract

import "testing"

func BenchmarkCanonicalEvidence(b *testing.B) {
	evidence := MinimalScenarios()[1].Evidence
	b.ReportMetric(float64(evidence.Measurement.SourceBytes), "fixture-bytes")
	b.ReportMetric(float64(evidence.Measurement.Nodes), "fixture-nodes")
	b.ReportMetric(float64(evidence.Measurement.Facts), "fixture-facts")
	b.ReportMetric(float64(evidence.Measurement.Regions), "fixture-regions")
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := CanonicalJSON(evidence); err != nil {
			b.Fatal(err)
		}
	}
}
