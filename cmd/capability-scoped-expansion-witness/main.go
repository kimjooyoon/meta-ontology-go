package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	expansion "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion"
	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/verify"
)

type options struct {
	source, output, subject, pinnedFile, logicalInput, sandbox, repository, providerJSON, artifactRoot string
}

type judgeReport struct {
	Schema                     string                `json:"schema"`
	SourceDigest               string                `json:"source_digest"`
	SemanticDigest             string                `json:"semantic_digest"`
	ProviderDigest             string                `json:"provider_digest"`
	SubjectSHA                 string                `json:"subject_sha"`
	Decision                   string                `json:"decision"`
	SourceReconstruction       string                `json:"source_reconstruction"`
	SourceReconstructionPasses int                   `json:"source_reconstruction_passes"`
	SourceReconstructionTotal  int                   `json:"source_reconstruction_total"`
	ConsumerReplayPasses       int                   `json:"consumer_replay_passes"`
	ConsumerReplayTotal        int                   `json:"consumer_replay_total"`
	ProducerImportNumerator    int                   `json:"producer_import_numerator"`
	ProducerImportDenominator  int                   `json:"producer_import_denominator"`
	DependencyObservation      dependencyObservation `json:"dependency_observation"`
	Judgements                 []independent.Verdict `json:"judgements"`
}

