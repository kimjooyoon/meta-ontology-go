package toolchainrelease

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type nativeSmokeSite struct {
	ExpressionID string `json:"expression_id"`
	Root         string `json:"root_activity_id"`
	Activity     string `json:"activity_id"`
	Operator     string `json:"operator"`
	Expression   string `json:"expression"`
	Projection   string `json:"projection_sha256"`
	Start        int    `json:"start"`
	End          int    `json:"end"`
}

type nativeSmokeDelivery struct {
	ID                      string `json:"activity_id"`
	Producer                string `json:"producer_id"`
	Input, Actual, Expected json.RawMessage
	Passed                  *bool
	Fault                   *struct {
		Kind        string
		Site        nativeSmokeSite
		Left, Right *int64
	}
	Blocked []string `json:"blocked_by"`
	Inputs  []struct {
		Producer string `json:"producer_id"`
		Value    json.RawMessage
	}
}

type nativeSmokeRuntime struct {
	Schema, Stage, Failure string
	SHA                    string `json:"generated_sha256"`
	Source                 string `json:"original_source_sha256"`
	ObservedProjection     string `json:"observed_projection_sha256"`
	ObservedDriver         string `json:"observed_driver_sha256"`
	Passed                 *int   `json:"finite_passed"`
	Total                  *int   `json:"finite_total"`
	Calls                  *int   `json:"model_calls"`
	Projection             *bool  `json:"projection_replayed"`
	Replay                 *bool  `json:"runtime_replayed"`
	Outcomes               *struct{ Matched, Mismatched, Faulted, Blocked, Unobserved *int }
	Sites                  []nativeSmokeSite `json:"fault_sites"`
	Runs                   []json.RawMessage
	Traces                 []struct {
		Index      int `json:"case_index"`
		Deliveries []nativeSmokeDelivery
	}
}

func validateNativeSmokeRuntime(r nativeSmokeRuntime, counts [5]int, rows int) error {
	if r.Schema != "gooo/body-composition-runtime/v3" || r.Stage != "COMPLETE" || r.Failure != "" ||
		!nativeSmokeDigest(r.SHA) || !nativeSmokeDigest(r.ObservedProjection) || !nativeSmokeDigest(r.ObservedDriver) ||
		!jointSmokeBool(r.Projection, true) || !jointSmokeBool(r.Replay, true) || !jointSmokeInt(r.Calls, 0) ||
		!jointSmokeInt(r.Passed, counts[0]) || !jointSmokeInt(r.Total, counts[0]+counts[1]+counts[2]+counts[3]+counts[4]) ||
		r.Outcomes == nil || len(r.Sites) != 1 || len(r.Traces) != rows {
		return fmt.Errorf("native arithmetic runtime identity or denominator differs")
	}
	stored := [5]*int{r.Outcomes.Matched, r.Outcomes.Mismatched, r.Outcomes.Faulted, r.Outcomes.Blocked, r.Outcomes.Unobserved}
	for i, n := range counts {
		if !jointSmokeInt(stored[i], n) {
			return fmt.Errorf("native arithmetic outcome categories differ")
		}
	}
	for i, row := range r.Traces {
		if row.Index != i {
			return fmt.Errorf("native arithmetic case order differs")
		}
	}
	return validateNativeSmokeRuns(r.Runs)
}

func validateNativeSmokeFault(r nativeSmokeRuntime, d nativeSmokeDelivery, activity string, left int64) error {
	f := d.Fault
	if f == nil || f.Kind != "ZERO_DIVISOR" || f.Left == nil || *f.Left != left || f.Right == nil || *f.Right != 0 ||
		!nativeSmokeNull(d.Actual) || len(d.Blocked) > 0 || !reflect.DeepEqual(f.Site, r.Sites[0]) ||
		f.Site.Root != d.ID || f.Site.Activity != activity || f.Site.Operator != "/" || f.Site.Projection != r.SHA ||
		f.Site.Start < 0 || f.Site.End <= f.Site.Start || f.Site.End-f.Site.Start != len(f.Site.Expression) {
		return fmt.Errorf("native arithmetic fault operands or source binding differ")
	}
	site := f.Site
	id := site.ExpressionID
	site.ExpressionID = ""
	raw, err := json.Marshal(site)
	if err != nil {
		return err
	}
	if id != fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) {
		return fmt.Errorf("native arithmetic expression identity differs")
	}
	return nil
}

func validateNativeSmokeRuns(rows []json.RawMessage) error {
	if len(rows) != 2 {
		return fmt.Errorf("native arithmetic requires two observed executions")
	}
	first := ""
	for _, raw := range rows {
		var r struct {
			Started, Completed, Canceled bool
			TimedOut                     bool   `json:"timed_out"`
			Truncated                    bool   `json:"output_truncated"`
			Diagnostics                  int    `json:"diagnostics_bytes"`
			Exit                         *int   `json:"exit_code"`
			Stdout                       string `json:"stdout_sha256"`
			Stderr                       string `json:"stderr_sha256"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		if !r.Started || !r.Completed || r.Canceled || r.TimedOut || r.Truncated || r.Diagnostics != 0 || !jointSmokeInt(r.Exit, 0) ||
			!nativeSmokeDigest(r.Stdout) || r.Stderr != fmt.Sprintf("sha256:%x", sha256.Sum256(nil)) || first != "" && first != r.Stdout {
			return fmt.Errorf("native arithmetic process incomplete or replay differs")
		}
		first = r.Stdout
	}
	return nil
}

func nativeSmokeNull(raw json.RawMessage) bool {
	return len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
func nativeSmokeDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
