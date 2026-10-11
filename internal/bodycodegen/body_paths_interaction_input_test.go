package bodycodegen

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestInteractionContractExportsEveryConditionAndEmptyStream(t *testing.T) {
	name := writeInteractionContractModel(t, contractdecision.MeanPooling)
	checkContractConditionStreams(t, name, checkInteractionContractExport)
}

func checkContractConditionStreams(t *testing.T, name string, check func(*testing.T, TypedPathContextExport, pathplan.Document)) {
	t.Helper()
	for _, count := range []int{0, 1, 128} {
		source := interactionConditionsSource(count)
		doc := declaredContractDocument(t, source)
		before, err := ExportTypedPathModelContext(context.Background(), "conditions.gooo", source, "Choose", doc, name, "")
		if err != nil {
			t.Fatal(err)
		}
		check(t, before, doc)
		rows := before.ContractConditions
		if rows.Count != count || len(rows.Features) != count || !reflect.DeepEqual(before.Context.ContractConditions, &rows.ContractConditionContext) {
			t.Fatal("condition rows or empty-channel identity lost", count)
		}
		h := sha256.New()
		h.Write([]byte(decision.DeclaredConditionFeatureVersion + "\x00"))
		if err := binary.Write(h, binary.LittleEndian, uint32(count)); err != nil {
			t.Fatal(err)
		}
		for _, row := range rows.Features {
			if err := binary.Write(h, binary.LittleEndian, row); err != nil {
				t.Fatal(err)
			}
		}
		if rows.FeatureSHA != fmt.Sprintf("sha256:%x", h.Sum(nil)) {
			t.Fatal("condition framing differs from the ranking ABI")
		}
	}
}

func interactionConditionsSource(count int) []byte {
	var source strings.Builder
	for line := range strings.SplitSeq(string(conditionModelSource()), "\n") {
		if strings.HasPrefix(line, " condition_case ") {
			continue
		}
		if strings.HasPrefix(line, " attempts ") {
			for i := range count {
				x := int64(i) + 9007199254740993
				fmt.Fprintf(&source, " condition_case \"comparison\" input \"%d\" -> \"true\"\n", x)
			}
		}
		source.WriteString(line + "\n")
	}
	return []byte(source.String())
}

func TestInteractionContractGoalAndOperandChangesHaveSeparateInputs(t *testing.T) {
	name := writeInteractionContractModel(t, contractdecision.MeanPooling)
	original := string(conditionModelSource())
	variants := []string{original,
		strings.Replace(original, `input "9007199254740995" -> "true"`, `input "9007199254740995" -> "false"`, 1),
		strings.Replace(original, `if input < 0`, `if 0 < input`, 1),
		strings.Replace(original, `return 0 - input`, `return input - 0`, 1),
	}
	var exports []TypedPathContextExport
	for _, source := range variants {
		doc := declaredContractDocument(t, []byte(source))
		e, err := ExportTypedPathModelContext(context.Background(), "changed.gooo", []byte(source), "Choose", doc, name, "")
		if err != nil {
			t.Fatal(err)
		}
		exports = append(exports, e)
	}
	if !reflect.DeepEqual(exports[0].Inputs, exports[1].Inputs) ||
		exports[0].ContractConditions.FeatureSHA == exports[1].ContractConditions.FeatureSHA {
		t.Fatal("Boolean target change was lost or contaminated source input")
	}
	for _, changed := range exports[2:] {
		if reflect.DeepEqual(exports[0].Inputs, changed.Inputs) ||
			!reflect.DeepEqual(exports[0].ContractCases, changed.ContractCases) ||
			!reflect.DeepEqual(exports[0].ContractConditions, changed.ContractConditions) {
			t.Fatal("ordered expression change was lost or contaminated declared goals")
		}
	}
}

func TestInteractionContractBoundedDigestPreservesLegacyWidths(t *testing.T) {
	for _, size := range []int{decision.FeatureDim, decision.ExecutionFeatureDim, decision.ExecutionFlowFeatureDim, contractdecision.OrderedFeatureDim} {
		features := make([]float32, size)
		for i := range features {
			features[i] = math.Float32frombits(uint32(i + 1))
		}
		h := sha256.New()
		if err := binary.Write(h, binary.LittleEndian, features); err != nil {
			t.Fatal(err)
		}
		if candidateFeatureDigest(features) != fmt.Sprintf("sha256:%x", h.Sum(nil)) {
			t.Fatal("FP32 bytes changed", size)
		}
	}
}

func TestInteractionContractCancellationAndContradictoryCases(t *testing.T) {
	checkContractCancellationAndContradictoryCases(t, writeInteractionContractModel(t, contractdecision.ExtremePooling))
}

func checkContractCancellationAndContradictoryCases(t *testing.T, name string) {
	t.Helper()
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	source := conditionModelSource()
	doc := declaredContractDocument(t, source)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := g.Generate(ctx, "cancel.gooo", source, "Choose", doc, TypedPathOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost", err)
	}
	changed := []byte(strings.Replace(string(source), `case "9" -> "9"`, `case "9" -> "10"`, 1))
	r, err := g.Generate(context.Background(), "partial.gooo", changed, "Choose", declaredContractDocument(t, changed), TypedPathOptions{StepAttempts: 1})
	if err != nil || r.Report.BodyPaths.FunctionalCompleteness >= 100 ||
		len(r.Report.BodyPaths.Search.Attempts) != 4 || r.Report.BodyPaths.ContractRanking.Calls != 1 {
		t.Fatal("contradiction accepted or ranking repeated", err)
	}
}
