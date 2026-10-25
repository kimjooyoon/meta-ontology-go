package workspacesplit

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

const (
	maxDiscoveryEntries = 65536
	maxDiscoveryJSON    = 4096
	maxDiscoveryBytes   = 64 << 20
)

// Discovery retains every workspace definition in the supplied filesystem, even
// those not using the selected occurrence. A later change can make one a consumer.
// Paths are canonical, slash-separated addresses relative to that filesystem.
// This observes authored source lists; it neither executes nor discovers Go
// embeds, dynamic readers, or definitions outside the supplied scope.
type Discovery struct {
	Definitions []DefinitionObservation
	Consumers   []string
	JSONFiles   int
	JSONBytes   int
}

// DiscoverFiles scans bounded JSON inputs and selects definitions declaring the
// given source occurrence. No symlink is traversed or silently excluded. Missing
// or noncanonical source addresses and unreadable/ambiguous JSON refuse the whole
// observation, rather than returning a usable prefix. Sequential reads are not
// an atomic filesystem snapshot; publication must repeat this observation.
func DiscoverFiles(repository fs.FS, source string) (Discovery, error) {
	if repository == nil || !fs.ValidPath(source) {
		return Discovery{}, fmt.Errorf("consumer discovery requires a filesystem and a canonical source address")
	}
	jsonFiles, sources, err := discoverAddresses(repository)
	if err != nil {
		return Discovery{}, err
	}
	if !sources[source] {
		return Discovery{}, fmt.Errorf("selected source is absent or not canonical in discovery scope: %s", source)
	}
	result := Discovery{}
	for _, name := range jsonFiles {
		raw, err := readDiscoveryJSON(repository, name)
		if err != nil {
			return Discovery{}, fmt.Errorf("consumer discovery %q: %w", name, err)
		}
		result.JSONFiles++
		result.JSONBytes += len(raw)
		if result.JSONBytes > maxDiscoveryBytes {
			return Discovery{}, fmt.Errorf("consumer discovery JSON exceeds 64 MiB")
		}
		consumer, workspace, err := inspectDiscoveryDefinition(raw, name, source, sources)
		if err != nil {
			return Discovery{}, fmt.Errorf("consumer discovery %q: %w", name, err)
		}
		if workspace {
			result.Definitions = append(result.Definitions, DefinitionObservation{
				ManifestPath: name, SHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(raw)), ByteLength: len(raw)})
		}
		if consumer {
			result.Consumers = append(result.Consumers, name)
		}
	}
	return result, nil
}

func discoverAddresses(repository fs.FS) ([]string, map[string]bool, error) {
	var jsonFiles []string
	sources := map[string]bool{}
	count := 0
	err := fs.WalkDir(repository, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		count++
		if count > maxDiscoveryEntries {
			return fmt.Errorf("consumer discovery exceeds %d directory entries", maxDiscoveryEntries)
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("consumer discovery cannot observe nonregular entry: %s", name)
		}
		if strings.HasSuffix(name, ".json") {
			jsonFiles = append(jsonFiles, name)
			if len(jsonFiles) > maxDiscoveryJSON {
				return fmt.Errorf("consumer discovery exceeds %d JSON candidates", maxDiscoveryJSON)
			}
		}
		if strings.HasSuffix(name, ".gooo") || strings.HasSuffix(name, ".gooo.fixture") {
			sources[name] = true
		}
		return nil
	})
	return jsonFiles, sources, err
}

func readDiscoveryJSON(repository fs.FS, name string) ([]byte, error) {
	file, err := repository.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, workspaceinput.MaxDefinitionBytes+1))
	if err != nil {
		return nil, err
	}
	return raw, completeness.CheckJSON(raw, workspaceinput.MaxDefinitionBytes)
}

func inspectDiscoveryDefinition(raw []byte, name, selected string, sources map[string]bool) (bool, bool, error) {
	if bytes.TrimSpace(raw)[0] != '{' {
		return false, false, nil
	}
	var header struct{ Schema json.RawMessage }
	if err := json.Unmarshal(raw, &header); err != nil {
		return false, false, err
	}
	var schema string
	if json.Unmarshal(header.Schema, &schema) != nil {
		return false, false, nil
	}
	if schema != workspaceinput.Schema {
		if strings.HasPrefix(schema, "gooo/package-workspace-manifest/") {
			return false, true, fmt.Errorf("unsupported workspace schema: %s", schema)
		}
		return false, false, nil
	}
	definition, err := workspaceinput.Decode(raw)
	if err != nil {
		return false, true, err
	}
	consumer, count := false, 0
	for _, pkg := range definition.Packages {
		for _, declared := range pkg.Sources {
			count++
			relative, err := workspaceinput.SourcePath(declared)
			if err != nil || count > workspaceinput.MaxSourceCount {
				return false, true, fmt.Errorf("invalid or excessive source references: %q", declared)
			}
			address := path.Join(path.Dir(name), filepath.ToSlash(relative))
			if !sources[address] {
				return false, true, fmt.Errorf("source reference is absent or not canonical: %s", address)
			}
			consumer = consumer || address == selected
		}
	}
	return consumer, true, nil
}
