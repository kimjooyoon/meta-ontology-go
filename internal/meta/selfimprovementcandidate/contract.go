package selfimprovementcandidate

import (
	"bytes"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func compileContract(repository fs.FS, path string) (ContractEvidence, string) {
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return ContractEvidence{}, ReasonContractUnknown
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return ContractEvidence{}, ReasonContractInvalid
	}
	canonical, err := syntax.Format(file)
	if err != nil || !contractDeclarationsKnown(file) {
		return ContractEvidence{}, ReasonContractInvalid
	}
	lines := bytes.Count(raw, []byte{'\n'})
	if len(raw) != 0 && raw[len(raw)-1] != '\n' {
		lines++
	}
	return ContractEvidence{ContractID: ContractID, Path: path,
		Package: file.Package.Name, Namespace: file.Namespace.Name,
		EntityCount: 3, ActivityCount: 2, SourceLines: lines,
		SourceDigest: digestBytes(raw), CanonicalDigest: digestBytes([]byte(canonical))}, ""
}
