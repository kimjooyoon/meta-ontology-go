package billing

//gooo:bind id="billing://entity/order"
type Order struct {
	//gooo:bind id="billing://field/order-number"
	OrderNumber string

	//gooo:bind id="billing://field/order-number"
	DisplayNumber string
}
