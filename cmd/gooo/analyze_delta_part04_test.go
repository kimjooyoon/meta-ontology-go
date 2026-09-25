package main

const billingAnalyzeAuthority = `package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`
const billingAnalyzeRenamedAuthority = `package billing
namespace billing

entity PurchaseOrder id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(PurchaseOrder, PaymentMethod) -> Payment
`
const billingAnalyzeAmbiguousAuthority = `package billing
namespace billing

entity Order id "billing://entity/order"
entity AlternateOrder id "billing://entity/order-alt"
entity Payment id "billing://entity/payment"

activity PayOrder(Order) -> Payment
`
const billingAnalyzeAmbiguousGo = `package billing

//gooo:semantic entity id="billing://entity/order" namespace=billing
type Order struct{}

//gooo:semantic entity id="billing://entity/order-alt" namespace=billing
type Order struct{}
type Payment struct{}

func PayOrder(order Order) Payment {
	return Payment{}
}
`
