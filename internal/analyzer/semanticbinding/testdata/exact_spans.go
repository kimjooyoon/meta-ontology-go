package billing

//gooo:bind id="billing://entity/order" role="HANDWRITTEN_IMPL"
type Order struct{}

//gooo:bind id="billing://entity/payment" role="GENERATED_IMPL"
type Payment struct{}
