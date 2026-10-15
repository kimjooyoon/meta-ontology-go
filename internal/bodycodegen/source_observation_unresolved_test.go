package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestSourceObserveUnresolvedResultsRemainReplayable(t *testing.T) {
	ctx := context.Background()
	for _, source := range []string{strings.Replace(sourceObserveFixture, `  oracle "Reference"`, "", 1),
		strings.Replace(sourceObserveFixture, `max_candidates "2"`, `max_candidates "1"`, 1)} {
		doc, err := DecodeSourcePathDocument(ctx, "unresolved.gooo", []byte(source), "Fee", nil)
		if err != nil {
			t.Fatal(err)
		}
		r, err := GenerateWithTypedPaths(ctx, "unresolved.gooo", []byte(source), "Fee", doc, "")
		if err != nil {
			t.Fatal(err)
		}
		if pathResolved(r.Report.BodyPaths) {
			t.Fatal("unresolved alternatives disappeared")
		}
		if err := VerifyTypedPathProjection(ctx, "unresolved.gooo", []byte(source), doc, r); err != nil {
			t.Fatal("an unresolved observation lost its replayable frontier", err)
		}
	}
}
