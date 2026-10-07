package bodyexecution

import (
	"fmt"
	"reflect"
)

// DecodeExecutionReceipt shares the execution envelope's exact JSON rules with
// callers that wrap composition records. It retains numbers and rejects duplicate
// fields, aliases and trailing documents, within the existing 32 MiB limit.
func DecodeExecutionReceipt(data []byte, target any) error {
	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return fmt.Errorf("execution receipt target must be a non-nil pointer")
	}
	return decode(data, target, 32<<20)
}
