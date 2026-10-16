package bodyexecution

import (
	"os"
	"time"
)

// Cleanup runs after the executor has regained its child-completion gate.
// Six backoffs schedule at most 630 ms; filesystem call duration is separate.
func removeRuntimeWorkspace(root string) error {
	return removeWorkspaceWithRetry(root, os.RemoveAll, retryableWorkspaceRemoval, time.Sleep)
}

func removeWorkspaceWithRetry(root string, remove func(string) error,
	retryable func(error) bool, pause func(time.Duration)) error {
	delay := 10 * time.Millisecond
	for attempt := 0; ; attempt++ {
		err := remove(root)
		if err == nil || attempt == 6 || !retryable(err) {
			return err
		}
		pause(delay)
		delay *= 2
	}
}
