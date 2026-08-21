package billing

import fraud "example.com/fraud"

//gooo:semantic activity id="billing://activity/pay-order"
func PayOrder(order Order) Payment {
	fraud.Check(order)
	return Payment{OrderID: order.ID}
}

//gooo:semantic entity id="billing://entity/order"
type Order struct {
	ID string
}

//gooo:semantic entity id="billing://entity/payment"
type Payment struct {
	OrderID string
}
