package selfimprovementattestation

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func readArchive(filename string) (string, Producer, string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", Producer{}, "", err
	}
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return "", Producer{}, "", err
	}
	defer reader.Close()
	producerData, err := readUniqueEntry(&reader.Reader, "producer.json")
	if err != nil {
		return "", Producer{}, "", err
	}
	observation, err := readUniqueEntry(&reader.Reader, "first.json")
	if err != nil {
		return "", Producer{}, "", err
	}
	var producer Producer
	if err := json.Unmarshal(producerData, &producer); err != nil {
		return "", Producer{}, "", err
	}
	return digestBytes(data), producer, digestBytes(observation), nil
}

func readUniqueEntry(reader *zip.Reader, name string) ([]byte, error) {
	var match *zip.File
	for _, file := range reader.File {
		if filepath.Base(file.Name) != name {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("archive entry %q is ambiguous", name)
		}
		match = file
	}
	if match == nil {
		return nil, fmt.Errorf("archive entry %q is missing", name)
	}
	stream, err := match.Open()
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	return io.ReadAll(stream)
}
