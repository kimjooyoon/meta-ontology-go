package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// ObservationBinding identifies the runner invocation observed by the trace.
type ObservationBinding struct {
	Fixture   string    `json:"fixture"`
	Operation Operation `json:"operation"`
	RunID     string    `json:"run_id"`
}

// ObserverPaths are the paths captured outside the producer response.
type ObserverPaths struct {
	SourcePath string `json:"source_path"`
	OutputPath string `json:"output_path"`
	TempRoot   string `json:"temp_root"`
}

// LstatIdentity retains metadata needed to detect replacement with equal bytes.
type LstatIdentity struct {
	Exists          bool   `json:"exists"`
	Device          string `json:"device,omitempty"`
	Inode           string `json:"inode,omitempty"`
	Mode            string `json:"mode,omitempty"`
	Size            int64  `json:"size"`
	ModTimeUnixNano int64  `json:"mtime_unix_nano"`
}

// FileObservation is an observer-owned path snapshot.
type FileObservation struct {
	Path       string        `json:"path"`
	Kind       string        `json:"kind"`
	Exists     bool          `json:"exists"`
	ByteDigest string        `json:"byte_digest,omitempty"`
	Lstat      LstatIdentity `json:"lstat"`
}

// TempArtifactSnapshot is a canonical recursive snapshot rooted at TempRoot.
type TempArtifactSnapshot struct {
	Root    LstatIdentity     `json:"root"`
	Digest  string            `json:"digest"`
	Entries []FileObservation `json:"entries"`
}

// FilesystemState contains the source, output, and temporary artifact state.
type FilesystemState struct {
	Source FileObservation      `json:"source"`
	Output FileObservation      `json:"output"`
	Temp   TempArtifactSnapshot `json:"temp"`
}

// NoWriteObservation is returned by an independent NoWriteObserver.
// The private stamp prevents a response or decoded wire payload from becoming proof.
type NoWriteObservation struct {
	Binding ObservationBinding `json:"binding"`
	Paths   ObserverPaths      `json:"paths"`
	Before  FilesystemState    `json:"before"`
	After   FilesystemState    `json:"after"`
	stamp   *observerStamp
}

type observerStamp struct {
	digest [sha256.Size]byte
}

type NoWriteObserver struct {
	binding  ObservationBinding
	paths    ObserverPaths
	before   FilesystemState
	stamp    *observerStamp
	finished bool
}

// NewNoWriteObserver captures the pre-invocation state using os.Lstat and bytes.
func NewNoWriteObserver(binding ObservationBinding, paths ObserverPaths) (*NoWriteObserver, error) {
	if err := binding.validate(); err != nil {
		return nil, err
	}
	normalized, err := normalizeObserverPaths(paths)
	if err != nil {
		return nil, err
	}
	before, err := captureState(normalized)
	if err != nil {
		return nil, fmt.Errorf("capture before state: %w", err)
	}
	return &NoWriteObserver{
		binding: binding, paths: normalized, before: before, stamp: &observerStamp{},
	}, nil
}

// Finish captures the post-invocation state. Differences are verified by the oracle.
func (o *NoWriteObserver) Finish() (NoWriteObservation, error) {
	if o == nil || o.stamp == nil || o.finished {
		return NoWriteObservation{}, fmt.Errorf("observer is not initialized")
	}
	o.finished = true
	after, err := captureState(o.paths)
	if err != nil {
		return NoWriteObservation{}, fmt.Errorf("capture after state: %w", err)
	}
	observation := NoWriteObservation{
		Binding: o.binding, Paths: o.paths, Before: o.before, After: after, stamp: o.stamp,
	}
	o.stamp.digest = observationSeal(observation)
	return observation, nil
}

func (b ObservationBinding) validate() error {
	if strings.TrimSpace(b.Fixture) == "" || strings.TrimSpace(string(b.Operation)) == "" || strings.TrimSpace(b.RunID) == "" {
		return fmt.Errorf("observation fixture, operation, and run_id are required")
	}
	if !knownOperation(b.Operation) {
		return fmt.Errorf("unsupported observation operation %q", b.Operation)
	}
	return nil
}

