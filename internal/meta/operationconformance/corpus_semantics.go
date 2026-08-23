package operationconformance

var semanticCorpus = []CorpusCase{
	{
		ID: "import/pass/alias-preserved", IndicatorID: "go.import.identity/v1",
		Expected: VerdictPass,
		Facts:    []Fact{{Key: "metadata.complete", Value: "true"}, {Key: "before.binding", Value: "telemetry=example.com/acme/log"}, {Key: "after.binding", Value: "telemetry=example.com/acme/log"}},
	},
	{
		ID: "import/fail/path-base-assumed", IndicatorID: "go.import.identity/v1",
		Expected: VerdictFail,
		Facts:    []Fact{{Key: "metadata.complete", Value: "true"}, {Key: "before.binding", Value: "telemetry=example.com/acme/log"}, {Key: "after.binding", Value: "log=example.com/acme/log"}},
	},
	{
		ID: "import/unknown/package-metadata-missing", IndicatorID: "go.import.identity/v1",
		Expected: VerdictUnknown,
		Facts:    []Fact{{Key: "metadata.complete", Value: "false"}, {Key: "import.path", Value: "example.com/acme/log"}},
	},
	{
		ID: "initialization/pass/graph-and-order-equal", IndicatorID: "go.initialization.order/v1",
		Expected: VerdictPass,
		Facts:    []Fact{{Key: "dependency_graph.equal", Value: "true"}, {Key: "lexical_file_order.equal", Value: "true"}, {Key: "hidden_dependencies", Value: "false"}},
	},
	{
		ID: "initialization/fail/init-functions-reordered", IndicatorID: "go.initialization.order/v1",
		Expected: VerdictFail,
		Facts:    []Fact{{Key: "before.init_order", Value: "initA,initB"}, {Key: "after.init_order", Value: "initB,initA"}, {Key: "observation.complete", Value: "true"}},
	},
	{
		ID: "initialization/unknown/hidden-interface-dependency", IndicatorID: "go.initialization.order/v1",
		Expected: VerdictUnknown,
		Facts:    []Fact{{Key: "dependency_graph.equal", Value: "true"}, {Key: "hidden_dependencies", Value: "unresolved"}},
	},
	{
		ID: "package/pass/all-selected-files-agree", IndicatorID: "go.package.conformance/v1",
		Expected: VerdictPass,
		Facts:    []Fact{{Key: "target_set.complete", Value: "true"}, {Key: "before.packages", Value: "alpha"}, {Key: "after.packages", Value: "alpha"}},
	},
	{
		ID: "package/fail/output-package-mismatch", IndicatorID: "go.package.conformance/v1",
		Expected: VerdictFail,
		Facts:    []Fact{{Key: "target_set.complete", Value: "true"}, {Key: "before.packages", Value: "alpha"}, {Key: "after.packages", Value: "alpha,beta"}},
	},
	{
		ID: "package/unknown/selected-files-unresolved", IndicatorID: "go.package.conformance/v1",
		Expected: VerdictUnknown,
		Facts:    []Fact{{Key: "target_set.complete", Value: "false"}, {Key: "after.packages", Value: "alpha"}},
	},
}
