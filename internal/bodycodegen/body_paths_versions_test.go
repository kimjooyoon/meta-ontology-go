package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Synthetic fixtures prefer failing paths and force finite feedback. Expanded
// ternary matrices are all zero; their original FP32 output biases stay intact.
func writeVersionedThreeContract(t *testing.T, compact bool, variant, feature, arithmetic string) string {
	t.Helper()
	var name string
	if compact {
		name = writeSharedThreeContractModel(t, variant)
	} else {
		name = writeThreeContractModel(t)
	}
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var meta jointdecision.Metadata
	if err = json.Unmarshal(b, &meta); err != nil {
		t.Fatal(err)
	}
	meta.Feature, meta.Arithmetic = feature, arithmetic
	if !compact && variant != "fp32" {
		weights := filepath.Join(filepath.Dir(name), meta.WeightsFile)
		original, err := os.ReadFile(weights)
		if err != nil {
			t.Fatal(err)
		}
		var packed []byte
		for i := range meta.Tensors {
			tensor := &meta.Tensors[i]
			block := original[tensor.Offset : tensor.Offset+tensor.Bytes]
			if tensor.Name == "w1" || tensor.Name == "w2" {
				for _, value := range block {
					if value != 0 {
						t.Fatal("fixture matrix must be positive zero")
					}
				}
				block = make([]byte, (tensor.Count+4)/5)
				for j := range block {
					block[j] = 121
				} // Five base-3 digits of one, including padding.
				tensor.Encoding = "ternary_base3_5"
			}
			tensor.Offset, tensor.Bytes = int64(len(packed)), int64(len(block))
			packed = append(packed, block...)
		}
		meta.Variant, meta.WeightsSHA = variant, strings.TrimPrefix(digest(packed), "sha256:")
		if err = os.WriteFile(weights, packed, 0600); err != nil {
			t.Fatal(err)
		}
	}
	b, err = json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(name, b, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func assertThreeVersionIdentity(t *testing.T, result Result, info RetainedModelInfo, feature, arithmetic string) {
	t.Helper()
	r := result.Report.BodyPaths
	if !info.Loaded || info.FeatureVersion != feature || info.ArithmeticVersion != arithmetic ||
		r.ModelContext == nil || r.ModelContext.FeatureVersion != feature || r.ModelContext.ArithmeticVersion != arithmetic {
		t.Fatal("model/context identity differs", info, r.ModelContext)
	}
	if r.ModelRetention != nil && *r.ModelRetention != info {
		t.Fatal("retained identity changed")
	}
	check := func(receipt *pathplan.ThreeReceipt) {
		if receipt == nil || receipt.Schema != info.ModelSchema || receipt.Feature != feature || receipt.Arithmetic != arithmetic {
			t.Fatal("initial/feedback version differs", receipt)
		}
	}
	check(r.Search.Selection.Three)
	for _, f := range r.Feedback {
		if f.Three != nil {
			check(f.Three)
		}
	}
	for _, value := range []any{info, r.ModelContext, r.Search.Selection.Three} {
		b, err := json.Marshal(value)
		if err != nil || strings.Contains(string(b), "arithmetic_version") != (arithmetic != "") {
			t.Fatal("legacy omission or explicit arithmetic identity differs", err)
		}
	}
	consumePathCompleteness(t, result.Report.CompletenessReceipt)
}

func TestVersionedThreeNativeContracts(t *testing.T) {
	for _, feature := range []string{jointdecision.ThreeFeatureVersion, jointdecision.ThreeBagFeatureVersion} {
		for _, arithmetic := range []string{"", jointdecision.SeparateArithmeticVersion} {
			for _, variant := range []string{"fp32", "ptq_ternary", "qat_ternary"} {
				for _, compact := range []bool{false, true} {
					t.Run(feature+"/"+arithmetic+"/"+variant+map[bool]string{false: "/expanded", true: "/compact"}[compact], func(t *testing.T) {
						source, doc := threeNativeFixture(t)
						name := writeVersionedThreeContract(t, compact, variant, feature, arithmetic)
						g, err := NewTypedPathGenerator(name)
						if err != nil {
							t.Fatal(err)
						}
						options := TypedPathOptions{StepAttempts: 1, FeedbackRounds: 7, FeedbackUnfixed: true}
						first, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", doc, options)
						if err != nil {
							t.Fatal(err)
						}
						assertThreeVersionIdentity(t, first, g.Info(), feature, arithmetic)
						if first.Report.BodyPaths.Search.Selection.ModelCalls != 7 || first.Report.BodyPaths.FunctionalCompleteness != 100 {
							t.Fatal("synthetic finite TDD accounting differs")
						}
						for _, d := range doc.Plan.Decisions {
							if !strings.Contains(first.Report.BodyPaths.Search.Selection.Three.Input, d.Intent) {
								t.Fatal("full bilingual intent lost")
							}
						}
						direct, err := GenerateWithTypedPathUnfixedFeedback(context.Background(), "fixture.gooo", source, "ThreeTest", doc, name, 1, 7, nil)
						if err != nil || direct.Source != first.Source {
							t.Fatal("direct/retained generation differs", err)
						}
						assertThreeVersionIdentity(t, direct, g.Info(), feature, arithmetic)
						for _, file := range []string{name, filepath.Join(filepath.Dir(name), "weights.bin")} {
							if err = os.Rename(file, file+".moved"); err != nil {
								t.Fatal(err)
							}
						}
						var wg sync.WaitGroup
						for range 4 {
							wg.Go(func() {
								got, e := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", doc, options)
								if e != nil || got.Source != first.Source {
									t.Error("immutable retained request differs", e)
									return
								}
								assertThreeVersionIdentity(t, got, g.Info(), feature, arithmetic)
							})
						}
						wg.Wait()
						canceled, cancel := context.WithCancel(context.Background())
						cancel()
						_, err = g.Generate(canceled, "fixture.gooo", source, "ThreeTest", doc, options)
						var failure *BodyPathError
						if !errors.As(err, &failure) || !errors.Is(err, context.Canceled) || failure.Receipt.Search.Selection.ModelCalls != 0 {
							t.Fatal("canceled versioned request inferred", err)
						}
					})
				}
			}
		}
	}
}

func TestVersionedThreeDeclinesPreserveFullInput(t *testing.T) {
	for _, compact := range []bool{false, true} {
		name := writeVersionedThreeContract(t, compact, "qat_ternary", jointdecision.ThreeBagFeatureVersion, jointdecision.SeparateArithmeticVersion)
		for _, kind := range []string{"initial-overflow", "feedback-overflow", "two", "four"} {
			source, doc := threeNativeDeclineFixture(t, kind)
			if kind == "initial-overflow" {
				for i := range doc.Plan.Decisions {
					doc.Plan.Decisions[i].Intent = "한국어 and English: " + doc.Plan.Decisions[i].Intent
				}
			}
			got, err := GenerateWithTypedPathUnfixedFeedback(context.Background(), "fixture.gooo", source, "ThreeTest", doc, name, 1, 7, nil)
			if err != nil {
				t.Fatal(err)
			}
			r := got.Report.BodyPaths
			if r.ModelContext.FeatureVersion != jointdecision.ThreeBagFeatureVersion || r.ModelContext.ArithmeticVersion != jointdecision.SeparateArithmeticVersion {
				t.Fatal("decline changed model identity")
			}
			if kind == "feedback-overflow" {
				assertThreeFeedbackOverflow(t, r)
			} else {
				assertThreeDeterministicDecline(t, source, doc, got)
			}
			consumePathCompleteness(t, got.Report.CompletenessReceipt)
		}
	}
}

func TestVersionedThreeUnknownMetadataRejected(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, bad := range []string{"feature", "arithmetic"} {
			feature, arithmetic := jointdecision.ThreeBagFeatureVersion, jointdecision.SeparateArithmeticVersion
			if bad == "feature" {
				feature = "unknown-v5"
			} else {
				arithmetic = "unknown-rounding"
			}
			name := writeVersionedThreeContract(t, compact, "fp32", feature, arithmetic)
			if _, err := NewTypedPathGenerator(name); err == nil {
				t.Fatal("unknown contract accepted", bad)
			}
		}
	}
}
