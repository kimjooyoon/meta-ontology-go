package workspacesplit

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// Request is an explicit instance of the source-owned SplitGoooSections input.
// All addresses are relative to the exact repository revision being executed.
// The native route always includes that repository's syntax registration.
type Request struct {
	Workspace     string `json:"workspace"`
	Package       string `json:"package"`
	Source        string `json:"source"`
	Into          string `json:"into"`
	DiscoveryRoot string `json:"discovery_root"`
	Cases         string `json:"cases"`
	LineLimit     int    `json:"line_limit"`
}

func (request Request) Subject() string {
	return path.Join(path.Dir(request.Workspace), request.Source)
}

func (request Request) Validate() error {
	for _, name := range []string{request.Workspace, request.Source, request.Into, request.Cases, request.DiscoveryRoot} {
		if !fs.ValidPath(name) || strings.ContainsAny(name, "\\\x00:") {
			return fmt.Errorf("workspace split addresses must be canonical repository-relative paths")
		}
	}
	if request.Workspace == "." || request.Cases == "." || request.Package == "" ||
		strings.TrimSpace(request.Package) != request.Package || request.LineLimit <= 0 ||
		path.Ext(request.Source) != ".gooo" || path.Ext(request.Into) != ".gooo" || request.Source == request.Into {
		return fmt.Errorf("workspace split needs a workspace, package, distinct source/output, cases and positive limit")
	}
	if request.DiscoveryRoot != "." && !strings.HasPrefix(request.Workspace, request.DiscoveryRoot+"/") {
		return fmt.Errorf("primary workspace must belong to the explicit consumer discovery scope")
	}
	return nil
}

func (request Request) Digest() string {
	raw, _ := json.Marshal(request)
	return fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
}

func DecodeRequests(raw []byte) ([]Request, error) {
	if err := completeness.CheckJSON(raw, 1<<20); err != nil {
		return nil, err
	}
	var input struct {
		Schema   string    `json:"schema"`
		Requests []Request `json:"requests"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return nil, err
	}
	if input.Schema != "gooo/workspace-split-inputs/v1" || len(input.Requests) == 0 || len(input.Requests) > MaxOwners {
		return nil, fmt.Errorf("workspace split input requires a supported schema and 1..%d requests", MaxOwners)
	}
	for _, request := range input.Requests {
		if err := request.Validate(); err != nil {
			return nil, err
		}
	}
	return input.Requests, nil
}
