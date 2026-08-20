package main

type config struct {
	root        string
	physical    string
	work        string
	expectedSHA string
	index       string
}

type manifest struct {
	Schema    string  `json:"schema"`
	SourceSHA string  `json:"source_sha"`
	Entries   []entry `json:"entries"`
}

type entry struct {
	Logical       string `json:"logical"`
	Backing       string `json:"backing"`
	ObjectSHA     string `json:"object_sha256"`
	ContentSHA    string `json:"content_sha256"`
	Kind          string `json:"kind"`
	Language      string `json:"language,omitempty"`
	Mode          uint32 `json:"mode"`
	Lines         int    `json:"lines"`
}

type indexState struct {
	TreeOID       string
	Replacement   string
	Unbound       int
	Dirty         int
	Unexpected    int
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
	Schema            string      `json:"schema"`
	CurrentSHA        string      `json:"current_sha"`
	LogicalOriginSHA  string      `json:"logical_origin_sha"`
	LogicalTreeOID    string      `json:"logical_tree_oid"`
	ReplacementCommit string      `json:"replacement_commit,omitempty"`
	Entries           int         `json:"entries"`
	Restored           int         `json:"restored"`
	Indicators         []indicator `json:"indicators"`
}

func (item entry) gitMode() string {
	if item.Kind == "symlink" {
		return "120000"
	}
	if item.Mode&0o111 != 0 {
		return "100755"
	}
	return "100644"
}
