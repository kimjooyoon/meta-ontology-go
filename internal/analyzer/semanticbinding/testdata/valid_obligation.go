package billing

//gooo:obligation id="billing://obligation/order-number" subject="billing://entity/order" pressure="billing://pressure/order-number"
type Order struct {
	OrderNumber string
}
