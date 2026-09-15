package artifactemit

import "fmt"

func validateSymbolicValueArtifact(input symbolicValueArtifactInput, subjectSHA string) error {
	if !validSymbolicValueHexDigest(subjectSHA, 20) {
		return fmt.Errorf("subject sha must be 40 lowercase hexadecimal characters")
	}
	if input.Schema != "gooo/symbolic-invocation-schema-artifact/v1" || input.Decision != "PASS" {
		return fmt.Errorf("symbolic artifact identity is not accepted")
	}
	if !validSymbolicValueSHA256(input.Digest) {
		return fmt.Errorf("symbolic artifact digest is not sha256")
	}
	conformance := input.Conformance
	if conformance.Schema != "gooo/symbolic-invocation-conformance/v1" ||
		conformance.Decision != "PASS" || conformance.Resolution != "STRUCTURAL_ONLY" {
		return fmt.Errorf("symbolic conformance identity is not accepted")
	}
	if conformance.GeneratedVectors != 2 || len(conformance.Vectors) != 2 {
		return fmt.Errorf("symbolic conformance must expose exactly two generated vectors")
	}
	if conformance.EmbeddedHandwrittenVectors != 0 {
		return fmt.Errorf("embedded handwritten vectors are not compiler authority")
	}
	if conformance.Effects.RepositoryWrites != 0 || conformance.Effects.MutationAuthority {
		return fmt.Errorf("symbolic conformance must remain read-only")
	}
	return nil
}
