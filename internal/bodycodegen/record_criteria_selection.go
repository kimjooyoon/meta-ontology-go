package bodycodegen

func usableRecordAttempt(a RecordAssemblyAttempt) bool {
	if a.Status != "" {
		return false
	}
	return a.Total > 0 || (a.Criteria != nil && a.Criteria.InputsCompleted > 0)
}

func completeRecordAttempt(a RecordAssemblyAttempt) bool {
	return usableRecordAttempt(a) && a.Passed == a.Total &&
		(a.Criteria == nil || a.Criteria.Passed == a.Criteria.Total)
}

func selectedRecordAttempt(r *RecordAssemblyReceipt) RecordAssemblyAttempt {
	return RecordAssemblyAttempt{Mask: r.SelectedMask, Passed: r.Passed, Total: r.Total,
		FieldsPassed: r.FieldsPassed, FieldsTotal: r.FieldsTotal, Criteria: r.Criteria}
}

func selectRecordAttempt(r *RecordAssemblyReceipt, a RecordAssemblyAttempt, cases []RecordAssemblyCase) {
	r.SelectedMask, r.Cases, r.Criteria = a.Mask, cases, a.Criteria
	r.Passed, r.Total, r.FieldsPassed, r.FieldsTotal = a.Passed, a.Total, a.FieldsPassed, a.FieldsTotal
}

// Each policy row asks whether one exact record matches or all predicates on
// one criteria input hold. Predicate and field denominators remain separate.
func recordPolicyScore(a RecordAssemblyAttempt) (int, int) {
	passed, total := a.Passed, a.Total
	if a.Criteria != nil {
		passed += a.Criteria.InputsPassed
		total += a.Criteria.InputsTotal
	}
	return passed, total
}

// Criteria selections prefer complete exact-value cases, then satisfied
// predicate/input pairs, then matching fields. Ties retain the earlier rank.
// Existing programs without criteria keep their historical ordering.
func betterRecordAttempt(a, b RecordAssemblyAttempt, policy bool) bool {
	if policy {
		x, _ := recordPolicyScore(a)
		y, _ := recordPolicyScore(b)
		if x != y {
			return x > y
		}
	}
	if a.Criteria != nil && b.Criteria != nil {
		if a.Passed != b.Passed {
			return a.Passed > b.Passed
		}
		if a.Criteria.Passed != b.Criteria.Passed {
			return a.Criteria.Passed > b.Criteria.Passed
		}
	}
	if a.FieldsPassed != b.FieldsPassed {
		return a.FieldsPassed > b.FieldsPassed
	}
	return a.Passed > b.Passed
}