type dependencyObservation struct {
	ConsumerPackage       string   `json:"consumer_package"`
	ConsumerDependencies  []string `json:"consumer_dependencies"`
	ProducerPackage       string   `json:"producer_package"`
	ProducerImports       int      `json:"producer_imports"`
	ProducerImportTotal   int      `json:"producer_import_total"`
	EffectPackage         string   `json:"effect_package"`
	BrokerPackage         string   `json:"broker_package"`
	EngineEffectsImports  int      `json:"engine_effects_imports"`
	BrokerDependencyFound bool     `json:"broker_dependency_found"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	options, err := parseOptions()
	if err != nil {
		return err
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return err
	}
	if err := expansion.ValidateShape(source); err != nil {
		return err
	}
	if err := requireOutsideRepository(options.output, options.repository); err != nil {
		return err
	}
	providerRaw, err := os.ReadFile(options.providerJSON)
	if err != nil {
		return fmt.Errorf("read raw provider artifact: %w", err)
	}
	if err := writeFile(filepath.Join(options.output, "provider-observations.json"), providerRaw); err != nil {
		return err
	}
	context := expansion.EvaluationContext{ArtifactRoot: options.artifactRoot}
	judgeContext := independent.ProviderContext{RepositoryRoot: options.repository, PinnedFile: options.pinnedFile, LogicalInputPath: options.logicalInput, SandboxRoot: options.sandbox}
	suite, receipts, judgements, err := evaluateAndJudge(source, providerRaw, options.subject, context, judgeContext)
	if err != nil {
		return err
	}
	for index := range suite.Cases {
		suite.Cases[index].IndependentJudge = judgements[index].Status
		suite.Cases[index].IndependentReason = judgements[index].Reason
		if judgements[index].Status != "PASS" {
			return fmt.Errorf("independent judge rejected %s: %s", suite.Cases[index].CaseID, judgements[index].Reason)
		}
	}
	suite.IndependentJudge = "PASS"
	dependency, err := inspectDependencies(options.repository)
	if err != nil {
		return err
	}
	applyIndependentMetrics(&suite, judgements, dependency)
	if suite.Summary.CasesPassed != suite.Summary.CasesTotal || suite.Summary.CasesTotal != expansion.FixedCaseTotal {
		return fmt.Errorf("capability-scoped expansion suite failed: %d/%d", suite.Summary.CasesPassed, suite.Summary.CasesTotal)
	}
	if err := writeReceipts(options.output, suite, receipts); err != nil {
		return err
	}
	interventions, err := writeInterventions(filepath.Join(options.output, "interventions"), source, providerRaw, options.subject, suite, receipts, context, judgeContext)
	if err != nil {
		return err
	}
	suite.Denominator.Interventions = len(interventions)
	if suite.Denominator.Interventions != expansion.FixedInterventionTotal {
		return fmt.Errorf("intervention denominator is %d, want %d", suite.Denominator.Interventions, expansion.FixedInterventionTotal)
	}
	suite = expansion.SealSuite(suite)
	if err := writeJSON(filepath.Join(options.output, "suite.json"), suite); err != nil {
		return err
	}
	judge := judgeReport{
		Schema: "gooo/capability-scoped-expansion-independent-judge/v3", SourceDigest: suite.SourceDigest, SemanticDigest: suite.SemanticDigest,
		ProviderDigest: digest(providerRaw), SubjectSHA: options.subject, Decision: "PASS", SourceReconstruction: "PASS",
		SourceReconstructionPasses: suite.Summary.SourceReconstructionPasses, SourceReconstructionTotal: suite.Summary.SourceReconstructionTotal,
		ConsumerReplayPasses: suite.Summary.ConsumerReplayPasses, ConsumerReplayTotal: suite.Summary.ConsumerReplayTotal,
		ProducerImportNumerator: suite.Summary.ProducerImportNumerator, ProducerImportDenominator: suite.Summary.ProducerImportDenominator,
		DependencyObservation: dependency, Judgements: judgements,
	}
	if err := writeJSON(filepath.Join(options.output, "independent-judge.json"), judge); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(options.output, "dependency-observation.json"), dependency); err != nil {
		return err
	}
	return writeSummary(filepath.Join(options.output, "summary.md"), suite)
}

func evaluateAndJudge(source, providerRaw []byte, subject string, context expansion.EvaluationContext, judgeContext independent.ProviderContext) (expansion.Suite, []expansion.Receipt, []independent.Verdict, error) {
	suite, receipts, err := expansion.EvaluateSuiteWithContext(source, providerRaw, subject, context)
	if err != nil {
		return expansion.Suite{}, nil, nil, err
	}
	judgements := make([]independent.Verdict, 0, len(receipts))
	for index, receipt := range receipts {
		raw, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			return expansion.Suite{}, nil, nil, err
		}
		judgements = append(judgements, independent.JudgeWithContext(source, providerRaw, append(raw, '\n'), judgeContext))
		suite.Cases[index].ClaimStatus = claimStatus(receipt, "capability-scope-exact")
	}
	return suite, receipts, judgements, nil
}

func applyIndependentMetrics(suite *expansion.Suite, judgements []independent.Verdict, dependency dependencyObservation) {
	suite.Summary.ConsumerReplayTotal = len(judgements)
	suite.Summary.ConsumerReplayPasses = 0
	for _, judgement := range judgements {
		if judgement.Status == "PASS" {
			suite.Summary.ConsumerReplayPasses++
		}
	}
	suite.Summary.SourceReconstructionTotal = 1
	if len(judgements) > 0 && suite.Summary.ConsumerReplayPasses == len(judgements) {
		suite.Summary.SourceReconstructionPasses = 1
	}
	suite.Summary.ProducerImportNumerator = dependency.ProducerImports
	suite.Summary.ProducerImportDenominator = dependency.ProducerImportTotal
}

func writeInterventions(outputDir string, source, providerRaw []byte, subject string, base expansion.Suite, baseReceipts []expansion.Receipt, context expansion.EvaluationContext, judgeContext independent.ProviderContext) ([]expansion.Intervention, error) {
	policySource := bytes.Replace(source, []byte("authorization=exact-current"), []byte("authorization=deny-all"), 1)
	if bytes.Equal(policySource, source) {
		return nil, fmt.Errorf("policy intervention did not change a semantic value")
	}
	commentSource := append(append([]byte(nil), source...), []byte("\n// comment-only intervention\n")...)
	graphSource := bytes.Replace(source, []byte("AuthorizeBeforeExpand(ExpansionPolicy)"), []byte("AuthorizeBeforeExpand(ExpansionRequest)"), 1)
	if bytes.Equal(graphSource, source) {
		return nil, fmt.Errorf("graph intervention did not remove the policy edge")
	}
	policy, err := interventionRecord("policy-deny-all", "semantic-policy", source, policySource, providerRaw, subject, base, baseReceipts, context, judgeContext, false)
	if err != nil {
		return nil, err
	}
	comment, err := interventionRecord("comment-only", "comment", source, commentSource, providerRaw, subject, base, baseReceipts, context, judgeContext, true)
	if err != nil {
		return nil, err
	}
	graph, err := interventionRecord("graph-edge-removal", "graph", source, graphSource, providerRaw, subject, base, baseReceipts, context, judgeContext, false)
	if err != nil {
		return nil, err
	}
	forged, forgedRaw, forgedReceiptRaw, err := forgedProviderIntervention(source, providerRaw, subject, base, baseReceipts, judgeContext)
	if err != nil {
		return nil, err
	}
	items := []expansion.Intervention{policy, comment, graph, forged}
	for _, item := range items {
		if err := writeJSON(filepath.Join(outputDir, item.ID+".json"), item); err != nil {
			return nil, err
		}
	}
	if err := writeFile(filepath.Join(outputDir, "forged-provider-input.json"), append(forgedRaw, '\n')); err != nil {
		return nil, err
	}
	if err := writeFile(filepath.Join(outputDir, "forged-provider-resealed-receipt.json"), append(forgedReceiptRaw, '\n')); err != nil {
		return nil, err
	}
	return items, nil
}

func interventionRecord(id, kind string, baseSource, changedSource, providerRaw []byte, subject string, base expansion.Suite, baseReceipts []expansion.Receipt, context expansion.EvaluationContext, judgeContext independent.ProviderContext, preserveDecision bool) (expansion.Intervention, error) {
	changedContext := context
	changedContext.ArtifactRoot = filepath.Join(context.ArtifactRoot, "interventions", id)
	changed, changedReceipts, judgements, err := evaluateAndJudge(changedSource, providerRaw, subject, changedContext, judgeContext)
	if err != nil {
		return expansion.Intervention{}, err
	}
	for _, judgement := range judgements {
		if judgement.Status != "PASS" {
			return expansion.Intervention{}, fmt.Errorf("intervention %s failed independent judge: %s", id, judgement.Reason)
		}
	}
	baseCase := caseResult(base, "allow-current-file-time")
	changedCase := caseResult(changed, "allow-current-file-time")
	if baseCase == nil || changedCase == nil {
		return expansion.Intervention{}, fmt.Errorf("intervention %s cannot find allow case", id)
	}
	baseReceipt := receiptForReceipts(baseReceipts, "allow-current-file-time")
	changedReceipt := receiptForReceipts(changedReceipts, "allow-current-file-time")
	if baseReceipt == nil || changedReceipt == nil {
		return expansion.Intervention{}, fmt.Errorf("intervention %s cannot find allow receipt", id)
	}
	decisionPreserved := baseCase.ObservedDecision == changedCase.ObservedDecision
	semanticPreserved := base.SemanticDigest == changed.SemanticDigest
	outputDigestPreserved := baseReceipt.Artifact.ContentDigest == changedReceipt.Artifact.ContentDigest
	propositionsPreserved := propositionDigest(baseReceipt.Propositions) == propositionDigest(changedReceipt.Propositions)
	tokenPreserved := tokenDigest(baseReceipt.TokenAttempts) == tokenDigest(changedReceipt.TokenAttempts)
	if preserveDecision != decisionPreserved || (preserveDecision && (!semanticPreserved || !outputDigestPreserved || !propositionsPreserved || !tokenPreserved)) || (!preserveDecision && (decisionPreserved || baseCase.ClaimStatus == changedCase.ClaimStatus)) {
		return expansion.Intervention{}, fmt.Errorf("intervention %s did not satisfy its preservation/change contract", id)
	}
	return expansion.Intervention{
		ID: id, Kind: kind, BaseSourceDigest: base.SourceDigest, ChangedSourceDigest: changed.SourceDigest, BaseSemanticDigest: base.SemanticDigest, ChangedSemanticDigest: changed.SemanticDigest,
		BaseDecision: baseCase.ObservedDecision, ChangedDecision: changedCase.ObservedDecision, BaseClaimStatus: baseCase.ClaimStatus, ChangedClaimStatus: changedCase.ClaimStatus,
		DecisionPreserved: decisionPreserved, SemanticDigestPreserved: semanticPreserved, IndependentJudge: "PASS",
		BaseArtifactPresent: baseReceipt.Artifact.Present, ChangedArtifactPresent: changedReceipt.Artifact.Present, BaseArtifactDigest: baseReceipt.Artifact.ContentDigest, ChangedArtifactDigest: changedReceipt.Artifact.ContentDigest,
		BaseOutputDigest: baseReceipt.Artifact.ContentDigest, ChangedOutputDigest: changedReceipt.Artifact.ContentDigest, OutputDigestPreserved: outputDigestPreserved, PropositionsPreserved: propositionsPreserved, TokenDecisionPreserved: tokenPreserved,
		BaseExecutionClaim: baseReceipt.Execution.ClaimState, ChangedExecutionClaim: changedReceipt.Execution.ClaimState,
		GraphComplete: baseReceipt.Graph.Complete, ChangedGraphComplete: changedReceipt.Graph.Complete,
	}, nil
}

func forgedProviderIntervention(source, providerRaw []byte, subject string, base expansion.Suite, baseReceipts []expansion.Receipt, judgeContext independent.ProviderContext) (expansion.Intervention, []byte, []byte, error) {
	var forgedProvider expansion.ProviderObservation
	if err := json.Unmarshal(providerRaw, &forgedProvider); err != nil {
		return expansion.Intervention{}, nil, nil, err
	}
	if len(forgedProvider.FileReads) != 1 {
		return expansion.Intervention{}, nil, nil, fmt.Errorf("cannot forge provider without pinned file observation")
	}
	forgedProvider.FileReads[0].ContentDigest = "sha256:forged-provider-observation"
	forgedRaw, err := json.MarshalIndent(forgedProvider, "", "  ")
	if err != nil {
		return expansion.Intervention{}, nil, nil, err
	}
	allow := receiptForReceipts(baseReceipts, "allow-current-file-time")
	if allow == nil {
		return expansion.Intervention{}, nil, nil, fmt.Errorf("cannot forge provider without allow receipt")
	}
	forgedReceipt := *allow
	forgedReceipt.ProviderDigest = digest(forgedRaw)
	forgedReceipt = expansion.SealReceipt(forgedReceipt)
	forgedReceiptRaw, err := json.MarshalIndent(forgedReceipt, "", "  ")
	if err != nil {
		return expansion.Intervention{}, nil, nil, err
	}
	verdict := independent.JudgeWithContext(source, forgedRaw, append(forgedReceiptRaw, '\n'), judgeContext)
	if verdict.Status == "PASS" {
		return expansion.Intervention{}, nil, nil, fmt.Errorf("forged provider was accepted by independent judge")
	}
	return expansion.Intervention{ID: "forged-provider", Kind: "provider-evidence", BaseSourceDigest: base.SourceDigest, ChangedSourceDigest: base.SourceDigest, BaseSemanticDigest: base.SemanticDigest, ChangedSemanticDigest: base.SemanticDigest, BaseDecision: expansion.DecisionAllow, ChangedDecision: verdict.Decision, BaseClaimStatus: claimStatus(*allow, "capability-scope-exact"), ChangedClaimStatus: "OPEN", DecisionPreserved: false, SemanticDigestPreserved: true, IndependentJudge: "PASS", BaseArtifactPresent: allow.Artifact.Present, ChangedArtifactPresent: false, BaseArtifactDigest: allow.Artifact.ContentDigest, ChangedArtifactDigest: "", BaseOutputDigest: allow.Artifact.ContentDigest, ChangedOutputDigest: "", OutputDigestPreserved: false, PropositionsPreserved: false, TokenDecisionPreserved: false, BaseExecutionClaim: allow.Execution.ClaimState, ChangedExecutionClaim: "OPEN", GraphComplete: allow.Graph.Complete, ChangedGraphComplete: allow.Graph.Complete, ForgedProviderRejected: true, ForgedProviderDigest: digest(forgedRaw), ForgedReceiptDigest: forgedReceipt.ReportDigest, ForgedJudgeReason: verdict.Reason}, forgedRaw, forgedReceiptRaw, nil
}

func caseResult(suite expansion.Suite, id string) *expansion.CaseResult {
	for index := range suite.Cases {
		if suite.Cases[index].CaseID == id {
			return &suite.Cases[index]
		}
	}
	return nil
}

func receiptForReceipts(receipts []expansion.Receipt, id string) *expansion.Receipt {
	for index := range receipts {
		if receipts[index].CaseID == id {
			return &receipts[index]
		}
	}
	return nil
}

func propositionDigest(value []expansion.Proposition) string { return digestJSON(value) }
func tokenDigest(value []expansion.TokenIssuance) string     { return digestJSON(value) }

func inspectDependencies(repository string) (dependencyObservation, error) {
	consumer := "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/verify"
	producer := "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion"
	effects := "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/effects"
	broker := "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/broker"
	dependencies, err := goListDeps(repository, "./internal/meta/capabilityscopedexpansion/verify")
	if err != nil {
		return dependencyObservation{}, err
	}
	engineDependencies, err := goListDeps(repository, "./internal/meta/capabilityscopedexpansion/engine")
	if err != nil {
		return dependencyObservation{}, err
	}
	effectDependencies, err := goListDeps(repository, "./internal/meta/capabilityscopedexpansion/effects")
	if err != nil {
		return dependencyObservation{}, err
	}
	return dependencyObservation{ConsumerPackage: consumer, ConsumerDependencies: dependencies, ProducerPackage: producer, ProducerImports: countExact(dependencies, producer), ProducerImportTotal: 1, EffectPackage: effects, BrokerPackage: broker, EngineEffectsImports: countExact(engineDependencies, effects), BrokerDependencyFound: countExact(effectDependencies, broker) == 1}, nil
}

func goListDeps(repository, packagePattern string) ([]string, error) {
	command := exec.Command("go", "list", "-deps", packagePattern)
	command.Dir = repository
	raw, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go list -deps %s: %w", packagePattern, err)
	}
	lines := strings.Fields(string(raw))
	sort.Strings(lines)
	return lines, nil
}

func countExact(values []string, wanted string) int {
	for _, value := range values {
		if value == wanted {
			return 1
		}
	}
	return 0
}

func writeReceipts(output string, suite expansion.Suite, receipts []expansion.Receipt) error {
	allowWritten, denyWritten, unknownWritten := false, false, false
	for index, receipt := range receipts {
		if err := writeJSON(filepath.Join(output, "receipts", suite.Cases[index].CaseID+"-receipt.json"), receipt); err != nil {
			return err
		}
		switch receipt.Decision {
		case expansion.DecisionAllow:
			if !allowWritten {
				if err := writeJSON(filepath.Join(output, "allow-receipt.json"), receipt); err != nil {
					return err
				}
				allowWritten = true
			}
		case expansion.DecisionDeny:
			if !denyWritten {
				if err := writeJSON(filepath.Join(output, "deny-receipt.json"), receipt); err != nil {
					return err
				}
				denyWritten = true
			}
		case expansion.DecisionUnknown:
			if !unknownWritten {
				if err := writeJSON(filepath.Join(output, "unknown-receipt.json"), receipt); err != nil {
					return err
				}
				unknownWritten = true
			}
		}
	}
	return nil
}

func parseOptions() (options, error) {
	result := options{}
	flags := flag.NewFlagSet("capability-scoped-expansion-witness", flag.ContinueOnError)
	flags.StringVar(&result.source, "source", "", "single Gooo source used by every case")
	flags.StringVar(&result.output, "output-dir", "", "directory outside the repository for receipts")
	flags.StringVar(&result.subject, "subject-sha", "", "exact CI subject SHA")
	flags.StringVar(&result.pinnedFile, "pinned-file", "", "CI-created file observed by the provider")
	flags.StringVar(&result.logicalInput, "logical-input", "", "CI-created deterministic logical input")
	flags.StringVar(&result.sandbox, "sandbox", "", "temporary sandbox")
	flags.StringVar(&result.repository, "repository-root", "", "repository root")
	flags.StringVar(&result.providerJSON, "provider-json", "", "raw provider artifact created by provider command")
	flags.StringVar(&result.artifactRoot, "artifact-root", "", "directory outside repository for expansion artifacts")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return result, err
	}
	if result.source == "" || result.output == "" || result.subject == "" || result.pinnedFile == "" || result.logicalInput == "" || result.sandbox == "" || result.repository == "" || result.providerJSON == "" || result.artifactRoot == "" {
		return result, fmt.Errorf("source, output-dir, subject-sha, pinned-file, logical-input, sandbox, repository-root, provider-json, and artifact-root are required")
	}
	return result, nil
}

func requireOutsideRepository(path, repository string) error {
	target, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output-dir must be outside repository: %s", path)
	}
	return nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(raw, '\n'))
}

func writeFile(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func writeSummary(path string, suite expansion.Suite) error {
	content := fmt.Sprintf("# Capability-scoped expansion\n\n- denominator: cases `%d`, declarations `%d`, capability requests `%d`, evidence slots `%d`, effect token requests `%d`, claims `%d`, interventions `%d`; indicators per receipt `%d`\n- source reconstruction: `%d/%d`; consumer replay: `%d/%d`; producer import numerator/denominator: `%d/%d`\n- source digests: raw `%s`; semantic `%s`\n- cases: `%d/%d` (ALLOW `%d`, DENY `%d`, UNKNOWN `%d`)\n- capability requests: `%d`; authorized `%d`, denied `%d`, UNKNOWN `%d`\n- CURRENT_EVIDENCE capabilities: `%d/%d`; HISTORICAL_FIXTURE declarations: `%d`\n- broker effect token requests/issued/denied: `%d/%d/%d`; artifacts executed/absent for blocked `%d/%d`\n- repository snapshot writes `%d`; sandbox writes `%d`; mutation authority `%s`; promotion authority `%s`\n- independent judge: `%s`\n", suite.Denominator.Cases, suite.Denominator.Declarations, suite.Denominator.CapabilityRequests, suite.Denominator.EvidenceSlots, suite.Denominator.EffectTokenRequests, suite.Denominator.Claims, suite.Denominator.Interventions, suite.Denominator.IndicatorsPerReceipt, suite.Summary.SourceReconstructionPasses, suite.Summary.SourceReconstructionTotal, suite.Summary.ConsumerReplayPasses, suite.Summary.ConsumerReplayTotal, suite.Summary.ProducerImportNumerator, suite.Summary.ProducerImportDenominator, suite.SourceDigest, suite.SemanticDigest, suite.Summary.CasesPassed, suite.Summary.CasesTotal, suite.Summary.AllowCases, suite.Summary.DenyCases, suite.Summary.UnknownCases, suite.Summary.CapabilityRequests, suite.Summary.CapabilityAuthorized, suite.Summary.CapabilityDenied, suite.Summary.CapabilityUnknown, suite.Summary.CurrentEvidenceCapabilities, suite.Summary.CurrentEvidenceDenominator, suite.Summary.HistoricalFixtureDeclarations, suite.Summary.EffectTokenRequests, suite.Summary.TokensIssued, suite.Summary.TokenDenials, suite.Summary.ArtifactExecutions, suite.Summary.ArtifactsAbsentForBlocked, suite.Summary.RepositoryWrites, suite.Summary.SandboxWrites, suite.Summary.MutationAuthority, suite.Summary.PromotionAuthority, suite.IndependentJudge)
	return writeFile(path, []byte(content))
}

func digest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digest(raw)
}

func claimStatus(receipt expansion.Receipt, id string) string {
	for _, claim := range receipt.Claims {
		if claim.ID == id {
			return claim.Status
		}
	}
	return ""
}
