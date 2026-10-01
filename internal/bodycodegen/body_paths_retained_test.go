package bodycodegen

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestRetainedGeneratorOwnsModelAndFreshRequestState(t *testing.T) {
	source, document := typedPathFixture(t)
	path := writeTypedPathContractModel(t, false)
	g, err := NewTypedPathGenerator(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(path, path+".moved"); err != nil {
		t.Fatal(err)
	}
	info := g.Info()
	if !info.Loaded || info.MetadataSHA256 == "" || info.WeightsSHA256 == "" || info.ResidentTensorBytes <= 0 {
		t.Fatal("missing retained identity")
	}
	options := TypedPathOptions{StepAttempts: 1, FeedbackRounds: 3, FeedbackUnfixed: true}
	first, err := g.Generate(context.Background(), "fixture.gooo", source, "Combined", document, options)
	if err != nil {
		t.Fatal(err)
	}
	p := first.Report.BodyPaths
	if p.ModelRetention == nil || *p.ModelRetention != info || p.Timing.ModelLoadMS != 0 ||
		!p.SourceBaseMatched || p.Search.Selection.ModelCalls == 0 || !first.Report.TypecheckPassed {
		t.Fatal("retained native contract differs")
	}
	p.ModelRetention.Scope = "caller mutation"
	badSource := []byte(strings.Replace(string(source), "input + 2", "input + 99", 1))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, input := range []struct {
		ctx    context.Context
		source []byte
	}{{context.Background(), badSource}, {ctx, source}} {
		_, err := g.Generate(input.ctx, "fixture.gooo", input.source, "Combined", document, options)
		var failure *BodyPathError
		if !errors.As(err, &failure) || failure.Receipt.Search.Selection.ModelCalls != 0 {
			t.Fatal("request failed to reject before prediction")
		}
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			result, err := g.Generate(context.Background(), "fixture.gooo", source, "Combined", document, options)
			if err != nil || result.Source != first.Source || result.Report.BodyPaths.FunctionalCompleteness != 100 ||
				*result.Report.BodyPaths.ModelRetention != info || result.Report.RepositoryWrites != 0 {
				t.Error("concurrent request inherited state or changed body")
			}
		})
	}
	group.Wait()
}

func TestRetainedGeneratorDisconnectedAndInvalidModes(t *testing.T) {
	source, document := typedPathFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil || g.Info().Loaded {
		t.Fatal("disconnected model unexpectedly loaded")
	}
	result, err := g.Generate(context.Background(), "fixture.gooo", source, "Combined", document, TypedPathOptions{})
	if err != nil || result.Report.BodyPaths.Search.Selection.ModelCalls != 0 || !result.Report.DeterministicReplay {
		t.Fatal("disconnected construction changed")
	}
	for _, options := range []TypedPathOptions{{StepAttempts: -1}, {StepAttempts: 65},
		{FeedbackRounds: 17}, {FeedbackUnfixed: true}, {StepAttempts: 1, FeedbackRounds: 1}} {
		if _, err := g.Generate(context.Background(), "fixture.gooo", source, "Combined", document, options); err == nil {
			t.Fatal("invalid retained configuration accepted")
		}
	}
	if _, err := NewTypedPathGenerator(writeTypedPathContractModel(t, true)); err == nil {
		t.Fatal("operator model accepted as structural model")
	}
}
