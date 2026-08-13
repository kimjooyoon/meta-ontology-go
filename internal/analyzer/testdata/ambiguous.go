package billing

import fraud "example.com/fraud"

//gooo:semantic activity id="billing://activity/settle" namespace=billing
func Settle(order Order) {
	fraud.Check(order)
}

//gooo:semantic entity id="billing://entity/order" namespace=billing
type Order struct{}
