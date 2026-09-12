package conceptoperation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	interventionverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

type receiptFixture struct {
	receipt        Receipt
	inputs         SourceInputs
	repository     string
	subjectSHA     string
	repositoryRoot string
	repositoryFS   fs.FS
}

func TestProducerReplayAndConsumerBindOneReceipt(t *testing.T) {
	fixture := buildReceiptFixture(t)
	if err := VerifyReceipt(fixture.receipt, fixture.repository, fixture.subjectSHA); err != nil {
		t.Fatal(err)
	}
	if fixture.receipt.Status != "VERIFIED" || fixture.receipt.ExpectedCount == 0 || fixture.receipt.BoundCount != fixture.receipt.ExpectedCount || fixture.receipt.CoverageBPS != 10000 {
		t.Fatalf("producer receipt is not complete: %+v", fixture.receipt)
	}
	if err := VerifySource(fixture.receipt, fixture.inputs, fixture.repositoryFS, fixture.repositoryRoot, fixture.repository, fixture.subjectSHA); err != nil {
		t.Fatal(err)
	}
}

func TestVerifiedReceiptJSONKeepsFailureClassBoundary(t *testing.T) {
	fixture := buildReceiptFixture(t)
	payload, err := json.Marshal(fixture.receipt)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatal(err)
	}
	value, present := fields["failure_class"]
	if !present || value != "" {
		t.Fatalf("verified receipt failure_class = %#v, present=%t", value, present)
	}
}

func TestVerifySourceRejectsRepositoryScratch(t *testing.T) {
	fixture := buildReceiptFixture(t)
	inputs := fixture.inputs
	inputs.ScratchDirectory = fixture.repositoryRoot
	if err := VerifySource(fixture.receipt, inputs, fixture.repositoryFS, fixture.repositoryRoot, fixture.repository, fixture.subjectSHA); err == nil {
		t.Fatal("repository scratch directory was accepted")
	}
}

func TestVerifySourceMissingInputFrontierIsDeterministic(t *testing.T) {
	fixture := buildReceiptFixture(t)
	inputs := fixture.inputs
	inputs.Strategy = nil
	inputs.StrategyVerification = nil
	var first string
	for attempt := range 5 {
		err := VerifySource(fixture.receipt, inputs, fixture.repositoryFS, fixture.repositoryRoot, fixture.repository, fixture.subjectSHA)
		if err == nil {
			t.Fatal("missing source inputs were accepted")
		}
		if attempt == 0 {
			first = err.Error()
		} else if err.Error() != first {
			t.Fatalf("missing source error changed: %q != %q", err, first)
		}
	}
}

func TestConsumerRejectsResealedCohortChanges(t *testing.T) {
	fixture := buildReceiptFixture(t)
	cases := []struct {
		name   string
		mutate func(*Receipt)
	}{
		{name: "omitted-direct-missing", mutate: func(receipt *Receipt) {
			receipt.Expected = receipt.Expected[:len(receipt.Expected)-1]
			receipt.ExpectedCount = len(receipt.Expected)
			receipt.ObservedCount = receipt.ExpectedCount - 1
			receipt.BoundCount = receipt.ExpectedCount - 1
			receipt.CoverageBPS = receipt.BoundCount * 10000 / receipt.ExpectedCount
			receipt.Status = "FAIL_CLOSED"
			receipt.FailureClass = "UNKNOWN"
			receipt.Unknown = &UnknownCausal{UnknownClass: "DIRECT_MISSING", Stage: "CONCEPT_OPERATION_BINDING", Step: "RECONSTRUCT_EXACT_COHORT", Reason: "CONCEPT_OPERATION_DIRECT_EVIDENCE_MISSING", NextOperation: "CAPTURE_EXACT_PRODUCER_INPUTS", BlockedBy: []string{}}
		}},
		{name: "replaced-resealed", mutate: func(receipt *Receipt) {
			receipt.Expected[0].Operation = "replaced-operation"
			receipt.Expected[0].IndicatorID = metricstrategy.ConceptOperationIndicatorID("replaced-operation")
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			receipt := fixture.receipt
			receipt.Expected = append([]OperationBinding(nil), receipt.Expected...)
			testCase.mutate(&receipt)
			sealed, err := seal(receipt)
			if err != nil {
				t.Fatal(err)
			}
			if err := VerifySource(sealed, fixture.inputs, fixture.repositoryFS, fixture.repositoryRoot, fixture.repository, fixture.subjectSHA); err == nil {
				t.Fatal("resealed cohort mutation was accepted")
			}
		})
	}
}

