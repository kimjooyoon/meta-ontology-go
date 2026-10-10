module github.com/kimjooyoon/meta-ontology-go

go 1.27.2

tool (
	github.com/kimjooyoon/meta-ontology-go/bootstrap/source-repacker
	github.com/kimjooyoon/meta-ontology-go/scripts/line-metrics
	github.com/kimjooyoon/meta-ontology-go/scripts/maintenance
	github.com/kimjooyoon/meta-ontology-go/scripts/meta-execution
	github.com/kimjooyoon/meta-ontology-go/scripts/meta-receipts
	github.com/kimjooyoon/meta-ontology-go/scripts/receipt-schema
	github.com/kimjooyoon/meta-ontology-go/scripts/refactor-metrics
	github.com/kimjooyoon/meta-ontology-go/scripts/source-splitter
	github.com/kimjooyoon/meta-ontology-go/scripts/verify
)

require (
	github.com/kimjooyoon/gooo-decision-runtime v0.2.27-experimental
	github.com/kimjooyoon/gooo-jev v0.0.0-20260928032625-e146e6dbb34a
)
