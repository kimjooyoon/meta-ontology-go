package semanticdeltareceipt

const equivalentBefore = `package fixture
namespace semanticdeltafixture

entity Order id "gooo://semantic-delta/entity/order"
entity Payment id "gooo://semantic-delta/entity/payment"

activity PayOrder(Order) -> Payment
`

const equivalentAfter = `package fixture
namespace semanticdeltafixture

// Reordered declarations and a presentation comment are still the same claim set.
entity Payment id "gooo://semantic-delta/entity/payment"
entity Order id "gooo://semantic-delta/entity/order"

activity PayOrder( Order ) -> Payment
`

const semanticBefore = equivalentBefore

const semanticAfter = `package fixture
namespace semanticdeltafixture

entity Order id "gooo://semantic-delta/entity/order"
entity Payment id "gooo://semantic-delta/entity/payment"
entity Reversal id "gooo://semantic-delta/entity/reversal"

activity PayOrder(Order) -> Reversal
`

const indeterminateAfter = `package fixture
namespace semanticdeltafixture

entity Order id "gooo://semantic-delta/entity/order"
entity Payment id "gooo://semantic-delta/entity/payment"

// This declaration is outside the bounded experiment grammar.
assert PayOrder preserves Payment
activity PayOrder(Order) -> Payment
`

func CaseInput(id, subjectSHA string) Input {
	input := Input{CaseID: id, SubjectSHA: subjectSHA, Before: []byte(equivalentBefore), After: []byte(equivalentAfter)}
	switch id {
	case "equivalent":
	case "semantic-change":
		input.Before, input.After = []byte(semanticBefore), []byte(semanticAfter)
	case "indeterminate":
		input.Before, input.After = []byte(semanticBefore), []byte(indeterminateAfter)
	default:
		input.Before, input.After = nil, nil
	}
	return input
}