func TestConsumerRejectsValidJSONResealedProgramTamper(t *testing.T) {
	fixture := buildReceiptFixture(t)
	inputs := fixture.inputs
	var program metricprogram.Program
	if err := json.Unmarshal(inputs.Program, &program); err != nil {
		t.Fatal(err)
	}
	program.SourceDigest = fixtureDigest("forged-source")
	program.Digest = ""
	program.Digest, _ = artifact.Digest(program)
	inputs.Program, _ = json.Marshal(program)
	var verification programverify.Report
	if err := json.Unmarshal(inputs.ProgramVerification, &verification); err != nil {
		t.Fatal(err)
	}
	verification.ProgramDigest = program.Digest
	verification.SourceDigest = program.SourceDigest
	verification.Digest = ""
	verification.Digest, _ = artifact.Digest(verification)
	inputs.ProgramVerification, _ = json.Marshal(verification)
	if err := VerifySource(fixture.receipt, inputs, fixture.repositoryFS, fixture.repositoryRoot, fixture.repository, fixture.subjectSHA); err == nil {
		t.Fatal("valid JSON resealed program tamper was accepted")
	}
}

func TestConsumerRejectsAlteredInterventionWithClaimedDigest(t *testing.T) {
	fixture := buildReceiptFixture(t)
	inputs := fixture.inputs
	var ledger metric.Ledger
	if err := json.Unmarshal(inputs.Intervention, &ledger); err != nil {
		t.Fatal(err)
	}
	ledger.Indicators[0].Actual = "tampered"
	inputs.Intervention, _ = json.Marshal(ledger)
	if err := VerifySource(fixture.receipt, inputs, fixture.repositoryFS, fixture.repositoryRoot, fixture.repository, fixture.subjectSHA); err == nil {
		t.Fatal("altered intervention with claimed digest was accepted")
	}
}

func TestConsumerRejectsResealedAndTamperedEvidence(t *testing.T) {
	fixture := buildReceiptFixture(t)
	cases := []struct {
		name   string
		mutate func(*Receipt, *SourceInputs)
	}{
		{name: "forged-resealed-receipt", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.Expected[0].Activity = "forged-activity"
			sealed, err := seal(*receipt)
			if err != nil {
				panic(err)
			}
			*receipt = sealed
		}},
		{name: "wrong-head", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.SubjectSHA = strings.Repeat("b", 40)
		}},
		{name: "wrong-source-digest", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.ProgramSourceDigest = fixtureDigest("wrong-source")
		}},
		{name: "wrong-semantic-digest", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.ProgramSemanticDigest = fixtureDigest("wrong-semantic")
		}},
		{name: "wrong-registry-digest", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.ProgramRegistryDigest = fixtureDigest("wrong-registry")
		}},
		{name: "wrong-strategy-digest", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.StrategyDigest = fixtureDigest("wrong-strategy")
		}},
		{name: "wrong-program-digest", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.ProgramDigest = fixtureDigest("wrong-program")
		}},
		{name: "wrong-intervention-digest", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.InterventionDigest = fixtureDigest("wrong-intervention")
		}},
		{name: "duplicate-cohort-entry", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.Expected[1].Operation = receipt.Expected[0].Operation
		}},
		{name: "omitted-cohort-entry", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.Expected = receipt.Expected[:len(receipt.Expected)-1]
			receipt.ExpectedCount = len(receipt.Expected)
		}},
		{name: "replaced-cohort-entry", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.Expected[0].Operation = "replaced-operation"
		}},
		{name: "extra-unknown-entry", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.Expected = append(receipt.Expected, OperationBinding{Operation: "unknown-operation", CarrierOperation: "unknown-operation", IndicatorID: metricstrategy.ConceptOperationIndicatorID("unknown-operation"), Expected: "REGISTERED_CONCEPT", Status: "UNSATISFIED"})
		}},
		{name: "empty-input", mutate: func(_ *Receipt, inputs *SourceInputs) {
			*inputs = SourceInputs{}
		}},
		{name: "writes-authority", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.RepositoryWorkspaceWrites = true
		}},
		{name: "promotion-authority", mutate: func(receipt *Receipt, _ *SourceInputs) {
			receipt.PromotionAuthorized = true
		}},
		{name: "replay-tamper", mutate: func(_ *Receipt, inputs *SourceInputs) {
			inputs.InterventionVerification = append([]byte(nil), inputs.InterventionVerification...)
			inputs.InterventionVerification[len(inputs.InterventionVerification)-1] ^= 1
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			receipt := fixture.receipt
			receipt.Expected = append([]OperationBinding(nil), receipt.Expected...)
			inputs := fixture.inputs
			testCase.mutate(&receipt, &inputs)
			if err := VerifySource(receipt, inputs, fixture.repositoryFS, fixture.repositoryRoot, fixture.repository, fixture.subjectSHA); err == nil {
				t.Fatal("tampered concept-operation evidence was accepted")
			}
		})
	}
}

