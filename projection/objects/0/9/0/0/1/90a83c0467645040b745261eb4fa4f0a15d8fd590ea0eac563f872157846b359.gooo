package provenance

const (
	manifestCommitted = "committed"
	manifestPrepared  = "prepared"
)

type ledgerState struct {
	records []Evidence
	bytes   []byte
	digest  string
	lines   int
}

// ledgerManifest is the commit record for the JSONL materialization. A
// prepared record is a transaction: Base is the last committed state and the
// top-level fields describe the proposed append. A committed record is the
// only metadata state that authorizes the top-level bytes.
type ledgerManifest struct {
	Schema   int                  `json:"schema"`
	Phase    string               `json:"phase"`
	Bytes    int64                `json:"bytes"`
	Lines    int                  `json:"lines"`
	Digest   string               `json:"digest"`
	LastID   string               `json:"last_id,omitempty"`
	LastHash string               `json:"last_hash,omitempty"`
	Data     string               `json:"data"`
	Base     *ledgerManifestState `json:"base,omitempty"`
}
type ledgerManifestState struct {
	Bytes    int64  `json:"bytes"`
	Lines    int    `json:"lines"`
	Digest   string `json:"digest"`
	LastID   string `json:"last_id,omitempty"`
	LastHash string `json:"last_hash,omitempty"`
	Data     string `json:"data"`
}
