package selfimprovementtransport

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func CompileContract(repository fs.FS, path string) (ContractEvidence, error) {
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return ContractEvidence{}, fmt.Errorf("read transport contract: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return ContractEvidence{}, fmt.Errorf("parse transport contract")
	}
	canonical, err := syntax.Format(file)
	if err != nil || !contractKnown(file) {
		return ContractEvidence{}, fmt.Errorf("transport contract declarations mismatch")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return ContractEvidence{}, fmt.Errorf("lower transport contract: %w", err)
	}
	policy, err := parseResolutionPolicy(&ir)
	if err != nil {
		return ContractEvidence{}, err
	}
	lines := bytes.Count(raw, []byte{'\n'})
	if len(raw) != 0 && raw[len(raw)-1] != '\n' {
		lines++
	}
	return ContractEvidence{
		ContractID: ContractID, Path: path, Package: file.Package.Name,
		Namespace: file.Namespace.Name, EntityCount: len(expectedEntities),
		ActivityCount: len(expectedActivities), ObligationTotal: fixedObligationTotal,
		SourceLines: lines, SourceDigest: digestBytes(raw), CanonicalDigest: digestBytes([]byte(canonical)),
		SemanticDigest: ir.StableHash(), ResolutionPolicy: policy,
	}, nil
}
