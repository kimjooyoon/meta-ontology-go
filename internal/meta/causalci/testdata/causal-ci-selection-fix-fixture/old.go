package fixfixture

import "sync"

// Run is intentionally written in the pre-Go 1.25 form so the Go 1.27
// waitgroupgo fixer must report an active diff in the isolated fixture.
func Run(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
	}()
}
