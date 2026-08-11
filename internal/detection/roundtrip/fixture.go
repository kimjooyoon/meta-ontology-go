package roundtrip

import (
	_ "embed"
)

//go:embed testdata/minimal.gooo
var minimalDSL []byte

//go:embed testdata/minimal.go
var minimalGo []byte

// MinimalFixture returns an executable projection witness used by tests and
// examples. Its semantic snapshots match the smallest billing activity graph.
func MinimalFixture() Observation {
	ir := IR{
		Package:   "billing",
		Namespace: "billing",
		Nodes: []Node{
			{ID: "billing://activity/pay-order", Kind: Activity, Name: "PayOrder"},
			{ID: "billing://entity/order", Kind: Entity, Name: "Order"},
			{ID: "billing://entity/payment", Kind: Entity, Name: "Payment"},
		},
		Facts: []Fact{
			{Subject: "billing://activity/pay-order", Predicate: "prov:used", Object: "billing://entity/order"},
			{Subject: "billing://entity/payment", Predicate: "prov:wasGeneratedBy", Object: "billing://activity/pay-order"},
		},
	}
	return Observation{DSL: ir, IR: ir, GoIR: ir, AfterGo: MinimalGo()}
}

// MinimalDSL returns a copy of the fixture's source view.
func MinimalDSL() []byte { return append([]byte(nil), minimalDSL...) }

// MinimalGo returns a copy of the fixture's generated Go view.
func MinimalGo() []byte { return append([]byte(nil), minimalGo...) }
