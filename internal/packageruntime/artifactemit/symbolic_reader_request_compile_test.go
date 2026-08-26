package artifactemit

import "testing"

func TestCompileSymbolicReaderRequest(t *testing.T) {
	projection := symbolicReaderRequestProjectionFixture()
	result, err := CompileSymbolicReaderRequest(
		symbolicReaderRequestSourceFixture("USER", "DECISION_AND_COUNTS_ONLY"),
		encodeSymbolicReaderRequestProjection(t, projection),
		"fixture-sha",
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != "PASS" || result.Resolution != "GOOO_REQUEST_BOUND_ONLY" {
		t.Fatalf("decision=%s resolution=%s", result.Decision, result.Resolution)
	}
	if result.Coordinates.Satisfied != 12 || result.Coordinates.Total != 12 {
		t.Fatalf("coordinates=%+v", result.Coordinates)
	}
	if result.View.Audience != "USER" || len(result.View.IndicatorIDs) != 5 {
		t.Fatalf("view=%+v", result.View)
	}
	if result.Effects.RepositoryWrites != 0 || result.Effects.MutationAuthority ||
		result.PromotionCreditBPS != 0 {
		t.Fatalf("effects=%+v promotion=%d", result.Effects, result.PromotionCreditBPS)
	}
	if !symbolicReaderValidDigest(result.Digest) {
		t.Fatalf("digest=%q", result.Digest)
	}
}
