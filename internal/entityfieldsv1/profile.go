package entityfieldsv1

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

const (
	Schema        = "gooo/entity-fields/v1"
	SourceID      = "billing://entity/order"
	FieldID       = "billing://field/order-number"
	SecondFieldID = "billing://field/customer-name"
)

// CanonicalSource is the profile fixture used by the V1 vertical slice.
const CanonicalSource = `package billing
namespace billing

entity Order id "billing://entity/order" fields {
    field OrderNumber id "billing://field/order-number" type string required one
    field CustomerName id "billing://field/customer-name" type string required one
}

activity LoadOrder(Order) -> Order computes "field.read:CustomerName=billing://field/customer-name"
`

func support() syntax.EntityFieldsSupport { return syntax.EntityFieldsV1Support() }
