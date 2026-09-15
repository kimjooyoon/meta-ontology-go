package toolchainrelease

var outcomeMetricIDs = []string{
	"gooo.metric.toolchain.cross-platform-release-readiness-bps.v1",
	"gooo.metric.toolchain.cross-platform-release-case-readiness-bps.v1",
	"gooo.metric.toolchain.cross-platform-release-proof-readiness-bps.v1",
}

var driverMetricIDs = []string{
	"gooo.metric.toolchain.cross-platform-release-executed-cases.v1",
	"gooo.metric.toolchain.cross-platform-release-platform-receipts.v1",
	"gooo.metric.toolchain.cross-platform-release-operating-systems.v1",
	"gooo.metric.toolchain.cross-platform-release-architectures.v1",
	"gooo.metric.toolchain.cross-platform-release-binary-builds.v1",
	"gooo.metric.toolchain.cross-platform-release-archive-builds.v1",
	"gooo.metric.toolchain.cross-platform-release-native-smokes.v1",
	"gooo.metric.toolchain.cross-platform-release-binary-replays.v1",
	"gooo.metric.toolchain.cross-platform-release-archive-replays.v1",
	"gooo.metric.toolchain.cross-platform-release-checksum-entries.v1",
	"gooo.metric.toolchain.cross-platform-release-toolchain-bindings.v1",
	"gooo.metric.toolchain.cross-platform-release-vcs-bindings.v1",
	"gooo.metric.toolchain.cross-platform-release-concept-bindings.v1",
	"gooo.metric.toolchain.cross-platform-release-code-bindings.v1",
	"gooo.metric.toolchain.cross-platform-release-metric-bindings.v1",
	"gooo.metric.toolchain.cross-platform-release-use-case-bindings.v1",
}

func MetricIDs() []string {
	result := append([]string(nil), outcomeMetricIDs...)
	result = append(result, driverMetricIDs...)
	return append(result, guardrailMetricIDs...)
}
