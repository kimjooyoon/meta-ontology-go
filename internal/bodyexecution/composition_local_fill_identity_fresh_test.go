package bodyexecution

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestLocalFillIdentity11503(t *testing.T) {
	// Synthetic metadata isolates identity and ambiguity; no producer is run.
	origin := bodycodegen.SourceFillOrigin{Schema: "gooo/source-fill-origin/v1", ActivityID: "urn:identity11503:activity",
		SourceSHA256: digest([]byte("assembly11503")), SelectedBodySHA256: digest([]byte("body11503")), CandidateID: "chosen11503",
		Fragments: []bodycodegen.SourceFillFragment{{SelectedDecoded: bodycodegen.FillOriginRange{Start: 0, End: 12}, Kind: "retained_template"}}}
	local := bodycodegen.BodyLocalOrigin{ActivityID: origin.ActivityID, BodySHA256: origin.SelectedBodySHA256, Start: 3, End: 8}
	step := func(o bodycodegen.SourceFillOrigin) CompositionStep {
		return CompositionStep{Generation: bodycodegen.Result{Report: bodycodegen.Report{
			BodyFill: &bodycodegen.IRBodyFillReceipt{SourceOrigin: &o}}}}
	}
	prior := Composition{Steps: []CompositionStep{step(origin)}, Preparations: []CompositionStep{step(origin)}}
	selected, reason := compositionLocalFillOrigin(prior, local)
	if selected == nil || reason != "" {
		t.Fatal("identical retained assembly became ambiguous", reason)
	}
	changed := origin
	changed.CandidateID = "other11503"
	prior.Preparations[0] = step(changed)
	if selected, reason := compositionLocalFillOrigin(prior, local); selected != nil || reason != "AMBIGUOUS_ASSEMBLY_INPUT" {
		t.Fatal("conflicting assembly selected arbitrarily", selected, reason)
	}
	site := CompositionLocalSite{Name: "value11503", Source: bodycodegen.BodyLocalSource{Construction: selected}}
	site.ExpressionID = localSiteIdentity(site)
	encoded, _ := json.Marshal(site)
	legacy := site
	legacy.Source.Construction = nil
	if localSiteIdentity(legacy) != site.ExpressionID || !equalRetainedLocalSites([]CompositionLocalSite{site}, []CompositionLocalSite{legacy}) {
		t.Fatal("optional explanation changed identity or omitted historical claim")
	}
	bad := site
	copy := *selected
	copy.CandidateID = "tampered11503"
	bad.Source.Construction = &copy
	if equalRetainedLocalSites([]CompositionLocalSite{site}, []CompositionLocalSite{bad}) {
		t.Fatal("present contradictory construction claim accepted")
	}
	after, _ := json.Marshal(site)
	if string(after) != string(encoded) || legacy.Source.Construction != nil {
		t.Fatal("comparison mutated metadata")
	}
	t.Log("synthetic identical origin merged; conflicting origin unresolved; stable site identity; optional absence preserved; present contradiction rejected; programs0/models0")
}
