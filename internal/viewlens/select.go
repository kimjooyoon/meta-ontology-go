package viewlens

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/originlenspolicy"
)

func ResolveView(view string) (string, error) {
	switch view {
	case "", "auto":
		return "", nil
	case "full":
		return Full, nil
	case "inputs":
		return NamedInputs, nil
	}
	if originlenspolicy.LensAvailable(view) {
		return view, nil
	}
	return "", fmt.Errorf("VIEW_UNAVAILABLE")
}

func validateRequest(request Request) error {
	if err := validateIdentity(request.ActorID, request.Audience); err != nil {
		return err
	}
	if !boundedIdentity(request.ActivityID, 512) || !validDigest(request.SourceDigest) {
		return fmt.Errorf("VIEW_CONTEXT_INVALID")
	}
	if _, err := ResolveView(request.View); err != nil {
		return err
	}
	_, err := ResolveView(request.Prefer)
	return err
}

func validateIdentity(actor, audience string) error {
	if !boundedIdentity(actor, 128) || (audience != "agent" && audience != "person") {
		return fmt.Errorf("VIEW_IDENTITY_INVALID")
	}
	return nil
}

func boundedIdentity(value string, limit int) bool {
	return value != "" && len(value) <= limit && utf8.ValidString(value) &&
		strings.IndexFunc(value, unicode.IsControl) < 0
}

func initialSnapshot(actor, audience string) Snapshot {
	return Snapshot{Schema: "gooo/view-lens-hint/v1", Domain: Domain, ActorID: actor, Audience: audience}
}

func choose(snapshot Snapshot, request Request) Choice {
	requested, _ := ResolveView(request.View)
	preferred := snapshot.Preferred
	if request.Prefer != "" {
		preferred, _ = ResolveView(request.Prefer)
	}
	return Choice{ViewID: originlenspolicy.SelectLens(requested, preferred,
		snapshot.FullSelections, snapshot.NamedSelections, request.Audience == "person"),
		Reason: originlenspolicy.SelectionReason(requested, preferred,
			snapshot.FullSelections, snapshot.NamedSelections, request.Audience == "person"),
		ActivityID: request.ActivityID, SourceDigest: request.SourceDigest,
		Requested: requested != "", Policy: originlenspolicy.Origin()}
}

// Select performs no external file I/O, locks, queues, clocks, networks or models.
// Invalid hints cannot turn a valid canonical report into a failed operation.
func Select(request Request, hint []byte) Selection {
	result := Selection{Schema: "gooo/view-lens-selection/v1", Domain: Domain,
		HintState: "DEFAULT", Issues: []string{}, Choice: Choice{ViewID: Full, Reason: "CANONICAL_FALLBACK"}}
	if err := originlenspolicy.VerifyProjection(); err != nil {
		result.HintState, result.Issues = "IGNORED", []string{"HINT_POLICY_UNAVAILABLE"}
		return result
	}
	if err := validateRequest(request); err != nil {
		result.HintState, result.Issues = "IGNORED", []string{err.Error()}
		return result
	}
	snapshot := initialSnapshot(request.ActorID, request.Audience)
	if len(hint) > 0 {
		loaded, err := decodeSnapshot(hint, request.ActorID, request.Audience)
		if err != nil {
			result.HintState, result.Issues = "IGNORED", []string{err.Error()}
		} else {
			snapshot, result.HintState = loaded, "LOADED"
		}
	}
	result.Choice = choose(snapshot, request)
	if result.HintState != "IGNORED" {
		result.Record = &Event{Schema: "gooo/view-lens-event/v1", Request: request,
			Snapshot: snapshot, Choice: result.Choice}
	}
	return result
}
