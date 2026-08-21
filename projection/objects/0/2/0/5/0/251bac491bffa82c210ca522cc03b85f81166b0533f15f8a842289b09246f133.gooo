package semanticdelta

import "testing"

func BenchmarkDetectInScopeRequest(b *testing.B) {
	request := benchmarkRequest()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Detect(request.Delta, request.Allowed); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeJSONRequest(b *testing.B) {
	request := benchmarkRequest()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := EncodeJSON(request); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkRequest() Request {
	return Request{
		Allowed: Scope{
			Prefixes:   []string{"billing://"},
			Predicates: []string{"prov:used", "gooo:invokes"},
		},
		Delta: Delta{
			AddedNodes: []Node{
				{ID: "billing://entity/receipt", Kind: "Entity"},
				{ID: "billing://entity/order", Kind: "Entity"},
			},
			AddedFacts: []Fact{
				{Subject: "billing://activity/pay-order", Predicate: "prov:used", Object: "billing://entity/order"},
				{Subject: "billing://activity/pay-order", Predicate: "gooo:invokes", Object: "billing://entity/receipt"},
			},
		},
	}
}
