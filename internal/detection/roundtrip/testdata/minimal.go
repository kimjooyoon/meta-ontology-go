package billinggen

//gooo:generated:start id="billing://entity/order" kind="entity"
type Order struct{}

//gooo:generated:end id="billing://entity/order"

//gooo:generated:start id="billing://entity/payment" kind="entity"
type Payment struct{}

//gooo:generated:end id="billing://entity/payment"

//gooo:generated:start id="billing://activity/pay-order" kind="activity"
func PayOrder(order Order) Payment { return Payment{} }

//gooo:generated:end id="billing://activity/pay-order"
