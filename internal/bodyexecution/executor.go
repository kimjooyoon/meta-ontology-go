package bodyexecution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

const driverContract = "gooo/stdlib-int64-array-driver/v1"

var ErrExecutorClosed = errors.New("native executor is closed")

// ArtifactObservation refers to a real build owned by this executor. Build in
// the enclosing observation describes only work performed by the current call.
type ArtifactObservation struct {
	Schema             string             `json:"schema"`
	KeySHA256          string             `json:"key_sha256"`
	Reused             bool               `json:"reused"`
	ExecutableVerified bool               `json:"executable_verified"`
	SourceBuild        ProcessObservation `json:"source_build"`
	SourceBuildSHA256  string             `json:"source_build_sha256"`
	MissReason         string             `json:"miss_reason,omitempty"`
	WaitNS             int64              `json:"wait_ns"`
}

type compiledArtifact struct {
	key, root, executable, digest string
	build                         ProcessObservation
}

// Executor owns at most one compiled workspace. Calls are serialized through a
// cancellable gate; each call replays current source and executes current cases.
// Close cancels active work, waits for child completion and removes the workspace.
type Executor struct {
	gate        chan struct{}
	lifetime    context.Context
	stop        context.CancelFunc
	artifact    *compiledArtifact
	toolchain   *ownedToolchain
	hashScratch *[fileHashBufferBytes]byte
}

func NewExecutor() *Executor {
	ctx, stop := context.WithCancel(context.Background())
	e := &Executor{gate: make(chan struct{}, 1), lifetime: ctx, stop: stop}
	e.gate <- struct{}{}
	return e
}

func (e *Executor) Execute(ctx context.Context, filename string, source []byte, document pathplan.Document,
	prior bodycodegen.Result, parent []byte, cases []pathplan.TestCase, goBinary string) (Result, error) {
	start := time.Now()
	gatePhase := bodytiming.Start(ctx, "executor_gate")
	reject := func(err error) (Result, error) {
		gatePhase.End(false)
		r := initialResult(source, prior, parent, cases)
		r.Observation.Failure, r.Observation.ElapsedNS = err.Error(), time.Since(start).Nanoseconds()
		phase := bodytiming.Start(ctx, "runtime_receipt")
		r.CompletenessReceipt = runtimeCompleteness(prior, r)
		phase.End(true)
		return r, err
	}
	if ctx == nil || e == nil || e.lifetime == nil {
		return reject(errors.New("initialized executor and context are required"))
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		return reject(ctx.Err())
	case <-e.lifetime.Done():
		return reject(ErrExecutorClosed)
	case <-e.gate:
	}
	defer func() { e.gate <- struct{}{} }()
	if err := ctx.Err(); err != nil {
		return reject(err)
	}
	if e.lifetime.Err() != nil {
		return reject(ErrExecutorClosed)
	}
	wait := time.Since(start).Nanoseconds()
	gatePhase.End(true)
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	r, err := execute(ctx, filename, source, document, prior, parent, cases, goBinary, e)
	if ctx.Err() != nil {
		err = errors.Join(err, ctx.Err(), e.drop())
	}
	if r.Observation.Artifact != nil {
		r.Observation.Artifact.WaitNS = wait
	}
	r.Observation.ElapsedNS = time.Since(start).Nanoseconds()
	if err != nil {
		r.Observation.Failure = err.Error()
	}
	phase := bodytiming.Start(ctx, "runtime_receipt")
	r.CompletenessReceipt = runtimeCompleteness(prior, r)
	phase.End(true)
	return r, err
}

func (e *Executor) Close() error {
	if e == nil || e.lifetime == nil {
		return nil
	}
	e.stop()
	<-e.gate
	defer func() { e.gate <- struct{}{} }()
	return e.drop()
}

func (e *Executor) drop() error {
	e.toolchain = nil
	e.hashScratch = nil
	return e.dropArtifact()
}

