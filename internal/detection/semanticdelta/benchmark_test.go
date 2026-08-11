package semanticdelta

import "testing"

func BenchmarkDetectInScopeFixture(b *testing.B) {
	data, err := fixtureFiles.ReadFile("testdata/hypothesis-in-scope.json")
	if err != nil {
		b.Fatal(err)
	}
	request, err := Decode(data)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ReportMetric(float64(len(request.Delta.AddedFacts)), "facts")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Detect(request.Delta, request.Allowed); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSONInScopeFixture(b *testing.B) {
	data, err := fixtureFiles.ReadFile("testdata/hypothesis-in-scope.json")
	if err != nil {
		b.Fatal(err)
	}
	request, err := Decode(data)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ReportMetric(float64(len(data)), "input-bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := EncodeJSON(request); err != nil {
			b.Fatal(err)
		}
	}
}
