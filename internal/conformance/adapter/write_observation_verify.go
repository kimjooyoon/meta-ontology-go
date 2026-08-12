package adapter

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const (
	OracleNW001   = "ORACLE-NW-001"
	OracleNW002   = "ORACLE-NW-002"
	OracleNW003   = "ORACLE-NW-003"
	OracleNW004   = "ORACLE-NW-004"
	OracleNW005   = "ORACLE-NW-005"
	OracleNW006   = "ORACLE-NW-006"
	OracleFAIL001 = "ORACLE-FAIL-001"
	OracleFAIL002 = "ORACLE-FAIL-002"
	OraclePASS001 = "ORACLE-PASS-001"
	OracleID001   = "ORACLE-ID-001"
)

// NoWriteEvidenceError is a stable observer-oracle rejection.
type NoWriteEvidenceError struct {
	Code   string
	Detail string
}

func (e NoWriteEvidenceError) Error() string { return e.Code + ": " + e.Detail }

func oracleError(code, detail string) NoWriteEvidenceError {
	return NoWriteEvidenceError{Code: code, Detail: detail}
}

// VerifyNoWrite accepts only a trusted observer capture bound to this request.
func (o *NoWriteObservation) VerifyNoWrite(request Request) error {
	if o == nil {
		return oracleError(OracleNW001, "observer evidence is required")
	}
	if o.stamp == nil {
		return oracleError(OracleNW003, "observer-owned capture marker is missing")
	}
	if o.Binding != requestObservationBinding(request) {
		return oracleError(OracleID001, "observer binding does not match request")
	}
	if o.stamp.digest != observationSeal(*o) {
		return oracleError(OracleNW003, "observer seal does not match captured evidence")
	}
	if err := validateObservation(*o); err != nil {
		return err
	}
	if err := comparePrimary(o.Before.Source, o.After.Source, "source"); err != nil {
		return err
	}
	if err := comparePrimary(o.Before.Output, o.After.Output, "output"); err != nil {
		return err
	}
	if o.Before.Temp.Root != o.After.Temp.Root || o.Before.Temp.Digest != o.After.Temp.Digest {
		return oracleError(OracleNW006, "temporary artifact snapshot changed")
	}
	current, err := captureState(o.Paths)
	if err != nil || !reflect.DeepEqual(current, o.After) {
		return oracleError(OracleNW002, "observer trace is stale")
	}
	return nil
}

func requestObservationBinding(request Request) ObservationBinding {
	return ObservationBinding{Fixture: request.Fixture, Operation: request.Operation, RunID: request.RunID}
}

func validateObservation(observation NoWriteObservation) error {
	if err := observation.Binding.validate(); err != nil {
		return oracleError(OracleNW002, "observation binding is stale or malformed")
	}
	if err := validatePaths(observation); err != nil {
		return oracleError(OracleNW003, err.Error())
	}
	if err := validateState(observation.Before, true); err != nil {
		return oracleError(OracleNW003, "before snapshot: "+err.Error())
	}
	if err := validateState(observation.After, true); err != nil {
		return oracleError(OracleNW003, "after snapshot: "+err.Error())
	}
	return nil
}

func validatePaths(observation NoWriteObservation) error {
	paths := observation.Paths
	if !filepath.IsAbs(paths.SourcePath) || !filepath.IsAbs(paths.OutputPath) || !filepath.IsAbs(paths.TempRoot) {
		return fmt.Errorf("observer paths must be absolute")
	}
	if observation.Before.Source.Path != paths.SourcePath || observation.After.Source.Path != paths.SourcePath {
		return fmt.Errorf("source path is not bound to observer paths")
	}
	if observation.Before.Output.Path != paths.OutputPath || observation.After.Output.Path != paths.OutputPath {
		return fmt.Errorf("output path is not bound to observer paths")
	}
	return nil
}

func validateState(state FilesystemState, requirePrimary bool) error {
	if err := validateFileObservation(state.Source, requirePrimary); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if err := validateFileObservation(state.Output, requirePrimary); err != nil {
		return fmt.Errorf("output: %w", err)
	}
	if err := validateTempSnapshot(state.Temp); err != nil {
		return fmt.Errorf("temp: %w", err)
	}
	return nil
}

func validateFileObservation(observation FileObservation, requireFile bool) error {
	if strings.TrimSpace(observation.Path) == "" {
		return fmt.Errorf("path is required")
	}
	if !observation.Exists {
		if observation.Kind != "missing" || observation.ByteDigest != "" || observation.Lstat.Exists {
			return fmt.Errorf("missing path has contradictory state")
		}
		return nil
	}
	if !validKind(observation.Kind) || observation.Kind == "missing" {
		return fmt.Errorf("invalid path kind %q", observation.Kind)
	}
	if !observation.Lstat.Exists || observation.Lstat.Mode == "" {
		return fmt.Errorf("lstat identity is incomplete")
	}
	if requireFile && observation.Kind != "file" {
		return fmt.Errorf("primary path is not a regular file")
	}
	if observation.Kind == "file" || observation.Kind == "symlink" {
		if !validDigest(observation.ByteDigest) {
			return fmt.Errorf("byte digest is incomplete")
		}
	} else if observation.ByteDigest != "" {
		return fmt.Errorf("directory or special path has bytes")
	}
	return nil
}

func validateTempSnapshot(snapshot TempArtifactSnapshot) error {
	if snapshot.Root.Exists && snapshot.Root.Mode == "" {
		return fmt.Errorf("temp root lstat identity is incomplete")
	}
	if !snapshot.Root.Exists && snapshot.Root != (LstatIdentity{}) {
		return fmt.Errorf("missing temp root has contradictory identity")
	}
	if !validDigest(snapshot.Digest) {
		return fmt.Errorf("snapshot digest is incomplete")
	}
	entries := append([]FileObservation{}, snapshot.Entries...)
	if !sort.SliceIsSorted(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path }) {
		return fmt.Errorf("snapshot entries are not canonical")
	}
	for index, entry := range entries {
		if filepath.IsAbs(entry.Path) || entry.Path == "." || entry.Path == ".." || strings.HasPrefix(entry.Path, "../") {
			return fmt.Errorf("snapshot entry path escapes temp root")
		}
		if index > 0 && entries[index-1].Path == entry.Path {
			return fmt.Errorf("snapshot entry path is duplicated")
		}
		if err := validateFileObservation(entry, false); err != nil {
			return err
		}
	}
	computed, err := digestTempSnapshot(snapshot.Root, entries)
	if err != nil || computed != snapshot.Digest {
		return fmt.Errorf("snapshot digest does not match entries")
	}
	return nil
}

func validKind(kind string) bool {
	switch kind {
	case "file", "directory", "symlink", "other":
		return true
	default:
		return false
	}
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func comparePrimary(before, after FileObservation, label string) error {
	if before.Exists != after.Exists || before.ByteDigest != after.ByteDigest {
		return oracleError(OracleNW004, label+" bytes or existence changed")
	}
	if before.Lstat != after.Lstat {
		return oracleError(OracleNW005, label+" lstat identity changed")
	}
	return nil
}
