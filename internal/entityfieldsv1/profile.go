package entityfieldsv1

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

const (
	Schema   = "gooo/entity-fields/v1"
	SourceID = "billing://entity/order"
	FieldID  = "billing://field/order-number"
)

// CanonicalSource is the profile fixture used by the V1 vertical slice.
const CanonicalSource = `package billing
namespace billing

entity Order id "billing://entity/order" fields {
    field OrderNumber id "billing://field/order-number" type string required one
}

activity ParseEntityFields(Order) -> Order
activity FormatEntityFields(Order) -> Order
activity BindEntityFieldsProfile(Order) -> Order
activity LowerEntityFields(Order) -> Order
activity ValidateStableIDs(Order) -> Order
activity PreserveDeclarationOrder(Order) -> Order
activity RoundTripGet(Order) -> Order
activity RoundTripPut(Order) -> Order
activity GenerateGoProjection(Order) -> Order
activity GenerateSourceMap(Order) -> Order
activity ResolveLSPNavigation(Order) -> Order
activity PublishEntityFieldsReceipt(Order) -> Order
`

func support() syntax.EntityFieldsSupport { return syntax.EntityFieldsV1Support() }
