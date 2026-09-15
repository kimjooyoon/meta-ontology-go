package selfimprovementloop

const (
	PatchSchema    = "gooo/self-improvement-patch-proposal/v1"
	EvidenceSchema = "gooo/self-improvement-evidence/v1"
	DossierSchema  = "gooo/self-improvement-dossier/v1"
)

func BuildArtifacts(in Input, report Report) Artifacts {
	proposal := PatchProposal{
		Schema: PatchSchema, Scenario: in.Scenario,
		OutputMode:         "caller-owned-temporary-output",
		RepositoryMutation: false, Patch: in.Transformation.Patch,
	}
	evidenceCells := make([]EvidenceRecord, 0, len(report.Cells))
	for _, cell := range report.Cells {
		evidenceCells = append(evidenceCells, EvidenceRecord{
			Cell: cell.Cell, Decision: cell.Decision, Reason: cell.Reason, Unknown: cell.Unknown,
		})
	}
	evidence := EvidenceBundle{
		Schema: EvidenceSchema, Scenario: in.Scenario, SourceDigest: in.SourceDigest,
		ToolchainDigest: in.ToolchainDigest, Decision: report.Decision,
		Cells: evidenceCells, Unknowns: append([]UnknownState(nil), report.Unknowns...),
		Pair: in.Pair, GraphHash: report.GraphHash,
	}
	canonicalEvidence := evidence
	canonicalEvidence.EvidenceDigest = ""
	evidence.EvidenceDigest = digestJSON(canonicalEvidence)
	dossier := Dossier{
		Schema: DossierSchema, Scenario: in.Scenario, SourceDigest: in.SourceDigest,
		ToolchainDigest: in.ToolchainDigest, Decision: report.Decision,
		GraphHash: report.GraphHash, ReportDigest: report.ReportDigest,
		PatchProposal: proposal, EvidenceDigest: evidence.EvidenceDigest,
		Unknowns: append([]UnknownState(nil), report.Unknowns...),
	}
	canonicalDossier := dossier
	canonicalDossier.DossierDigest = ""
	dossier.DossierDigest = digestJSON(canonicalDossier)
	return Artifacts{Report: report, PatchProposal: proposal, Evidence: evidence, Dossier: dossier}
}
