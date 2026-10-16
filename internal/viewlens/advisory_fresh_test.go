package viewlens

import (
	"encoding/json"
	"strings"
	"testing"
)

func freshHintRequest() Request {
	return Request{ActorID: "kimjooyoon", Audience: "person", View: "auto",
		ActivityID: "urn:fresh:origin-view:1801", SourceDigest: digest([]byte("original 한글 bytes\r\n"))}
}

func TestAdvisoryOriginHintFresh1801(t *testing.T) {
	request := freshHintRequest()
	cases := []struct {
		name, hint, view, want, state string
		record                        bool
	}{
		{"default-person", "", "auto", NamedInputs, "DEFAULT", true},
		{"explicit-full", "", "full", Full, "DEFAULT", true},
		{"malformed-json", "{", "auto", NamedInputs, "IGNORED", false},
		{"oversize", strings.Repeat(" ", hintLimit+1), "auto", NamedInputs, "IGNORED", false},
		{"unknown-view", "", "missing", Full, "IGNORED", false},
		{"wrong-domain", `{"schema":"wrong"}`, "auto", NamedInputs, "IGNORED", false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			request.View = item.view
			selection := Select(request, []byte(item.hint))
			if selection.Choice.ViewID != item.want || selection.HintState != item.state ||
				(selection.Record != nil) != item.record {
				t.Fatalf("unexpected advisory choice: %+v", selection)
			}
		})
	}
}

func TestAdvisorySnapshotScopeFresh1804(t *testing.T) {
	request := freshHintRequest()
	snapshot := initialSnapshot(request.ActorID, request.Audience)
	snapshot.Revision, snapshot.FullSelections = 7, 7
	snapshot.ProfileDigest = digest([]byte("prior-profile"))
	raw, _ := json.Marshal(snapshot)
	selection := Select(request, raw)
	if selection.Choice.ViewID != Full || selection.Choice.Reason != "MOST_SELECTED_VIEW" || selection.Record == nil {
		t.Fatalf("valid hint not selected: %+v", selection)
	}
	request.ActorID = "other-actor"
	if got := Select(request, raw); got.HintState != "IGNORED" || got.Choice.ViewID != NamedInputs || got.Record != nil {
		t.Fatalf("cross-actor hint accepted: %+v", got)
	}
	request.View, request.Prefer, request.Audience = "auto", "auto", "agent"
	if got := Select(request, nil); got.Choice.ViewID != Full || got.Choice.Reason != "AUDIENCE_DEFAULT" {
		t.Fatalf("agent default: %+v", got)
	}
}
