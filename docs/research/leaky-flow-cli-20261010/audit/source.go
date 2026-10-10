package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type sdkJudgment struct {
	ID, SourceSHA, InputSHA string
	Prediction              flowdecision.Prediction
	Selected                struct {
		Mask       uint16
		Outputs    []pathplan.TestResult
		Conditions []pathplan.ConditionResult
		Acceptable bool
	}
}

func originalJudgments(root string) map[string]sdkJudgment {
	raw, err := os.ReadFile(filepath.Join(root, "sdk-first-judgments.jsonl.gz"))
	must(err == nil, "original SDK records")
	must(digest(raw) == "sha256:4ad76a785a122b478128b88edce4a056b90e9f867c7e7d631d5a3cf05de61acf", "frozen SDK judgment bytes")
	rows := map[string]sdkJudgment{}
	for line := range strings.SplitSeq(string(read(root, "sdk-first-judgments.jsonl.gz")), "\n") {
		if line != "" {
			row := decode[sdkJudgment]([]byte(line))
			must(rows[row.ID].ID == "", "unique original source")
			rows[row.ID] = row
		}
	}
	must(len(rows) == 162, "all original SDK observations")
	return rows
}

func checkOriginal(root string, pre bodycodegen.TypedPathContextExport, p *bodycodegen.BodyPathReceipt, row sdkJudgment) {
	source := string(read(root, "source.gooo.gz"))
	suffix := "\nactivity Main(Integer) -> Integer computes \"return Choose(input)\"\n"
	must(strings.HasSuffix(source, suffix), "only known caller appended")
	must(digest([]byte(strings.TrimSuffix(source, suffix))) == "sha256:"+row.SourceSHA, "original SDK source retained")
	h := sha256.New()
	_, _ = h.Write([]byte{byte(len(pre.Inputs))})
	for _, input := range pre.Inputs {
		var raw [1536]byte
		for i, x := range input.FlowFeatures {
			binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(x))
		}
		_, _ = h.Write(raw[:])
	}
	must(fmt.Sprintf("%x", h.Sum(nil)) == row.InputSHA, "exact SDK first input")
	first := p.Search.Attempts[0]
	must(p.ConditionProgress[0].Ranking.Proposed == row.Prediction.Selected && first.Mask == row.Selected.Mask, "original prediction reaches immediate candidate")
	must(reflect.DeepEqual(first.Results, row.Selected.Outputs) && reflect.DeepEqual(first.Conditions, row.Selected.Conditions), "original first candidate outcomes")
}

func declaredOutputs(root string) []pathplan.TestCase {
	r := regexp.MustCompile(`(?m)^ case "(-?[0-9]+)" -> "(-?[0-9]+)"$`)
	rows := r.FindAllStringSubmatch(string(read(root, "source.gooo.gz")), -1)
	must(len(rows) == 8, "eight source-authored outputs")
	var result []pathplan.TestCase
	for _, row := range rows {
		input, e1 := strconv.ParseInt(row[1], 10, 64)
		want, e2 := strconv.ParseInt(row[2], 10, 64)
		must(e1 == nil && e2 == nil, "exact source integers")
		result = append(result, pathplan.TestCase{Input: input, Expected: want})
	}
	return result
}

func checkSourceOutputs(root string, p *bodycodegen.BodyPathReceipt) {
	cases := declaredOutputs(root)
	must(len(p.NativeCases) == len(cases), "all declared output results")
	for i, r := range p.NativeCases {
		must(r.Input == cases[i].Input && r.Expected == cases[i].Expected && r.Actual == r.Expected && r.Passed, "exact source output")
	}
	for _, attempt := range p.Search.Attempts {
		must(len(attempt.Results) == len(cases), "every attempted output retained")
		passed := 0
		for i, r := range attempt.Results {
			must(r.Input == cases[i].Input && r.Expected == cases[i].Expected && r.Passed == (r.Actual == r.Expected), "exact attempted source output")
			if r.Passed {
				passed++
			}
		}
		must(attempt.Total == len(cases) && attempt.Passed == passed, "attempt completeness count")
	}
}
