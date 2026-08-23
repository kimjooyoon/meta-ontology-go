package toolchainrelease

import "testing"

func validReportFixture(t *testing.T) Report {
	t.Helper()
	corpus := Corpus{Schema: CorpusSchema, Version: 1, Toolchain: ExpectedToolchain,
		Targets: Targets(), Cases: expectedCases()}
	corpusDigest, err := digestValue(corpus)
	if err != nil {
		t.Fatal(err)
	}
	evidence := make([]PlatformEvidence, 0, TargetCount)
	for _, target := range Targets() {
		receipt := PlatformReceipt{
			Schema: PlatformReceiptSchema, Decision: DecisionPass,
			Reason: "TOOLCHAIN_RELEASE_PLATFORM_READY", Resolution: ResolutionExact,
			HeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Toolchain: ExpectedToolchain,
			Platform: target, ArchiveFormat: target.ArchiveFormat,
			Build: BuildEvidence{VCSRevision: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Trimpath: true, BuildVCS: true},
			Binary: ReplayEvidence{Name: binaryName(target), Digest: "sha256:binary", Bytes: 1, Builds: 2, ReplayEqual: true},
			Archive: ReplayEvidence{Name: archiveName(target), Digest: "sha256:archive", Bytes: 1, Builds: 2, ReplayEqual: true},
			Smoke: SmokeEvidence{SchemaVersion: "gooo-version/v1", Language: "gooo", Version: "0.1.0-dev", Status: "development"},
		}
		receipt, err = FinalizeReceipt(receipt)
		if err != nil {
			t.Fatal(err)
		}
		evidence = append(evidence, PlatformEvidence{Receipt: receipt, ReceiptVerified: true, ArchiveVerified: true})
	}
	report, err := Evaluate(corpus, corpusDigest, evidence,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "sha256:concept", true)
	if err != nil {
		t.Fatal(err)
	}
	return report
}
