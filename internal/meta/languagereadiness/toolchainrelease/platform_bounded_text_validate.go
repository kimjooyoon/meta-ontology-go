package toolchainrelease

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type boundedTextQuery struct {
	Text   string `json:"text"`
	Needle int    `json:"needle"`
}

type boundedTextCases struct {
	Schema string
	Cases  []struct {
		Inputs   map[string]boundedTextQuery
		Expected map[string]int64
	}
}

func boundedTextReference(raw []byte) ([]int64, error) {
	var suite boundedTextCases
	if err := json.Unmarshal(raw, &suite); err != nil {
		return nil, err
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != 10 {
		return nil, fmt.Errorf("bounded text release cases differ")
	}
	values := make([]int64, len(suite.Cases))
	for i, row := range suite.Cases {
		query, hasInput := row.Inputs["app/search:Main"]
		expected, hasExpected := row.Expected["app/search:Main"]
		want := int64(-1)
		if query.Needle >= 0 && query.Needle <= 255 {
			want = int64(strings.IndexByte(query.Text, byte(query.Needle)))
		}
		if !hasInput || !hasExpected || len(row.Inputs) != 1 || len(row.Expected) != 1 ||
			len(query.Text) > 1024 || expected != want {
			return nil, fmt.Errorf("bounded text row %d differs from Go byte search", i)
		}
		values[i] = want
	}
	return values, nil
}

type boundedTextSmokeReceipt struct {
	Schema, Decision, Error string
	ManifestDigest          string `json:"manifest_digest"`
	CasesDigest             string `json:"cases_digest"`
	From                    string `json:"replayed_from_sha256"`
	Result                  struct {
		Program struct {
			Entry struct {
				PackagePath string `json:"package_path"`
				Activity    string
			}
		}
		Composition struct {
			SHA string `json:"generated_sha256"`
		}
		Replay *struct {
			Calls *int `json:"model_calls"`
		}
		Runtime struct {
			Calls      *int `json:"model_calls"`
			Passed     *int `json:"finite_passed"`
			Total      *int `json:"finite_total"`
			Projection bool `json:"projection_replayed"`
			Replay     bool `json:"runtime_replayed"`
			Traces     []struct {
				Index      int `json:"case_index"`
				Deliveries []struct {
					ID               string `json:"activity_id"`
					Actual, Expected *int64
					Passed           *bool
				}
			}
		}
	}
}

type boundedTextSmokeBinding struct {
	Program, Receipt string
}

func validateBoundedTextSmoke(raw, manifest, cases []byte, replay bool, selected boundedTextSmokeBinding) (boundedTextSmokeBinding, error) {
	values, err := boundedTextReference(cases)
	if err != nil {
		return boundedTextSmokeBinding{}, err
	}
	var receipt boundedTextSmokeReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return boundedTextSmokeBinding{}, err
	}
	x, sha := receipt.Result.Runtime, receipt.Result.Composition.SHA
	zero := func(v *int) bool { return v != nil && *v == 0 }
	if receipt.Schema != "gooo/workspace-body-execution-receipt/v1" || receipt.Decision != "PASS" || receipt.Error != "" ||
		receipt.ManifestDigest != fmt.Sprintf("sha256:%x", sha256.Sum256(manifest)) ||
		receipt.CasesDigest != fmt.Sprintf("sha256:%x", sha256.Sum256(cases)) ||
		receipt.Result.Program.Entry.PackagePath != "app/search" || receipt.Result.Program.Entry.Activity != "Main" ||
		x.Passed == nil || *x.Passed != 10 || x.Total == nil || *x.Total != 10 || !zero(x.Calls) ||
		!x.Projection || !x.Replay || !boundedTextDigest(sha) || len(x.Traces) != 10 || selected.Program != "" && selected.Program != sha {
		return boundedTextSmokeBinding{}, fmt.Errorf("bounded text source, cases or native results differ")
	}
	if replay && (selected.Program == "" || !boundedTextDigest(selected.Receipt) || receipt.From != selected.Receipt || receipt.Result.Replay == nil || !zero(receipt.Result.Replay.Calls)) {
		return boundedTextSmokeBinding{}, fmt.Errorf("bounded text replay is unbound or predicts again")
	}
	seen := [10]bool{}
	for _, trace := range x.Traces {
		if trace.Index < 0 || trace.Index >= len(seen) || seen[trace.Index] || len(trace.Deliveries) != 1 {
			return boundedTextSmokeBinding{}, fmt.Errorf("bounded text case identity differs")
		}
		seen[trace.Index] = true
		d := trace.Deliveries[0]
		if d.ID != "gooo://bounded-search/app/main" || d.Actual == nil || d.Expected == nil || d.Passed == nil ||
			!(*d.Passed) || *d.Actual != values[trace.Index] || *d.Expected != values[trace.Index] {
			return boundedTextSmokeBinding{}, fmt.Errorf("bounded text byte offset differs at row %d", trace.Index)
		}
	}
	return boundedTextSmokeBinding{Program: sha, Receipt: fmt.Sprintf("sha256:%x", sha256.Sum256(raw))}, nil
}

func boundedTextDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil && value == strings.ToLower(value)
}
