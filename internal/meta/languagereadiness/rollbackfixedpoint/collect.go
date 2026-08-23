package rollbackfixedpoint

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func Collect(guardPath, transformationPath, expectedHead string) Source {
	source := Source{ExpectedHeadSHA: expectedHead}
	guardRaw, err := os.ReadFile(guardPath)
	if err != nil {
		source.CollectionError = err.Error()
		return source
	}
	var guard guardedpromotion.Report
	if err = json.Unmarshal(guardRaw, &guard); err == nil {
		err = guardedpromotion.Validate(guard)
	}
	if err != nil {
		source.CollectionError = fmt.Sprintf("guard: %v", err)
		return source
	}
	source.Guard = GuardEvidence{FileSHA256: digestBytes(guardRaw),
		ReportDigest: guard.ReportDigest, HeadSHA: guard.Source.CurrentHeadSHA,
		Decision: guard.Decision, Reason: guard.Reason, Resolution: guard.Resolution,
		Satisfied: guard.Summary.Satisfied, Total: guard.Summary.Total,
		Unresolved: guard.Summary.Unresolved,
		RepositoryWrites: guard.Summary.RepositoryWrites,
		RepositoryMutationAuthorized: guard.Source.RepositoryMutationAuthorized}
	return collectTransformation(source, transformationPath)
}

func collectTransformation(source Source, path string) Source {
	raw, err := os.ReadFile(path)
	var ledger transformationeffect.Ledger
	if err == nil {
		err = json.Unmarshal(raw, &ledger)
	}
	if err == nil {
		err = transformationeffect.ValidateLedger(ledger)
	}
	if err != nil {
		source.CollectionError = fmt.Sprintf("transformation: %v", err)
		return source
	}
	source.Transformation = TransformationEvidence{FileSHA256: digestBytes(raw),
		LedgerDigest: ledger.LedgerDigest, HeadSHA: ledger.HeadSHA,
		Decision: ledger.Decision, Reason: ledger.Reason, WorkspaceMode: ledger.WorkspaceMode,
		WriteBoundary: ledger.WriteBoundary, Effects: len(ledger.Effects),
		SourceWorkspaceUnchanged: ledger.SourceWorkspaceUnchanged,
		PromotionAuthorized: ledger.PromotionAuthorized}
	source.RepositoryWrites = source.Guard.RepositoryWrites
	return source
}
