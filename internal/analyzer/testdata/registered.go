package billing

import (
	"encoding/json"
	"strings"

	fraud "example.com/fraud"
)

//gooo:semantic activity id="billing://activity/pay-order" namespace=billing
func PayOrder(order Order) Payment {
	normalized := strings.TrimSpace(order.ID)
	_, _ = json.Marshal(normalized)
	_ = normalized
	fraud.Check(order)
	return Payment{OrderID: order.ID}
}

//gooo:semantic entity id="billing://entity/order" namespace=billing
type Order struct {
	ID string
}

//gooo:semantic entity id="billing://entity/payment" namespace=billing
type Payment struct {
	OrderID string
}
