package metrictransition

import (
	"bytes"
	"fmt"
	"os"
)

// VerifyFiles rebuilds the transition from evidence and compares canonical bytes.
func VerifyFiles(options Options, statePath, ledgerPath string) error {
	result, err := Build(options)
	if err != nil {
		return err
	}
	expectedState, expectedLedger, err := Documents(result)
	if err != nil {
		return err
	}
	actualState, err := os.ReadFile(statePath)
	if err != nil {
		return err
	}
	actualLedger, err := os.ReadFile(ledgerPath)
	if err != nil {
		return err
	}
	if !bytes.Equal(actualState, expectedState) || !bytes.Equal(actualLedger, expectedLedger) {
		return fmt.Errorf("metric transition canonical replay mismatch")
	}
	return nil
}
