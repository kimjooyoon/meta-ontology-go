package languagesourceexecution

type Input struct {
	Contract      Contract
	HeadSHA       string
	Positive      []byte
	Replay        []byte
	UnknownEntry  []byte
	InvalidSyntax []byte
}
