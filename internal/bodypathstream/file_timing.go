package bodypathstream

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

const timingSchema = "gooo/body-path-file-timing/v1"
const timingScope = "Sequential current-request wall intervals through original artifact save and binding; " +
	"excludes initial input/model loading, timing sidecar/summary writes, stdout and cleanup. " +
	"Generation contains its existing sub-timings; do not add them twice. " +
	"Absent phases are unobserved. Saved-file hashes and interval consistency do not re-attest execution. " +
	"CPU and RSS remain separate current-child observations; whole-host CPU is unobserved."

type fileTiming struct {
	Schema            string              `json:"schema"`
	Sequence          uint64              `json:"sequence"`
	Status            string              `json:"status"`
	RequestSHA256     string              `json:"request_sha256"`
	ProducerSourceSHA string              `json:"producer_source_sha"`
	ProducerModified  string              `json:"producer_modified"`
	ResponseNS        int64               `json:"response_ns"`
	Wall              bodytiming.Snapshot `json:"wall"`
	Files             map[string]string   `json:"files"`
	Scope             string              `json:"scope"`
}

type fileTimingRow struct {
	Sequence     uint64             `json:"sequence"`
	Status       string             `json:"status"`
	File         string             `json:"file"`
	FileSHA256   string             `json:"file_sha256"`
	ResponseMS   float64            `json:"response_ms"`
	CaptureMS    float64            `json:"capture_ms"`
	UnassignedMS float64            `json:"unassigned_ms"`
	PhaseMS      map[string]float64 `json:"phase_ms"`
}

type fileTimingSummary struct {
	Schema string          `json:"schema"`
	Scope  string          `json:"scope"`
	Rows   []fileTimingRow `json:"rows"`
}

func savedNames(result Result) []string {
	stem := fmt.Sprintf("run-%d", result.Sequence)
	names := []string{stem + "-request.json", stem + "-response.json"}
	if result.Response != nil {
		names = append(names, stem+"-generation.json", stem+"-generated.go")
	}
	if result.Execution != nil {
		names = append(names, stem+"-runtime.json")
	}
	return names
}

func saveFileTiming(ctx context.Context, out string, request []byte, result Result,
	responseNS int64, recorder *bodytiming.Recorder) (fileTimingRow, error) {
	phase := bodytiming.Start(ctx, "artifact_bind")
	t := fileTiming{Schema: timingSchema, Sequence: result.Sequence, Status: result.Status,
		RequestSHA256: timingDigest(request), ResponseNS: responseNS, Scope: timingScope,
		Files: make(map[string]string), ProducerSourceSHA: "unobserved", ProducerModified: "unobserved"}
	if build, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range build.Settings {
			if setting.Key == "vcs.revision" {
				t.ProducerSourceSHA = setting.Value
			}
			if setting.Key == "vcs.modified" {
				t.ProducerModified = setting.Value
			}
		}
	}
	for _, name := range savedNames(result) {
		raw, err := readTimingFile(out, name)
		if err != nil {
			phase.End(false)
			return fileTimingRow{}, err
		}
		t.Files[name] = timingDigest(raw)
	}
	phase.End(true)
	t.Wall = recorder.Snapshot()
	name := fmt.Sprintf("run-%d-timing.json", result.Sequence)
	if err := writeFileJSON(out, name, t); err != nil {
		return fileTimingRow{}, err
	}
	raw, err := readTimingFile(out, name)
	return timingRow(t, name, raw), err
}

func timingRow(t fileTiming, name string, raw []byte) fileTimingRow {
	r := fileTimingRow{Sequence: t.Sequence, Status: t.Status, File: name, FileSHA256: timingDigest(raw),
		ResponseMS: float64(t.ResponseNS) / 1e6, CaptureMS: float64(t.Wall.CaptureNS) / 1e6,
		UnassignedMS: float64(t.Wall.UnassignedNS) / 1e6, PhaseMS: make(map[string]float64)}
	for _, p := range t.Wall.Phases {
		r.PhaseMS[p.Name] += float64(p.EndNS-p.StartNS) / 1e6
	}
	return r
}

func timingDigest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }

// readTimingFile accepts only a bounded regular file directly in this directory.
// It does not follow a symlink to a different artifact.
func readTimingFile(out, name string) ([]byte, error) {
	if filepath.Base(name) != name || name == "." || name == ".." {
		return nil, errors.New("invalid timing filename")
	}
	path := filepath.Join(out, name)
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("timing artifact %s must be a regular file", name)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) || after.Size() > 16<<20 {
		return nil, errors.New("timing artifact changed or exceeds 16 MiB")
	}
	raw, err := io.ReadAll(io.LimitReader(f, (16<<20)+1))
	if err != nil || len(raw) > 16<<20 {
		return nil, errors.New("cannot read bounded timing artifact")
	}
	return raw, nil
}

