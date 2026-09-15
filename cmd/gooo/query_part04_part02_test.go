package main

const billingSource = `package billing
namespace billing
entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"
activity PayOrder(Order, PaymentMethod) -> Payment
`
