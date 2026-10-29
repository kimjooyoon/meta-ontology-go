package bodycodegen

import (
	"fmt"
	"io"
)

func writeOriginLoopText(w io.Writer, flow *RecordValueFlow) error {
	for _, loop := range flow.Loops {
		if _, err := fmt.Fprintf(w, "\nLoop #%d: collection=#%d index=%s parent-loop=#%d\n",
			loop.ID, loop.Collection, loop.IndexKind, loop.Parent); err != nil {
			return err
		}
	}
	for _, carry := range flow.LoopCarries {
		if _, err := fmt.Fprintf(w, "  source recurrence #%d: initial=#%d next=#%d exit=#%d loop=#%d\n",
			carry.Header, carry.Seed, carry.Yield, carry.Exit, carry.Loop); err != nil {
			return err
		}
	}
	return nil
}
