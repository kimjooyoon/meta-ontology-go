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
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	interventionverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

type receiptFixture struct {
	receipt      Receipt
	inputs       SourceInputs
	repository   string
	subjectSHA   string
	repositoryFS fs.FS
}

func TestProducerReplayAndConsumerBindOneReceipt(t *testing.T) {
	fixture := buildReceiptFixture(t)
	if err := VerifyReceipt(fixture.receipt, fixture.repository, fixture.subjectSHA); err != nil {
		t.Fatal(err)
	}
	if err := VerifySource(fixture.receipt, fixture.inputs, fixture.repositoryFS, fixture.repository, fixture.subjectSHA); err != nil {
		t.Fatal(err)
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
			if err := VerifySource(receipt, inputs, fixture.repositoryFS, fixture.repository, fixture.subjectSHA); err == nil {
				t.Fatal("tampered concept-operation evidence was accepted")
			}
		})
	}
}

func buildReceiptFixture(t *testing.T) receiptFixture {
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
	directory := t.TempDir()
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
	return receiptFixture{receipt: receipt, inputs: SourceInputs{Strategy: strategyPayload, StrategyVerification: strategyVerificationPayload, SourceMetrics: metricsPayload, Intervention: ledgerPayload, InterventionVerification: interventionVerificationPayload, Program: programPayload, ProgramSource: programSource, ProgramVerification: programVerificationPayload}, repository: repository, subjectSHA: subjectSHA, repositoryFS: os.DirFS(root)}
}

func fixtureDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}
