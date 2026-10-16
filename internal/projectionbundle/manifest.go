// Package projectionbundle binds an owned generated projection across physical files.
package projectionbundle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

const Schema = "gooo/owned-projection-bundle/v1"

type Member struct {
	Path      string   `json:"path"`
	Digest    string   `json:"digest"`
	RegionIDs []string `json:"region_ids"`
}

type Manifest struct {
	Schema       string   `json:"schema"`
	SourceDigest string   `json:"source_digest"`
	Members      []Member `json:"members"`
}

func Digest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }

func Decode(raw []byte) (Manifest, error) {
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return manifest, fmt.Errorf("projection manifest has trailing data")
	}
	if manifest.Schema != Schema || !validDigest(manifest.SourceDigest) || len(manifest.Members) == 0 {
		return manifest, fmt.Errorf("projection manifest identity is invalid")
	}
	paths, identities := map[string]bool{}, map[string]bool{}
	for _, member := range manifest.Members {
		if !fs.ValidPath(member.Path) || strings.ContainsAny(member.Path, "/\\*?[]") ||
			!strings.HasSuffix(member.Path, ".generated.go") || paths[member.Path] ||
			!validDigest(member.Digest) || len(member.RegionIDs) == 0 {
			return manifest, fmt.Errorf("projection member identity is invalid")
		}
		paths[member.Path] = true
		for _, identity := range member.RegionIDs {
			if identity == "" || identities[identity] {
				return manifest, fmt.Errorf("projection region identity is empty or duplicated")
			}
			identities[identity] = true
		}
	}
	return manifest, nil
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	raw, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(raw) == sha256.Size && strings.ToLower(value) == value
}
