package userjourneyscorecard

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
)

func (s *inspection) inspectEvidence(root, executable, head string) error {
	if !reflect.DeepEqual(s.contract, expectedContract()) {
		s.unknowns++
		s.lowerResolution = true
	}
	if err := metacli.Validate(s.upstream, head); err != nil {
		s.unknowns++
		s.lowerResolution = true
	} else {
		s.upstreamPassed = s.upstream.Decision == metacli.DecisionPass && s.upstream.Resolution == metacli.ResolutionExact && s.upstream.Summary.Satisfied == 12
	}
	profileKnown := s.profile.Schema == "gooo/user-journey-resource-observation/v1" &&
		s.profile.SubjectSHA == head && validHead(head) && strings.HasPrefix(s.profile.Runner.GoVersion, "go1.27") &&
		s.profile.Runner.OS != "" && s.profile.Runner.Architecture != "" && s.profile.Runner.Image != ""
	if !profileKnown {
		s.unknowns++
		s.lowerResolution = true
	}
	binaryDigest, binarySize, err := digestFile(executable)
	if err != nil {
		return err
	}
	s.binaryBound = binaryDigest == s.profile.Executable.Digest && binaryDigest == s.upstream.Source.ExecutableDigest && binarySize == s.profile.Executable.SizeBytes
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(s.contract.Source)))
	if err != nil {
		return err
	}
	s.sourceBound = s.profile.SourcePath == s.contract.Source && s.profile.SourceDigest == digestBytes(content)
	if binarySize > s.contract.BinarySizeLimit {
		s.binaryViolations = 1
	}
	s.repositoryWrites = s.upstream.RepositoryWrites
	for _, result := range s.upstream.Cases {
		if result.Definition.MetaOperation != "" && knownProof(result.Definition.ProofChoice) {
			s.metaBindings++
		} else {
			s.unknowns++
			s.lowerResolution = true
		}
	}
	return nil
}

func validHead(value string) bool {
	return len(value) == 40 && strings.Trim(value, "0123456789abcdef") == ""
}

func knownProof(value string) bool {
	return value == "FOUNDATION" || value == "COHERENCE" || value == "REGRESSION"
}
