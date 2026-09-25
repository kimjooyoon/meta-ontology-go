package roundtrip

import (
	_ "embed"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

//go:embed testdata/minimal.gooo
var minimalDSL []byte

//go:embed testdata/minimal.go
var minimalGo []byte

// MinimalFixture returns an executable projection witness using the current
// semantic IR and current generated marker grammar.
func MinimalFixture() Observation {
	ir := minimalIR()
	return Observation{
		DSL:      ir,
		IR:       ir,
		GoIR:     ir,
		BeforeGo: MinimalGo(),
		AfterGo:  MinimalGo(),
	}
}

// MinimalDSL returns a copy of the fixture's source view.
func MinimalDSL() []byte { return append([]byte(nil), minimalDSL...) }

// MinimalGo returns a copy of the fixture's generated Go view.
func MinimalGo() []byte { return append([]byte(nil), minimalGo...) }

func minimalIR() semantic.IR {
	ir := semantic.NewIR("billing", semantic.Namespace("billing"))
	order := mustNode(semantic.Entity, "billing://entity/order", "billing", "Order")
	payment := mustNode(semantic.Entity, "billing://entity/payment", "billing", "Payment")
	payOrder := mustNode(semantic.Activity, "billing://activity/pay-order", "billing", "PayOrder")
	for _, node := range []semantic.Node{payOrder, order, payment} {
		if err := ir.AddNode(node); err != nil {
			panic(err)
		}
	}
	for _, fact := range []semantic.Fact{
		semantic.NewUsedFact(payOrder.ID, order.ID),
		semantic.NewWasGeneratedByFact(payment.ID, payOrder.ID),
	} {
		if err := ir.AddFact(fact); err != nil {
			panic(err)
		}
	}
	return ir
}

func mustNode(kind semantic.Kind, id, namespace, name string) semantic.Node {
	node, err := semantic.NewNodeFromStrings(kind, id, namespace, name)
	if err != nil {
		panic(fmt.Sprintf("minimal fixture node %q: %v", id, err))
	}
	return node
}
