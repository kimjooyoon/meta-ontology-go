package main

type config struct {
	root        string
	work        string
	expectedSHA string
}

type trackedFile struct {
	logical   string
	kind      string
	language  string
	mode      uint32
	lines     int
	data      []byte
	objectSHA string
}

type storedObject struct {
	id      string
	ext     string
	backing string
	data    []byte
}

type manifestEntry struct {
	Logical    string `json:"logical"`
	Backing    string `json:"backing"`
	ObjectSHA  string `json:"object_sha256"`
	ContentSHA string `json:"content_sha256"`
	Kind       string `json:"kind"`
	Language   string `json:"language,omitempty"`
	Mode       uint32 `json:"mode"`
	Lines      int    `json:"lines,omitempty"`
}

type manifest struct {
	Schema    string          `json:"schema"`
	SourceSHA string          `json:"source_sha"`
	Proof     string          `json:"proof_choice"`
	Authority string          `json:"proof_authority"`
	Entries   []manifestEntry `json:"entries"`
}

type indicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type evidence struct {
	Schema       string      `json:"schema"`
	SourceSHA    string      `json:"source_sha"`
	TrackedFiles int         `json:"tracked_files"`
	Objects      int         `json:"stored_objects"`
	Indicators   []indicator `json:"indicators"`
}
