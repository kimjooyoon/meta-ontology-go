package operationconformance

var topologyCorpus = []CorpusCase{
	{
		ID: "atomic/pass/same-directory-rename", IndicatorID: "filesystem.atomic-replacement/v1",
		Expected: VerdictPass,
		Facts:    []Fact{{Key: "receipt.present", Value: "true"}, {Key: "temp.same_directory", Value: "true"}, {Key: "commit.rename", Value: "true"}},
	},
	{
		ID: "atomic/fail/truncate-target", IndicatorID: "filesystem.atomic-replacement/v1",
		Expected: VerdictFail,
		Facts:    []Fact{{Key: "receipt.present", Value: "true"}, {Key: "commit.rename", Value: "false"}, {Key: "target.truncated", Value: "true"}},
	},
	{
		ID: "atomic/unknown/no-receipt", IndicatorID: "filesystem.atomic-replacement/v1",
		Expected: VerdictUnknown,
		Facts:    []Fact{{Key: "receipt.present", Value: "false"}},
	},
	{
		ID: "filename/pass/equivalent-target-set", IndicatorID: "go.filename.build-semantics/v1",
		Expected: VerdictPass,
		Facts:    []Fact{{Key: "target_set.complete", Value: "true"}, {Key: "before.matches", Value: "linux/amd64,darwin/arm64"}, {Key: "after.matches", Value: "linux/amd64,darwin/arm64"}},
	},
	{
		ID: "filename/fail/os-suffix-broadened", IndicatorID: "go.filename.build-semantics/v1",
		Expected: VerdictFail,
		Facts:    []Fact{{Key: "target_set.complete", Value: "true"}, {Key: "before.matches", Value: "linux/*"}, {Key: "after.matches", Value: "*/*"}},
	},
	{
		ID: "filename/unknown/incomplete-target-set", IndicatorID: "go.filename.build-semantics/v1",
		Expected: VerdictUnknown,
		Facts:    []Fact{{Key: "target_set.complete", Value: "false"}, {Key: "custom_tags.observed", Value: "false"}},
	},
	{
		ID: "header/pass/exact-prefix", IndicatorID: "go.header.preserved/v1",
		Expected: VerdictPass,
		Facts:    []Fact{{Key: "before.header", Value: "license-a"}, {Key: "after.header", Value: "license-a"}},
	},
	{
		ID: "header/fail/license-dropped", IndicatorID: "go.header.preserved/v1",
		Expected: VerdictFail,
		Facts:    []Fact{{Key: "before.header", Value: "license-a"}, {Key: "after.header", Value: ""}},
	},
	{
		ID: "header/unknown/before-missing", IndicatorID: "go.header.preserved/v1",
		Expected: VerdictUnknown,
		Facts:    []Fact{{Key: "before.observed", Value: "false"}, {Key: "after.header", Value: "license-a"}},
	},
}