// bindFile uses scratch owned by the executor's existing serialization gate.
// It still reads current bytes on every call; Close and cancellation release it.
func (e *Executor) bindFile(path string) (string, error) {
	if e == nil {
		return fileDigest(path)
	}
	if e.hashScratch == nil {
		e.hashScratch = new([fileHashBufferBytes]byte)
	}
	return fileDigestBuffer(path, e.hashScratch[:])
}

func (e *Executor) dropArtifact() error {
	if e.artifact == nil {
		return nil
	}
	if err := os.RemoveAll(e.artifact.root); err != nil {
		return err
	}
	e.artifact = nil
	return nil
}

func copyProcess(p ProcessObservation) ProcessObservation {
	if p.ExitCode != nil {
		v := *p.ExitCode
		p.ExitCode = &v
	}
	if p.PeakRSSBytes != nil {
		v := *p.PeakRSSBytes
		p.PeakRSSBytes = &v
	}
	return p
}

func artifactKey(prior bodycodegen.Result, r *Observation, goBinary string) string {
	env := childEnvironment()
	sort.Strings(env)
	key, _ := json.Marshal([]any{driverContract, digest([]byte(prior.Source)), prior.Report.Activity,
		r.GoToolSHA256, goBinary, r.GoVersion, runtime.GOOS, runtime.GOARCH, r.ProducerSourceSHA, runtime.Version(), env})
	return digest(key)
}

func executableFor(ctx context.Context, prior bodycodegen.Result, goBinary string,
	r *Observation, owner *Executor) (root, executable string, release func(), err error) {
	release = func() {}
	key := ""
	if owner != nil {
		key = artifactKey(prior, r, goBinary)
		r.Artifact = &ArtifactObservation{Schema: "gooo/owned-native-artifact/v1", KeySHA256: key, MissReason: "empty"}
		if cached := owner.artifact; cached != nil {
			r.Artifact.MissReason = "key_changed"
			if cached.key == key {
				observed, readErr := owner.bindFile(cached.executable)
				if readErr == nil && observed == cached.digest {
					r.ExecutableSHA256 = observed
					r.Artifact.Reused, r.Artifact.ExecutableVerified, r.Artifact.MissReason = true, true, ""
					r.Artifact.SourceBuild = copyProcess(cached.build)
					bindSourceBuild(r)
					return cached.root, cached.executable, release, nil
				}
				r.Artifact.MissReason = "executable_changed_or_missing"
			}
			if err := owner.dropArtifact(); err != nil {
				return "", "", release, err
			}
		}
	}
	root, err = os.MkdirTemp("", "gooo-body-runtime-")
	if err != nil {
		return "", "", release, fmt.Errorf("cannot create runtime workspace")
	}
	release = func() { _ = os.RemoveAll(root) }
	if err := prepareWorkspace(root, prior); err != nil {
		return root, "", release, err
	}
	executable = filepath.Join(root, "observed-body")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	r.Stage = "BUILD"
	_, r.Build, err = process(ctx, root, goBinary, nil, "build", "-trimpath", "-buildvcs=false", "-o", executable, ".")
	if err != nil {
		return root, executable, release, err
	}
	r.ExecutableSHA256, err = owner.bindFile(executable)
	if err != nil {
		return root, executable, release, fmt.Errorf("cannot bind emitted executable")
	}
	if owner != nil {
		owner.artifact = &compiledArtifact{key: key, root: root, executable: executable, digest: r.ExecutableSHA256, build: copyProcess(r.Build)}
		r.Artifact.ExecutableVerified, r.Artifact.SourceBuild = true, copyProcess(r.Build)
		bindSourceBuild(r)
		release = func() {}
	}
	return root, executable, release, nil
}

func bindSourceBuild(r *Observation) {
	b, _ := json.Marshal([]any{r.Artifact.KeySHA256, r.ExecutableSHA256, r.Artifact.SourceBuild})
	r.Artifact.SourceBuildSHA256 = digest(b)
}
