package toolchainrelease

type CorpusCase struct {
	ID       string `json:"id"`
	TargetID string `json:"target_id,omitempty"`
	Kind     string `json:"kind"`
	Expected string `json:"expected"`
}

func expectedCases() []CorpusCase {
	cases := make([]CorpusCase, 0, CaseCount)
	for _, target := range targetRegistry {
		prefix := target.ID + "-"
		cases = append(cases,
			CorpusCase{ID: prefix + "receipt", TargetID: target.ID, Kind: "RECEIPT", Expected: "EXACT_RECEIPT"},
			CorpusCase{ID: prefix + "native-runtime", TargetID: target.ID, Kind: "NATIVE_RUNTIME", Expected: target.GOOS + "/" + target.GOARCH},
			CorpusCase{ID: prefix + "go127-toolchain", TargetID: target.ID, Kind: "TOOLCHAIN", Expected: ExpectedToolchain},
			CorpusCase{ID: prefix + "binary-replay", TargetID: target.ID, Kind: "BINARY_REPLAY", Expected: "2_BYTE_EQUAL_BUILDS"},
			CorpusCase{ID: prefix + "archive-replay", TargetID: target.ID, Kind: "ARCHIVE_REPLAY", Expected: "2_BYTE_EQUAL_ARCHIVES"},
			CorpusCase{ID: prefix + "version-smoke", TargetID: target.ID, Kind: "VERSION_SMOKE", Expected: "gooo-version/v1"},
		)
	}
	return append(cases,
		CorpusCase{ID: "release-set-completeness", Kind: "RELEASE_SET", Expected: "3_OF_3_TARGETS"},
		CorpusCase{ID: "release-checksum-manifest", Kind: "CHECKSUM_MANIFEST", Expected: "3_SORTED_SHA256_ENTRIES"},
	)
}
