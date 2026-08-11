package billing

// PayOrderLogic is the irreducible implementation slot for the PayOrder activity.
// Structural boundaries are generated from main.gooo; domain calculation remains here.
func PayOrderLogic(order Order, method PaymentMethod) Payment {
	return Payment{OrderID: order.ID, MethodID: method.ID}
}

type Order struct{ ID string }
type PaymentMethod struct{ ID string }
type Payment struct {
	OrderID  string
	MethodID string
}