func buildFreshReceiptFixture(t *testing.T, directory string) receiptFixture {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate repository source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "../../../.."))
	repository := "kimjooyoon/meta-ontology-go"
	subjectSHA := strings.Repeat("a", 40)
	metricsReport, err := linecaps.AnalyzeProjectedLineMetrics(root, root)
	if err != nil {
		t.Fatal(err)
	}
	metricsReport.Repository, metricsReport.CommitSHA = repository, subjectSHA
	metricsPayload, err := json.Marshal(metricsReport)
	if err != nil {
		t.Fatal(err)
	}
	metricsPath := filepath.Join(directory, "source-metrics.json")
	if err := os.WriteFile(metricsPath, metricsPayload, 0o600); err != nil {
		t.Fatal(err)
	}
	ledger, err := metric.Generate(metricsPath, repository, subjectSHA)
	if err != nil {
		t.Fatal(err)
	}
	ledgerPayload, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(directory, "intervention-ledger.json")
	if err := os.WriteFile(ledgerPath, ledgerPayload, 0o600); err != nil {
		t.Fatal(err)
	}
	interventionReceipt, err := interventionverify.Replay(metricsPath, ledger)
	if err != nil {
		t.Fatal(err)
	}
	interventionVerificationPayload, err := json.Marshal(interventionReceipt)
	if err != nil {
		t.Fatal(err)
	}
	interventionVerificationPath := filepath.Join(directory, "intervention-verification.json")
	if err := os.WriteFile(interventionVerificationPath, interventionVerificationPayload, 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := metricstrategy.Generate(os.DirFS(root), metricsPath, ledgerPath, interventionVerificationPath, repository, subjectSHA)
	if err != nil {
		t.Fatal(err)
	}
	strategyPayload, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	strategyReceipt, err := strategyverify.Replay(os.DirFS(root), metricsPath, ledgerPath, interventionVerificationPath, plan)
	if err != nil {
		t.Fatal(err)
	}
	strategyVerificationPayload, err := json.Marshal(strategyReceipt)
	if err != nil {
		t.Fatal(err)
	}
	program, programSource, err := metricprogram.Compile(strategyPayload, strategyVerificationPayload)
	if err != nil {
		t.Fatal(err)
	}
	programPayload, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	programReceipt, err := programverify.Verify(strategyPayload, strategyVerificationPayload, programPayload, programSource)
	if err != nil {
		t.Fatal(err)
	}
	programVerificationPayload, err := json.Marshal(programReceipt)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := Build(strategyPayload, strategyVerificationPayload, ledgerPayload, programPayload, programSource, programVerificationPayload)
	if err != nil {
		t.Fatal(err)
	}
	return receiptFixture{receipt: receipt, inputs: SourceInputs{Strategy: strategyPayload, StrategyVerification: strategyVerificationPayload, SourceMetrics: metricsPayload, Intervention: ledgerPayload, InterventionVerification: interventionVerificationPayload, Program: programPayload, ProgramSource: programSource, ProgramVerification: programVerificationPayload, ScratchDirectory: directory}, repository: repository, subjectSHA: subjectSHA, repositoryRoot: root, repositoryFS: os.DirFS(root)}
}

func fixtureDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}
