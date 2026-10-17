package bodyexecution

// FieldParentColumns extends the field-only projection within the primitive
// record limit. The existing field/local and field-only input shapes stay intact.
// ReceiptSHA256 identifies saved file bytes; it does not authenticate a producer.
type FieldParentColumns struct {
	FieldLineageColumns
	ReceiptSHA256 string   `json:"receipt_sha256"`
	Frames        []int    `json:"frames"`
	TypeIDs       []string `json:"type_ids"`
}

// SelectRecordedFieldParent copies retained facts. The receiving Gooo program
// decides whether the selected write has an eligible direct parent.
func SelectRecordedFieldParent(runtime CompositionRuntime, query NativeOriginQuery, receiptSHA256 string) (FieldParentColumns, *NativeOriginInput, error) {
	base, report, err := SelectRecordedFieldLineage(runtime, query)
	if err != nil {
		return FieldParentColumns{}, report, err
	}
	columns := FieldParentColumns{FieldLineageColumns: base, ReceiptSHA256: receiptSHA256,
		Frames: make([]int, base.WriteCount), TypeIDs: make([]string, base.WriteCount)}
	for i, write := range runtime.Traces[base.CaseIndex].FieldWrites {
		site := runtime.FieldSites[write.SiteIndex]
		binding := runtime.FieldBindings[site.Binding]
		columns.Frames[i] = write.Frame
		columns.TypeIDs[i] = binding.Record.Fields[site.Field].TypeID
	}
	report.InputShape = "gooo/field-parent-columns/v1"
	report.Scope = "one selected field use and unsliced recorded field case; exact saved-file SHA, original field journal, write frames and types; receiving Gooo owns direct-parent selection and relation construction; no producer execution or authentication"
	return columns, report, nil
}
