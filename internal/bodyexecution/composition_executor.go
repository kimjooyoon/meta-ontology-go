package bodyexecution

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"sort"
	"time"
)

// ExecuteComposition retains one compiled graph in this executor. Each call
// replays its current source and executes its current case suite twice. The
// scalar and graph methods share the bounded workspace and cancellation gate.
func (e *Executor) ExecuteComposition(ctx context.Context, filename string, source []byte,
	prior Composition, suite CompositionCases, goBinary string) (CompositionRuntime, error) {
	started := time.Now()
	reject := func(err error) (CompositionRuntime, error) {
		r := initialCompositionRuntime(source, prior, suite)
		r.Schema, r.Stage = "gooo/body-composition-runtime/v2", "EXECUTOR_GATE"
		r.Failure, r.ElapsedNS = err.Error(), time.Since(started).Nanoseconds()
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
	wait := time.Since(started).Nanoseconds()
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	r, err := executeComposition(ctx, filename, source, prior, suite, goBinary, e)
	if ctx.Err() != nil {
		err = errors.Join(err, ctx.Err(), e.drop())
	}
	if r.Artifact != nil {
		r.Artifact.WaitNS = wait
	}
	r.ElapsedNS = time.Since(started).Nanoseconds()
	if err != nil {
		r.Failure = err.Error()
	}
	return r, err
}

func compositionArtifactKey(prior Composition, r *CompositionRuntime, goBinary string) string {
	env := childEnvironment()
	sort.Strings(env)
	raw, _ := json.Marshal([]any{"gooo/value-graph-driver/v1", prior.GeneratedSHA256, prior.DriverSHA256,
		r.GoToolSHA256, goBinary, r.GoVersion, runtime.GOOS, runtime.GOARCH, r.ProducerSourceSHA, runtime.Version(), env})
	return digest(raw)
}

func compositionExecutableFor(ctx context.Context, prior Composition, goBinary string,
	r *CompositionRuntime, owner *Executor) (root, executable string, release func(), err error) {
	release = func() {}
	key := ""
	if owner != nil {
		key = compositionArtifactKey(prior, r, goBinary)
		r.Artifact = &ArtifactObservation{Schema: "gooo/owned-native-artifact/v1", KeySHA256: key, MissReason: "empty"}
		if cached := owner.artifact; cached != nil {
			r.Artifact.MissReason = "key_changed"
			if cached.key == key {
				observed, readErr := owner.bindFile(cached.executable)
				if readErr == nil && observed == cached.digest {
					r.ExecutableSHA256 = observed
					r.Artifact.Reused, r.Artifact.ExecutableVerified, r.Artifact.MissReason = true, true, ""
					r.Artifact.SourceBuild = copyProcess(cached.build)
					bindCompositionBuild(r)
					return cached.root, cached.executable, release, nil
				}
				r.Artifact.MissReason = "executable_changed_or_missing"
			}
			if err := owner.dropArtifact(); err != nil {
				return "", "", release, err
			}
		}
	}
	root, err = os.MkdirTemp("", "gooo-composition-runtime-")
	if err != nil {
		return "", "", release, err
	}
	release = func() { _ = os.RemoveAll(root) }
	executable, err = buildCompositionExecutable(ctx, root, prior.Source, prior.Driver, goBinary, r)
	if err != nil {
		return root, executable, release, err
	}
	if owner != nil {
		owner.artifact = &compiledArtifact{key: key, root: root, executable: executable,
			digest: r.ExecutableSHA256, build: copyProcess(r.Build)}
		r.Artifact.ExecutableVerified, r.Artifact.SourceBuild = true, copyProcess(r.Build)
		bindCompositionBuild(r)
		release = func() {}
	}
	return root, executable, release, nil
}

func bindCompositionBuild(r *CompositionRuntime) {
	raw, _ := json.Marshal([]any{r.Artifact.KeySHA256, r.ExecutableSHA256, r.Artifact.SourceBuild})
	r.Artifact.SourceBuildSHA256 = digest(raw)
}
