package languageprofile

import "github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"

type Measurer interface {
	Measure(func() sourceexecution.Receipt) (sourceexecution.Receipt, Measurement)
}
