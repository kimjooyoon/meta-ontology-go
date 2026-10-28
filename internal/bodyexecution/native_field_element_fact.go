package bodyexecution

// Reuse the original read fact; the field ID continues to name the selected
// member while element.field_id names the input collection's declared field.
func fillFieldElementFact(fact *CompositionFieldFact, r CompositionRuntime, t CompositionTrace) {
	if t.Bridge == nil || t.Bridge.ElementReads == nil || fact.AssignmentIndex < 0 {
		return
	}
	read := t.Bridge.ElementReads[fact.AssignmentIndex]
	if read == -2 {
		fact.ElementReason = "READ_LIMIT"
	} else if read >= 0 {
		fact.ElementReadIndex = &read
		element := compositionReadFact(r, t.CaseIndex, t.Reads[read])
		fact.Element = &element
	}
}
