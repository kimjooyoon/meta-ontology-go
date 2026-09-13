package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

type upstreamSnapshot struct {
	Manifest string
	Resume []byte
	OriginalDigest string
	Digests map[string]string
}

func snapshotUpstream(filename, directory string) (upstreamSnapshot, error) {
	var result upstreamSnapshot
	data, err := readRegular(filename)
	if err != nil {
		return result, err
	}
	var manifest map[string]json.RawMessage
	if err := decodeUpstream(data, &manifest); err != nil {
		return result, err
	}
	var schema, root string
	var names []string
	if json.Unmarshal(manifest["schema"], &schema) != nil || schema != "gooo/public-self-improvement-orchestration-verification-input/v1" ||
		json.Unmarshal(manifest["published_root"], &root) != nil || !filepath.IsAbs(root) ||
		json.Unmarshal(manifest["published_artifacts"], &names) != nil || len(names) == 0 {
		return result, errors.New("upstream manifest identity or member list is missing")
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return result, err
	}
	members := make(map[string][]byte, len(names))
	result.Digests = make(map[string]string, len(names))
	for _, name := range names {
		if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
			return result, errors.New("upstream member name escapes its declared bundle")
		}
		if _, exists := members[name]; exists {
			return result, errors.New("upstream member name is duplicated")
		}
		member, err := readRegular(filepath.Join(root, name))
		if err != nil {
			return result, err
		}
		members[name] = member
		result.Digests[name] = cache.HashBytes(member).String()
	}
	for key, raw := range manifest {
		if key == "schema" || key == "published_root" || key == "published_artifacts" {
			continue
		}
		var original string
		if err := json.Unmarshal(raw, &original); err != nil || !filepath.IsAbs(original) {
			return result, fmt.Errorf("upstream role %s has no absolute member path", key)
		}
		resolved, err := filepath.EvalSymlinks(original)
		if err != nil {
			return result, err
		}
		name := filepath.Base(resolved)
		if filepath.Dir(resolved) != root {
			return result, fmt.Errorf("upstream role %s points outside the declared bundle", key)
		}
		member, exists := members[name]
		if !exists {
			return result, fmt.Errorf("upstream role %s is not a published member", key)
		}
		if key == "resume_report" {
			result.Resume = member
		}
		manifest[key] = upstreamJSONString(filepath.Join(directory, name))
	}
	if len(result.Resume) == 0 {
		return result, errors.New("upstream manifest does not bind a resume report")
	}
	for name, member := range members {
		if err := writeNew(filepath.Join(directory, name), member, 0o444); err != nil {
			return result, err
		}
	}
	manifest["published_root"] = upstreamJSONString(directory)
	result.Manifest = filepath.Join(directory, "relocated-verification-input.json")
	result.OriginalDigest = cache.HashBytes(data).String()
	return result, writeJSON(result.Manifest, manifest)
}

func upstreamJSONString(value string) json.RawMessage {
	data, _ := json.Marshal(value)
	return data
}
