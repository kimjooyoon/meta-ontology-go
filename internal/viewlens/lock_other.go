//go:build !darwin && !linux && !windows

package viewlens

import (
	"fmt"
	"os"
)

func tryLock(*os.File) (bool, error) {
	return false, fmt.Errorf("PROFILE_LOCK_PLATFORM_UNAVAILABLE")
}
