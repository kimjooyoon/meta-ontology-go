package changedsurfacereceipt

func CaseInput(id, subjectSHA string) Input {
	pass := func(surface string) Receipt {
		return Receipt{SurfaceID: surface, Decision: "PASS", Resolution: ResolutionExact}
	}
	input := Input{Schema: InputSchema, SubjectSHA: subjectSHA, ChangedSurfaces: []string{"cmd/gooo", "internal/query"}, Receipts: []Receipt{pass("cmd/gooo"), pass("internal/query")}}
	switch id {
	case "zero-change":
		input.ChangedSurfaces, input.Receipts = nil, nil
	case "missing":
		input.Receipts = input.Receipts[:1]
	case "orphan":
		input.ChangedSurfaces = input.ChangedSurfaces[:1]
	case "duplicate":
		input.ChangedSurfaces = input.ChangedSurfaces[:1]
		input.Receipts = []Receipt{pass("cmd/gooo"), pass("cmd/gooo")}
	case "unknown-top":
		input.ChangedSurfaces = input.ChangedSurfaces[:1]
		input.Receipts = []Receipt{{SurfaceID: "cmd/gooo", Decision: "UNKNOWN", Resolution: ResolutionUnknown}}
	}
	return input
}
