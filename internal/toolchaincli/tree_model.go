package toolchaincli

type treeRecord struct {
	Path          string `json:"path"`
	Mode          uint32 `json:"mode"`
	Size          int64  `json:"size"`
	ModifiedNanos int64  `json:"modified_nanos"`
	ContentDigest string `json:"content_digest"`
}

type treeSnapshot struct {
	Digest  string
	Records map[string]treeRecord
}
