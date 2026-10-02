package bodycodegen

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

// Controlled compact weights deliberately prefer failing masks. These fixtures
// check compiler TDD and receipt semantics; they do not measure learned quality.
func writeSharedThreeContractModel(t *testing.T, variant string) string {
	t.Helper()
	meta := jointdecision.Metadata{Schema: jointdecision.SharedThreeSchema, Feature: jointdecision.ThreeFeatureVersion,
		Variant: variant, FeatureDim: 768, HiddenDim: 8, MaxBytes: 1600, Temperature: 1, WeightsFile: "weights.bin"}
	var raw []byte
	for i, name := range [3]string{"w1", "b1", "w2"} {
		rows, cols := [3]int{8, 1, 2}[i], [3]int{256, 8, 8}[i]
		count, encoding := rows*cols, "float32_le"
		values := make([]float32, count)
		for j := range values {
			if i == 1 {
				values[j] = 1
			}
			if i == 2 && j >= 8 {
				values[j] = -1
			}
		}
		var block []byte
		if variant == "fp32" || i == 1 {
			block = make([]byte, 4*count)
			for j, v := range values {
				binary.LittleEndian.PutUint32(block[4*j:], math.Float32bits(v))
			}
		} else {
			encoding = "ternary_base3_5"
			block = make([]byte, (count+4)/5)
			for j := range block {
				power := 1
				for k := range 5 {
					digit := 1
					if j*5+k < len(values) {
						digit = int(values[j*5+k]) + 1
					}
					block[j] += byte(digit * power)
					power *= 3
				}
			}
		}
		meta.Tensors = append(meta.Tensors, decision.TensorMetadata{Name: name, Rows: rows, Cols: cols,
			Count: count, Encoding: encoding, Offset: int64(len(raw)), Bytes: int64(len(block)), Scale: 1})
		raw = append(raw, block...)
	}
	for i := range 8 {
		meta.Labels = append(meta.Labels, fmt.Sprintf("mask_%d", i))
	}
	meta.WeightsSHA = strings.TrimPrefix(digest(raw), "sha256:")
	dir := t.TempDir()
	name := filepath.Join(dir, "model.json")
	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(name, b, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "weights.bin"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestCompactSharedNativeTDDRetainedMemoryAndConcurrentCalls(t *testing.T) {
	for _, variant := range []string{"fp32", "ptq_ternary", "qat_ternary"} {
		t.Run(variant, func(t *testing.T) {
			source, doc := threeNativeFixture(t)
			model := writeSharedThreeContractModel(t, variant)
			g, err := NewTypedPathGenerator(model)
			if err != nil {
				t.Fatal(err)
			}
			resident := 8288
			if variant != "fp32" {
				resident = 2096
			}
			info := g.Info()
			if info.ModelSchema != jointdecision.SharedThreeSchema || info.ResidentTensorBytes != resident || !info.Loaded {
				t.Fatal("compact retained schema/storage differs", info)
			}
			for _, name := range []string{model, filepath.Join(filepath.Dir(model), "weights.bin")} {
				if err = os.Rename(name, name+".moved"); err != nil {
					t.Fatal(err)
				}
			}
			options := TypedPathOptions{StepAttempts: 1, FeedbackRounds: 7, FeedbackUnfixed: true}
			first, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", doc, options)
			if err != nil {
				t.Fatal(err)
			}
			r := first.Report.BodyPaths
			consumePathCompleteness(t, first.Report.CompletenessReceipt)
			if r.FunctionalCompleteness != 100 || r.Search.Selection.ModelCalls != 7 ||
				r.Search.Selection.Three.Schema != jointdecision.SharedThreeSchema || r.Timing.ModelLoadMS != 0 ||
				*r.ModelRetention != info || first.Report.RepositoryWrites != 0 {
				t.Fatal("compact native body, accounting or retained model differs")
			}
			for _, f := range r.Feedback {
				if f.ModelCalls > 0 && (f.Three == nil || f.Three.Schema != jointdecision.SharedThreeSchema) {
					t.Fatal("feedback artifact schema changed")
				}
			}
			oracle := "package main\nimport \"fmt\"\n" + withoutPackage(t, first.Source) + "\nfunc main() {\n"
			var expected []string
			for _, c := range doc.TestCases {
				oracle += fmt.Sprintf("fmt.Println(ThreeTest(%d))\n", c.Input)
				expected = append(expected, fmt.Sprint(c.Expected))
			}
			if got := runBodyFillGoOracle(t, oracle+"}\n"); strings.TrimSpace(got) != strings.Join(expected, "\n") {
				t.Fatal("compact generated Go differs from independent arithmetic", got)
			}
			var wg sync.WaitGroup
			for range 4 {
				wg.Go(func() {
					result, e := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", doc, options)
					if e != nil {
						t.Error(e)
						return
					}
					if result.Source != first.Source || result.Report.BodyPaths.Search.Selection.ModelCalls != 7 {
						t.Error("compact concurrent request shares mutable state")
					}
				})
			}
			wg.Wait()
			seeded := doc
			seeded.Seed = "compact-artifact-replay"
			a, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", seeded, options)
			if err != nil {
				t.Fatal(err)
			}
			b, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", seeded, options)
			if err != nil || a.Source != b.Source ||
				a.Report.BodyPaths.Search.Selection.Three.Sampled != b.Report.BodyPaths.Search.Selection.Three.Sampled {
				t.Fatal("compact same-artifact seed replay differs", err)
			}
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			_, err = g.Generate(canceled, "fixture.gooo", source, "ThreeTest", doc, options)
			var failure *BodyPathError
			if !errors.As(err, &failure) || !errors.Is(err, context.Canceled) || failure.Receipt.Search.Selection.ModelCalls != 0 {
				t.Fatal("canceled compact request inferred", err)
			}
		})
	}
}

func TestCompactSharedNativeFullBoundsAndUnsupportedArity(t *testing.T) {
	model := writeSharedThreeContractModel(t, "qat_ternary")
	for _, kind := range []string{"feedback-overflow", "initial-overflow", "two", "four"} {
		t.Run(kind, func(t *testing.T) {
			source, doc := threeNativeDeclineFixture(t, kind)
			result, err := GenerateWithTypedPathUnfixedFeedback(context.Background(), "fixture.gooo", source,
				"ThreeTest", doc, model, 1, 7, nil)
			if err != nil {
				t.Fatal(err)
			}
			consumePathCompleteness(t, result.Report.CompletenessReceipt)
			if kind == "feedback-overflow" {
				assertThreeFeedbackOverflow(t, result.Report.BodyPaths)
				for _, f := range result.Report.BodyPaths.Feedback[:6] {
					if f.Three.Schema != jointdecision.SharedThreeSchema {
						t.Fatal("overflow schema differs")
					}
				}
			} else {
				assertThreeDeterministicDecline(t, source, doc, result)
			}
		})
	}
}
