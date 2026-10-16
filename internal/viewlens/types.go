// Package viewlens selects advisory presentations. It has no compiler authority.
package viewlens

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/originlenspolicy"

const Schema = "gooo/view-lens-profile/v1"
const Domain = "gooo://view-domain/typed-value-origins/v1"
const Full = "gooo://origin-lens/full"
const NamedInputs = "gooo://origin-lens/named-inputs"
const historyLimit = 32
const countLimit = 1_000_000_000
const hintLimit = 4096

type Request struct {
	ActorID      string `json:"actor_id"`
	Audience     string `json:"audience"`
	View         string `json:"view"`
	Prefer       string `json:"prefer,omitempty"`
	ActivityID   string `json:"activity_id"`
	SourceDigest string `json:"source_digest"`
}

type Choice struct {
	ViewID       string                            `json:"view_id"`
	Reason       string                            `json:"reason"`
	ActivityID   string                            `json:"activity_id"`
	SourceDigest string                            `json:"source_digest"`
	Requested    bool                              `json:"explicit_view_request"`
	Policy       originlenspolicy.ProjectionOrigin `json:"policy"`
}

// Snapshot is the bounded hint a host preloads outside the rendering path.
type Snapshot struct {
	Schema          string `json:"schema"`
	Domain          string `json:"domain"`
	ActorID         string `json:"actor_id"`
	Audience        string `json:"audience"`
	Preferred       string `json:"preferred_view,omitempty"`
	FullSelections  int64  `json:"full_selections"`
	NamedSelections int64  `json:"named_input_selections"`
	Revision        int64  `json:"revision"`
	ProfileDigest   string `json:"profile_digest,omitempty"`
}

type Event struct {
	Schema   string   `json:"schema"`
	Request  Request  `json:"request"`
	Snapshot Snapshot `json:"snapshot"`
	Choice   Choice   `json:"choice"`
}

type Selection struct {
	Schema    string   `json:"schema"`
	Domain    string   `json:"domain"`
	Choice    Choice   `json:"choice"`
	HintState string   `json:"hint_state"`
	Issues    []string `json:"issues"`
	Record    *Event   `json:"record_event,omitempty"`
}

type Profile struct {
	Schema          string   `json:"schema"`
	Snapshot        Snapshot `json:"snapshot"`
	LastEventDigest string   `json:"last_event_digest"`
	Recent          []Choice `json:"recent"`
}

type StoreResult struct {
	Schema             string   `json:"schema"`
	State              string   `json:"state"`
	Issue              string   `json:"issue,omitempty"`
	Snapshot           Snapshot `json:"hint"`
	EventDigest        string   `json:"event_digest,omitempty"`
	ArchivePath        string   `json:"archive_path,omitempty"`
	ProfileArchivePath string   `json:"profile_archive_path,omitempty"`
	RawBytes           int      `json:"raw_bytes,omitempty"`
	StoredBytes        int      `json:"stored_bytes,omitempty"`
	Persisted          bool     `json:"persisted"`
}
