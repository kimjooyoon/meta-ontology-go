package pressureshadow

func rebindSafeCoverage(input *R4SafeInput) {
	selector := projectR4SafeInput(*input).Selector
	shadow := Input{Schema: SchemaVersion, Selector: selector, PathCoverage: input.PathCoverage}
	for i := range shadow.PathCoverage {
		shadow.PathCoverage[i].SnapshotDigest = selector.SnapshotDigest
		shadow.PathCoverage[i].PolicyDigest = selector.PolicyDigest
		shadow.PathCoverage[i].RegistryDigest = selector.RegistryDigest
		rebindCoverage(&shadow, shadow.PathCoverage[i].PathID)
	}
	input.PathCoverage = shadow.PathCoverage
}
