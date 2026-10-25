// Package workspaceinput reads the external Gooo workspace definition and its
// bounded source set. CLI and native metaprogram consumers share this contract.
package workspaceinput

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

const (
	Schema             = "gooo/package-workspace-manifest/v1"
	MaxDefinitionBytes = 1 << 20
	MaxSourceCount     = 512
	MaxSourceBytes     = 1 << 20
	MaxSourceSetBytes  = 16 << 20
)

type Definition struct {
	Schema   string                   `json:"schema"`
	Entry    packageruntime.EntrySpec `json:"entry"`
	Packages []Package                `json:"packages"`
}

type Package struct {
	Path    string   `json:"path"`
	Name    string   `json:"name"`
	Imports []string `json:"imports"`
	Sources []string `json:"sources"`
}

func Decode(data []byte) (Definition, error) {
	if len(data) > MaxDefinitionBytes {
		return Definition{}, fmt.Errorf("input exceeds maximum size of %d bytes", MaxDefinitionBytes)
	}
	var definition Definition
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, fmt.Errorf("decode workspace manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Definition{}, fmt.Errorf("workspace manifest contains trailing JSON")
		}
		return Definition{}, fmt.Errorf("decode workspace manifest trailing data: %w", err)
	}
	if err := validate(definition); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func validate(definition Definition) error {
	if definition.Schema != Schema {
		return fmt.Errorf("workspace manifest schema must be %q", Schema)
	}
	if len(definition.Packages) == 0 || len(definition.Packages) > 64 {
		return fmt.Errorf("workspace manifest must declare 1 to 64 packages")
	}
	for _, pkg := range definition.Packages {
		if len(pkg.Sources) == 0 || len(pkg.Sources) > MaxSourceCount {
			return fmt.Errorf("package %q must declare 1 to %d source files", pkg.Path, MaxSourceCount)
		}
	}
	return nil
}

func SourcePath(source string) (string, error) {
	if strings.TrimSpace(source) == "" || filepath.IsAbs(source) {
		return "", fmt.Errorf("workspace source path %q must be relative", source)
	}
	clean := filepath.Clean(filepath.FromSlash(source))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workspace source path %q must stay inside the workspace", source)
	}
	if !strings.HasSuffix(clean, ".gooo") && !strings.HasSuffix(clean, ".gooo.fixture") {
		return "", fmt.Errorf("workspace source path %q must name a .gooo source or .gooo.fixture", source)
	}
	return clean, nil
}