func normalizeObserverPaths(paths ObserverPaths) (ObserverPaths, error) {
	values := []string{paths.SourcePath, paths.OutputPath, paths.TempRoot}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return ObserverPaths{}, fmt.Errorf("observer source, output, and temp paths are required")
		}
	}
	var err error
	paths.SourcePath, err = filepath.Abs(filepath.Clean(paths.SourcePath))
	if err != nil {
		return ObserverPaths{}, fmt.Errorf("source path: %w", err)
	}
	paths.OutputPath, err = filepath.Abs(filepath.Clean(paths.OutputPath))
	if err != nil {
		return ObserverPaths{}, fmt.Errorf("output path: %w", err)
	}
	paths.TempRoot, err = filepath.Abs(filepath.Clean(paths.TempRoot))
	if err != nil {
		return ObserverPaths{}, fmt.Errorf("temp root: %w", err)
	}
	info, err := os.Lstat(paths.TempRoot)
	if err != nil || !info.IsDir() {
		return ObserverPaths{}, fmt.Errorf("temp root must be an existing directory")
	}
	return paths, nil
}

func captureState(paths ObserverPaths) (FilesystemState, error) {
	source, err := captureFile(paths.SourcePath)
	if err != nil {
		return FilesystemState{}, fmt.Errorf("source: %w", err)
	}
	output, err := captureFile(paths.OutputPath)
	if err != nil {
		return FilesystemState{}, fmt.Errorf("output: %w", err)
	}
	temp, err := captureTemp(paths.TempRoot)
	if err != nil {
		return FilesystemState{}, fmt.Errorf("temp: %w", err)
	}
	return FilesystemState{Source: source, Output: output, Temp: temp}, nil
}

func captureFile(path string) (FileObservation, error) {
	observation, err := capturePath(path)
	if err != nil {
		return FileObservation{}, err
	}
	if observation.Exists && observation.Kind != "file" {
		return FileObservation{}, fmt.Errorf("%s is not a regular file", path)
	}
	return observation, nil
}

func capturePath(path string) (FileObservation, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return FileObservation{Path: path, Kind: "missing", Lstat: LstatIdentity{}}, nil
	}
	if err != nil {
		return FileObservation{}, err
	}
	observation := FileObservation{
		Path: path, Kind: fileKind(info), Exists: true, Lstat: makeLstat(info),
	}
	if info.Mode().IsRegular() {
		data, err := os.ReadFile(path)
		if err != nil {
			return FileObservation{}, err
		}
		observation.ByteDigest = digestBytes(data)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return FileObservation{}, err
		}
		observation.ByteDigest = digestBytes([]byte(target))
	}
	if info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		after, err := os.Lstat(path)
		if err != nil || makeLstat(after) != observation.Lstat {
			return FileObservation{}, fmt.Errorf("path changed during capture")
		}
	}
	return observation, nil
}

func fileKind(info os.FileInfo) string {
	switch {
	case info.Mode().IsRegular():
		return "file"
	case info.IsDir():
		return "directory"
	case info.Mode()&os.ModeSymlink != 0:
		return "symlink"
	default:
		return "other"
	}
}

func makeLstat(info os.FileInfo) LstatIdentity {
	return LstatIdentity{
		Exists: true, Device: statNumber(info.Sys(), "Dev"), Inode: statNumber(info.Sys(), "Ino"),
		Mode: info.Mode().String(), Size: info.Size(), ModTimeUnixNano: info.ModTime().UnixNano(),
	}
}

func statNumber(value any, fieldName string) string {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() == reflect.Pointer {
		if reflected.IsNil() {
			return ""
		}
		reflected = reflected.Elem()
	}
	if !reflected.IsValid() || reflected.Kind() != reflect.Struct {
		return ""
	}
	field := reflected.FieldByName(fieldName)
	if !field.IsValid() {
		return ""
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprintf("%d", field.Uint())
	default:
		return ""
	}
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
