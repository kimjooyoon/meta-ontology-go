package main

const deferredEntityFieldsSource = `package billing
namespace billing

entity Order id "billing://entity/order" fields {
    field OrderNumber id "billing://field/order-number" type string required one
}

activity PayOrder(Order) -> Order
`
