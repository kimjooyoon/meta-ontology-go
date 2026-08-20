package adapter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func captureTemp(root string) (TempArtifactSnapshot, error) {
	rootObservation, err := capturePath(root)
	if err != nil {
		return TempArtifactSnapshot{}, err
	}
	entries := make([]FileObservation, 0)
	if rootObservation.Kind == "directory" {
		entries, err = walkTempEntries(root)
		if err != nil {
			return TempArtifactSnapshot{}, err
		}
		currentRoot, err := capturePath(root)
		if err != nil || currentRoot.Lstat != rootObservation.Lstat {
			return TempArtifactSnapshot{}, fmt.Errorf("temp root changed during capture")
		}
	}
	digest, err := digestTempSnapshot(rootObservation.Lstat, entries)
	if err != nil {
		return TempArtifactSnapshot{}, err
	}
	return TempArtifactSnapshot{Root: rootObservation.Lstat, Digest: digest, Entries: entries}, nil
}

func walkTempEntries(root string) ([]FileObservation, error) {
	entries := make([]FileObservation, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		observation, err := capturePath(path)
		if err != nil {
			return err
		}
		observation.Path = filepath.ToSlash(relative)
		entries = append(entries, observation)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

type tempDigestInput struct {
	Root    LstatIdentity     `json:"root"`
	Entries []FileObservation `json:"entries"`
}

func digestTempSnapshot(root LstatIdentity, entries []FileObservation) (string, error) {
	data, err := json.Marshal(tempDigestInput{Root: root, Entries: entries})
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}
