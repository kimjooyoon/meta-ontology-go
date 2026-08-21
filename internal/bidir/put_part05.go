package bidir

import (
	"fmt"
)

// CheckPutGet verifies that an accepted model is visible after write-back.
func CheckPutGet(document Document, model Model) error {
	written, err := Put(document, model)
	if err != nil {
		return err
	}
	observed, err := Get(written)
	if err != nil {
		return err
	}
	if !SemanticEquivalent(model, observed) {
		return fmt.Errorf("Put-Get violated: semantic model changed after write-back")
	}
	return nil
}
