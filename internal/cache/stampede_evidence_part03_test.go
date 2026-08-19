package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func runStampedeHelper(t *testing.T) {
	t.Helper()
	cache, err := Open(os.Getenv(stampedeRootEnv))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv(stampedeReadyEnv), []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := waitForFiles([]string{os.Getenv(stampedeReleaseEnv)}); err != nil {
		t.Fatal(err)
	}
	data, _, hit, err := cache.GetOrCompute(context.Background(), makeTestKey(t, "v1", "billing"), func() ([]byte, error) {
		if err := os.WriteFile(os.Getenv(stampedeComputeEnv), []byte("compute"), 0o600); err != nil {
			return nil, err
		}
		time.Sleep(25 * time.Millisecond)
		return []byte(stampedeExpectedProjection), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(struct {
		Hit  bool   `json:"hit"`
		Data string `json:"data"`
	}{Hit: hit, Data: string(data)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv(stampedeResultEnv), raw, 0o600); err != nil {
		t.Fatal(err)
	}
}
func waitForFiles(paths []string) error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ready := true
		for _, path := range paths {
			if _, err := os.Stat(path); err != nil {
				ready = false
				break
			}
		}
		if ready {
			return nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %v", paths)
}
