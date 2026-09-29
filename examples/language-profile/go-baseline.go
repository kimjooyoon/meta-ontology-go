// gooo:namespace billing
package billing

// gooo:id billing://entity/order
type Order struct{}

// gooo:id billing://entity/payment-method
type PaymentMethod struct{}

// gooo:id billing://entity/payment
type Payment struct{}

func PayOrder(order Order, method PaymentMethod) Payment { return Payment{} }
