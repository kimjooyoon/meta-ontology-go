package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
)

func (r *JointConstruction) runCandidates(ctx context.Context, filename string, source []byte,
	cases CompositionCases, options JointOptions, slots []jointSlot, calls *jointCallContext) error {
	executor := NewExecutor()
	defer executor.Close()
	for _, masks := range jointMaskOrder(slots, options.ProgramBudget) {
		if err := ctx.Err(); err != nil {
			return err
		}
		attempt, selected, program, err := materializeJoint(ctx, filename, source, cases, options.EntryActivity, slots, masks)
		if err != nil {
			return err
		}
		if attempt.Rejection == nil {
			attempt.Runtime, err = executor.executeWithJointCalls(ctx, filename, selected, program, cases, options.GoBinary, calls)
		} else if r.Schema != jointRejectionSchema && r.Schema != jointFillSchema && r.Schema != jointFillRejectionSchema {
			r.Scope += "; rejected local expressions consume attempts without a caller score; rejected local totals cover only the scored prefix"
		}
		r.Attempts = append(r.Attempts, attempt)
		if r.Schema != jointFaultSchema && hasCompositionFault(attempt.Runtime) {
			r.Scope += "; native integer zero-divisor outcomes preserve reached operation identities and operands; dependent activities are blocked; twice-reproduced language faults consume program attempts and permit continuation"
		}
		r.Schema = jointObservationSchema(slots, r.Attempts)
		if err != nil {
			return err
		}
		if attempt.Rejection == nil && (r.SelectedAttempt < 0 || betterJoint(attempt, r.Attempts[r.SelectedAttempt])) {
			r.SelectedAttempt, r.SelectedSource, r.Selected = len(r.Attempts)-1, string(selected), program
		}
		if raw, err := json.MarshalIndent(r, "", "  "); err != nil || len(raw) > 30<<20 {
			return fmt.Errorf("joint construction exceeds the 32 MiB observation limit")
		}
		if jointComplete(attempt) {
			break
		}
	}
	return nil
}
