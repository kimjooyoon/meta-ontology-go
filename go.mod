module github.com/kimjooyoon/meta-ontology-go

go 1.27.1

tool (
	github.com/kimjooyoon/meta-ontology-go/bootstrap/source-repacker
	github.com/kimjooyoon/meta-ontology-go/scripts/line-metrics
	github.com/kimjooyoon/meta-ontology-go/scripts/maintenance
	github.com/kimjooyoon/meta-ontology-go/scripts/meta-execution
	github.com/kimjooyoon/meta-ontology-go/scripts/meta-receipts
	github.com/kimjooyoon/meta-ontology-go/scripts/refactor-metrics
	github.com/kimjooyoon/meta-ontology-go/scripts/source-splitter
	github.com/kimjooyoon/meta-ontology-go/scripts/verify
)

require github.com/kimjooyoon/gooo-decision-runtime v0.2.12-experimental
