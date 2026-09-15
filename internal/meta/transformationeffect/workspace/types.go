package workspace

type Entry struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Mode   uint32 `json:"mode"`
	SHA256 string `json:"sha256"`
	data   []byte
}

type State struct {
	Entries []Entry
	Digest  string
}

type Change struct {
	Path               string `json:"path"`
	Kind               string `json:"kind"`
	BeforeSHA256       string `json:"before_sha256,omitempty"`
	AfterSHA256        string `json:"after_sha256,omitempty"`
	Mode               uint32 `json:"mode"`
	AfterContentBase64 string `json:"after_content_base64,omitempty"`
}

type Patch struct {
	Schema       string   `json:"schema"`
	HeadSHA      string   `json:"head_sha"`
	Changes      []Change `json:"changes"`
	ChangeDigest string   `json:"change_digest"`
	PatchDigest  string   `json:"patch_digest"`
}

type Sandbox struct {
	host   string
	parent string
	Root   string
}
