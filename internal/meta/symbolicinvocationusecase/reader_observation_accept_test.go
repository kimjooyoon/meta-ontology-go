package symbolicinvocationusecase

import "testing"

func TestSymbolicReaderObservationAcceptsExplicitCompilerResult(t *testing.T) {
	input := readerObservationFixture()
	report := EvaluateSymbolicReaderRequest(readerObservationTestSHA, readerObservationJSON(t, input))

	if report.Decision != "PASS" || report.Resolution != SymbolicReaderObservationResolution {
		t.Fatalf("decision=%s resolution=%s", report.Decision, report.Resolution)
	}
	if report.Coordinates.Satisfied != 10 || report.Coordinates.Total != 10 || report.Coordinates.BasisPoints != 10000 {
		t.Fatalf("coordinates=%+v", report.Coordinates)
	}
	if len(report.Classes) != 3 || report.Classes[0].Total != 3 || report.Classes[1].Total != 3 || report.Classes[2].Total != 4 {
		t.Fatalf("classes=%+v", report.Classes)
	}
	if len(report.Proofs) != 3 || report.Proofs[0].Total != 4 || report.Proofs[1].Total != 3 || report.Proofs[2].Total != 3 {
		t.Fatalf("proofs=%+v", report.Proofs)
	}
	if report.Effects.RepositoryWrites != 0 || report.Effects.MutationAuthority || report.PromotionCreditBPS != 0 {
		t.Fatalf("effects=%+v promotion=%d", report.Effects, report.PromotionCreditBPS)
	}
}
