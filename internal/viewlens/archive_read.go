package viewlens

import (
	"fmt"
	"path/filepath"
)

type ArchiveRead struct {
	Schema string `json:"schema"`
	State  string `json:"state"`
	Kind   string `json:"kind"`
	Digest string `json:"raw_digest"`
	Issue  string `json:"issue,omitempty"`
	Raw    []byte `json:"raw_bytes_base64,omitempty"`
}

// ReadArchive restores original persisted bytes, including historical policies.
// It grants no permission to execute, accept or replace their source meaning.
func ReadArchive(directory, kind, key string) ArchiveRead {
	result := ArchiveRead{Schema: "gooo/view-lens-archive/v1", State: "UNAVAILABLE", Kind: kind, Digest: key}
	if directory == "" || !validDigest(key) || (kind != "event" && kind != "profile") {
		result.Issue = "ARCHIVE_REQUEST_INVALID"
		return result
	}
	raw, _, err := readPacked(filepath.Join(directory, kind+"s", key[7:]+".gopack"))
	if err == nil && digest(raw) != key {
		err = fmt.Errorf("ARCHIVE_IDENTITY_MISMATCH")
	}
	if err != nil {
		result.Issue = err.Error()
		return result
	}
	result.Raw, result.State = raw, "RESTORED"
	return result
}
