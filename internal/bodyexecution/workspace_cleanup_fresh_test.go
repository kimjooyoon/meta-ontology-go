package bodyexecution

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWorkspaceRemovalBudgetFresh1711(t *testing.T) {
	sharing, permanent := errors.New("sharing"), errors.New("permanent")
	for _, tc := range []struct {
		name               string
		failures, attempts int
		failure            error
		wantDelay          time.Duration
		wantError          error
	}{
		{"first-success", 0, 1, sharing, 0, nil},
		{"one-lock", 1, 2, sharing, 10 * time.Millisecond, nil},
		{"three-locks", 3, 4, sharing, 70 * time.Millisecond, nil},
		{"permanent-error", 1, 1, permanent, 0, permanent},
		{"persistent-lock", 100, 7, sharing, 630 * time.Millisecond, sharing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, delayed := 0, time.Duration(0)
			remove := func(root string) error {
				if root != "owned-workspace" {
					t.Fatal("cleanup changed ownership path")
				}
				calls++
				if calls <= tc.failures {
					return tc.failure
				}
				return nil
			}
			err := removeWorkspaceWithRetry("owned-workspace", remove,
				func(err error) bool { return errors.Is(err, sharing) }, func(d time.Duration) { delayed += d })
			if err != tc.wantError || calls != tc.attempts || delayed != tc.wantDelay {
				t.Fatalf("cleanup error/attempts/scheduled delay: %v, %d, %s", err, calls, delayed)
			}
		})
	}
}

func TestWorkspaceCloseOwnershipFresh1717(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "owned.bin"), []byte("owned"), 0600); err != nil {
		t.Fatal(err)
	}
	e := NewExecutor()
	e.artifact = &compiledArtifact{root: root}
	<-e.gate
	var once sync.Once
	unlock := func() { once.Do(func() { e.gate <- struct{}{} }) }
	defer unlock()
	done := make(chan error, 1)
	go func() { done <- e.Close() }()
	select {
	case <-e.lifetime.Done():
	case <-time.After(time.Second):
		t.Fatal("close did not cancel its lifetime")
	}
	select {
	case <-done:
		t.Fatal("close removed a workspace before its caller released the gate")
	default:
	}
	unlock()
	select {
	case err := <-done:
		if err != nil || e.artifact != nil {
			t.Fatalf("close did not release owned artifact: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("close did not finish after gate release")
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("owned workspace remains: %v", err)
	}
}