func decodeTiming(raw []byte, target any) error {
	if err := decision.RejectDuplicateJSONKeys(raw); err != nil {
		return err
	}
	if err := timingKeys(raw, reflect.TypeOf(target).Elem()); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		return errors.New("timing artifact has trailing content")
	}
	return nil
}

// Struct keys must match their wire spelling exactly. Filename and phase maps
// keep their literal keys; encoding/json's case-insensitive aliases are rejected.
func timingKeys(raw []byte, shape reflect.Type) error {
	switch shape.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return err
		}
		allowed := make(map[string]reflect.Type, shape.NumField())
		for field := range shape.Fields() {
			allowed[field.Tag.Get("json")] = field.Type
		}
		for key, value := range fields {
			kind, ok := allowed[key]
			if !ok {
				return fmt.Errorf("unknown or noncanonical timing key: %s", key)
			}
			if err := timingKeys(value, kind); err != nil {
				return err
			}
		}
	case reflect.Slice:
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		for _, value := range values {
			if err := timingKeys(value, shape.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}

func verifyFileTiming(out string, t fileTiming) error {
	if t.Schema != timingSchema || t.Sequence < 1 || t.Sequence > 16 || t.Scope != timingScope ||
		t.Wall.Status != "OBSERVED" || t.ResponseNS <= 0 || t.ResponseNS > t.Wall.CaptureNS ||
		t.Wall.UnassignedNS < 0 || len(t.Wall.Phases) > bodytiming.Capacity {
		return errors.New("invalid or incomplete timing observation")
	}
	var end, total int64
	for _, p := range t.Wall.Phases {
		if !bodytiming.KnownPhase(p.Name) || p.StartNS < end || p.EndNS < p.StartNS ||
			p.EndNS > t.Wall.CaptureNS || (p.Outcome != "completed" && p.Outcome != "failed") {
			return errors.New("invalid timing interval")
		}
		end = p.EndNS
		total += p.EndNS - p.StartNS
	}
	if t.Wall.UnassignedNS != t.Wall.CaptureNS-total {
		return errors.New("timing interval accounting differs")
	}
	response, err := readTimingFile(out, fmt.Sprintf("run-%d-response.json", t.Sequence))
	if err != nil {
		return err
	}
	var result Result
	if err := json.Unmarshal(response, &result); err != nil || result.Schema != ResultSchema ||
		result.Sequence != t.Sequence || result.Status != t.Status {
		return errors.New("timing and response identity differ")
	}
	names := savedNames(result)
	if len(t.Files) != len(names) {
		return errors.New("timing saved-file inventory differs")
	}
	for _, name := range names {
		raw, err := readTimingFile(out, name)
		if err != nil {
			return err
		}
		if t.Files[name] != timingDigest(raw) {
			return fmt.Errorf("timing saved-file binding differs: %s", name)
		}
		if name == fmt.Sprintf("run-%d-request.json", t.Sequence) && t.RequestSHA256 != timingDigest(raw) {
			return errors.New("timing request binding differs")
		}
	}
	return nil
}

func verifyTimingDirectory(ctx context.Context, out string) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := readTimingFile(out, "timing-summary.json")
	if err != nil {
		return err
	}
	var s fileTimingSummary
	if err := decodeTiming(raw, &s); err != nil {
		return err
	}
	if s.Schema != "gooo/body-path-file-timing-summary/v1" || s.Scope != timingScope ||
		len(s.Rows) == 0 || len(s.Rows) > 16 {
		return errors.New("invalid timing summary")
	}
	oldRaw, err := readTimingFile(out, "summary.json")
	if err != nil {
		return err
	}
	var old struct {
		Schema string       `json:"schema"`
		Rows   []fileRunRow `json:"rows"`
	}
	if err := json.Unmarshal(oldRaw, &old); err != nil || old.Schema != "gooo/body-path-file-run/v1" ||
		len(old.Rows) != len(s.Rows) {
		return errors.New("timing and original summary inventory differ")
	}
	for i, row := range s.Rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := fmt.Sprintf("run-%d-timing.json", i+1)
		raw, err := readTimingFile(out, name)
		if err != nil {
			return err
		}
		var t fileTiming
		if err := decodeTiming(raw, &t); err != nil {
			return err
		}
		if t.Sequence != uint64(i+1) || old.Rows[i].Sequence != t.Sequence || old.Rows[i].Status != t.Status ||
			old.Rows[i].ResponseMS != row.ResponseMS || !reflect.DeepEqual(row, timingRow(t, name, raw)) {
			return errors.New("timing summary and original sidecar differ")
		}
		if err := verifyFileTiming(out, t); err != nil {
			return err
		}
	}
	return nil
}
