package workspaceexecution

import (
	"context"
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

// Executor owns one native artifact across workspace requests. Each request
// still reconstructs current sources and executes its current suite twice.
// Only preparation of the exact verified executable may be reused.
type Executor struct {
	native   *bodyexecution.Executor
	gate     chan struct{}
	lifetime context.Context
	stop     context.CancelFunc
}

func NewExecutor() *Executor {
	ctx, stop := context.WithCancel(context.Background())
	e := &Executor{native: bodyexecution.NewExecutor(), gate: make(chan struct{}, 1), lifetime: ctx, stop: stop}
	e.gate <- struct{}{}
	return e
}

func (e *Executor) Execute(ctx context.Context, manifest packageruntime.Manifest,
	suite bodyexecution.CompositionCases, options ExecuteOptions) (Result, error) {
	return e.request(ctx, func(current context.Context) (Result, error) {
		options.nativeExecutor = e.native
		return ExecuteWorkspaceWithOptions(current, manifest, suite, options)
	})
}

func (e *Executor) Replay(ctx context.Context, manifest packageruntime.Manifest, prior Result,
	suite bodyexecution.CompositionCases, goBinary string) (Result, error) {
	return e.request(ctx, func(current context.Context) (Result, error) {
		return replayWorkspace(current, manifest, prior, suite, goBinary, e.native)
	})
}

func (e *Executor) request(ctx context.Context, run func(context.Context) (Result, error)) (Result, error) {
	if ctx == nil || e == nil || e.lifetime == nil || e.native == nil || run == nil {
		return Result{}, errors.New("initialized workspace executor and context are required")
	}
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-e.lifetime.Done():
		return Result{}, bodyexecution.ErrExecutorClosed
	case <-e.gate:
	}
	defer func() { e.gate <- struct{}{} }()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if e.lifetime.Err() != nil {
		return Result{}, bodyexecution.ErrExecutorClosed
	}
	current, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	result, err := run(current)
	if current.Err() != nil && !errors.Is(err, current.Err()) {
		return result, errors.Join(err, current.Err())
	}
	return result, err
}

// Close cancels active work and waits for it before releasing the native owner.
// A blocked caller cannot enter a closed executor after the gate is released.
func (e *Executor) Close() error {
	if e == nil || e.lifetime == nil || e.native == nil {
		return nil
	}
	e.stop()
	<-e.gate
	defer func() { e.gate <- struct{}{} }()
	return e.native.Close()
}

func executeWorkspaceNative(ctx context.Context, source []byte, prior bodyexecution.Composition,
	suite bodyexecution.CompositionCases, goBinary string, owner *bodyexecution.Executor) (bodyexecution.CompositionRuntime, error) {
	if owner != nil {
		return owner.ExecuteComposition(ctx, "workspace.gooo", source, prior, suite, goBinary)
	}
	return bodyexecution.ExecuteComposition(ctx, "workspace.gooo", source, prior, suite, goBinary)
}
