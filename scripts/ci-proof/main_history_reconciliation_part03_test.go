package main

import (
	"fmt"
	"reflect"
	"sort"
)

func validateReconciliationLedger(ledger reconciliationLedger) error {
	if ledger.Schema != "gooo/main-history-reconciliation/v1" || ledger.Version != 1 || !validReconciliationSHA(ledger.CandidateMergeSHA) || !validReconciliationSHA(ledger.CandidateFirstParentSHA) || !validReconciliationSHA(ledger.CandidateSecondParentSHA) || !validReconciliationSHA(ledger.DevSHA) || !validReconciliationSHA(ledger.MainSHA) || !validReconciliationSHA(ledger.MergeBaseSHA) {
		return fmt.Errorf("reconciliation schema or topology SHA is malformed")
	}
	if ledger.CandidateMergeSHA != "6a95e73a391a114d8a22cdc23251e8a7623d7768" || ledger.DevSHA != "3b1647c7aa14558d07f868d649071a2c36e6cc88" || ledger.MainSHA != "c557daf1fd6748b2e61afca10fb632792683061f" || ledger.MergeBaseSHA != "0e838bff165dc73ed984ecbdca338d957b8abfa9" {
		return fmt.Errorf("reconciliation live topology SHA evidence is not exact")
	}
	if ledger.CandidateFirstParentSHA != ledger.DevSHA || ledger.CandidateSecondParentSHA != ledger.MainSHA || !ledger.TargetInvariant.MainIsAncestorOfCandidate {
		return fmt.Errorf("candidate parent topology is not exact")
	}
	if ledger.TargetInvariant.FutureMainPromotion != "exact_ci_backed_fast_forward_plus_main_dev_equality" {
		return fmt.Errorf("future promotion invariant is not exact")
	}
	expectedCommits := []reconciliationCommit{
		{
			SHA:     "a798c8be290fd7f96662edca173cd93cfe270482",
			Message: "chore: require Go 1.26.6",
			ResultingGoMod: reconciliationGoModEvidence{
				Path:        "go.mod",
				BlobSHA:     "f5d57a1b2e3c869aab3427dd11519f81fbcf31f8",
				Size:        57,
				SHA256:      "609503ca851a755543781fc951f391aff8d2c120aba175d862b1b7aeef14e236",
				GoDirective: "go 1.26.6",
			},
			Transition: "TRANSIENT_GO_1_26_6_BUMP",
		},
		{
			SHA:     "c557daf1fd6748b2e61afca10fb632792683061f",
			Message: "chore: use Go 1.26.5 toolchain",
			ResultingGoMod: reconciliationGoModEvidence{
				Path:        "go.mod",
				BlobSHA:     "0eaaa03f3d837fafab8ba1959b70adca21deeacd",
				Size:        57,
				SHA256:      "a2b49728e00943685ad402def409ffee32e15329a120b831ed58521616bed84b",
				GoDirective: "go 1.26.5",
			},
			Transition: "FINAL_GO_1_26_5_TIP_EQUIVALENT_TO_DEV",
		},
	}
	if !reflect.DeepEqual(ledger.MainOnlyCommits, expectedCommits) {
		return fmt.Errorf("main-only commit evidence is not exact")
	}
	if !sort.StringsAreSorted(ledger.MainOnlyPaths) || len(ledger.MainOnlyPaths) != 1 || ledger.MainOnlyPaths[0] != "go.mod" {
		return fmt.Errorf("main-only path inventory is not the exact sorted residual set")
	}
	expectedMergeBaseBlob := reconciliationGoModEvidence{
		Path:        "go.mod",
		BlobSHA:     "9b16c0cd3711c2444ebff1a28ad193c31f06be22",
		Size:        55,
		SHA256:      "af4562214f5d56647f7b953c1a74587286bfb1f97c31d3bd8403eda047186323",
		GoDirective: "go 1.23",
	}
	if !reflect.DeepEqual(ledger.MergeBaseBlob, expectedMergeBaseBlob) {
		return fmt.Errorf("merge-base go.mod evidence is not exact")
	}
	blob := ledger.BlobEquivalence
	if blob.Path != "go.mod" || !validReconciliationSHA(blob.DevBlobSHA) || blob.DevBlobSHA != blob.MainBlobSHA || blob.DevSize != 57 || blob.MainSize != 57 || !validReconciliationDigest("sha256:"+blob.DevSHA256) || blob.DevSHA256 != blob.MainSHA256 || blob.DevGoDirective != "go 1.26.5" || blob.MainGoDirective != "go 1.26.5" {
		return fmt.Errorf("go.mod blob equivalence is not exact")
	}
	if ledger.EquivalenceBasis != "FINAL_MAIN_TIP_VS_DEV_TIP_TREE_EQUIVALENCE_ONLY" {
		return fmt.Errorf("equivalence basis is not final-tip-only")
	}
	if ledger.Disposition != "TREE_EQUIVALENT_ALREADY_IN_DEV" || ledger.UnincorporatedMaterial {
		return fmt.Errorf("reconciliation disposition is unsafe")
	}
	if ledger.CanonicalDigest != reconciliationCanonicalDigest(ledger) {
		return fmt.Errorf("reconciliation canonical digest does not match")
	}
	return nil
}
