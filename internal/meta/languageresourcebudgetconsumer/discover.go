package languageresourcebudgetconsumer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Discovery struct {
	Activity             string `json:"activity"`
	SourceRawDigest      string `json:"source_raw_digest"`
	SourceSemanticDigest string `json:"source_semantic_digest"`
	TargetDigest         string `json:"target_digest"`
}

func DiscoverEntry(directory string) (string, error) {
	value, err := DiscoverSource(directory)
	if err != nil {
		return "", err
	}
	return value.Activity, nil
}

func DiscoverSource(directory string) (Discovery, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return Discovery{}, fmt.Errorf("SOURCE_DIRECTORY_READ_FAILED")
	}
	sources := make([]RawSource, 0, 2)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".gooo" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return Discovery{}, fmt.Errorf("SOURCE_READ_FAILED")
		}
		sources = append(sources, RawSource{Filename: filepath.ToSlash(filepath.Join(directory, entry.Name())), ContentBase64: base64.StdEncoding.EncodeToString(data)})
	}
	if len(sources) != 2 {
		return Discovery{}, fmt.Errorf("SOURCE_FILE_SET_INVALID")
	}
	input := Input{Contract: Contract{SourcePaths: []string{"source-a.gooo", "source-b.gooo"}}, Producer: Producer{SourceFiles: sources, SourceFileCount: len(sources)}}
	meaning, err := reconstructSource(input)
	if err != nil {
		return Discovery{}, err
	}
	return Discovery{Activity: meaning.Activity, SourceRawDigest: meaning.SourceDigest, SourceSemanticDigest: meaning.SemanticDigest, TargetDigest: meaning.TargetDigest}, nil
}

func EncodeDiscovery(value Discovery) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("SOURCE_DISCOVERY_ENCODE_FAILED")
	}
	return append(data, '\n'), nil
}
