module github.com/kimjooyoon/meta-ontology-go

go 1.27.0

tool (
	github.com/kimjooyoon/meta-ontology-go/bootstrap/source-repacker
	github.com/kimjooyoon/meta-ontology-go/scripts/line-metrics
	github.com/kimjooyoon/meta-ontology-go/scripts/maintenance
	github.com/kimjooyoon/meta-ontology-go/scripts/refactor-metrics
	github.com/kimjooyoon/meta-ontology-go/scripts/source-splitter
	github.com/kimjooyoon/meta-ontology-go/scripts/verify
)
