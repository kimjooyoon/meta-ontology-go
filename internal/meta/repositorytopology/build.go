package repositorytopology

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type inspection struct {
	source                                        SourceReport
	identityExact, ontologyExact, rootPolicyExact bool
	fileRowsExact, directoryRowsExact             int
	duplicates, metaBound, bindingWitnesses       int
	unknownDecisions, knownFailClosed             int
	rootTopology, rootREADME                      int
	goFiles, goooFiles, goLines, goooLines        int
	actualLines                                   map[string]int
	lowerResolution                               bool
}

func Evaluate(sourceJSON, rootOntology, bindingOntology []byte, root, repository, head string) (Report, error) {
	first, err := evaluateOnce(sourceJSON, rootOntology, bindingOntology, root, repository, head)
	if err != nil {
		return Report{}, err
	}
	second, err := evaluateOnce(sourceJSON, rootOntology, bindingOntology, root, repository, head)
	if err != nil {
		return Report{}, err
	}
	if !bytes.Equal(mustJSON(first), mustJSON(second)) {
		return Report{}, fmt.Errorf("repository topology replay diverged")
	}
	first.ReplayVerified = true
	seal(&first)
	return first, nil
}

func evaluateOnce(sourceJSON, rootOntology, bindingOntology []byte, root, repository, head string) (Report, error) {
	var source SourceReport
	if err := json.Unmarshal(sourceJSON, &source); err != nil {
		return Report{}, fmt.Errorf("decode source metrics: %w", err)
	}
	s := inspection{source: source, actualLines: map[string]int{}}
	s.inspectIdentity(root, repository, head)
	s.inspectOntologies(rootOntology, bindingOntology)
	s.inspectFiles(root)
	s.inspectDirectories()
	s.inspectMeta()
	report := s.buildReport(sourceJSON, rootOntology, bindingOntology)
	seal(&report)
	return report, nil
}

func (s *inspection) inspectIdentity(root, repository, head string) {
	validHead := len(head) == 40 && strings.Trim(head, "0123456789abcdef") == ""
	if !validHead || s.source.Meta.Schema != "gooo/indicator-report/v3" || s.source.Meta.Policy.Schema != "gooo/source-policy/v1" {
		s.lowerResolution = true
	}
	s.identityExact = validHead && repository != "" && s.source.Repository == repository && s.source.CommitSHA == head
	wantRoot, _ := filepath.Abs(root)
	gotRoot, _ := filepath.Abs(s.source.Root)
	s.identityExact = s.identityExact && filepath.Clean(wantRoot) == filepath.Clean(gotRoot)
}
