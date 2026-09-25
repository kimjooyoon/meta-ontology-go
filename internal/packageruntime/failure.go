package packageruntime

import "fmt"

type Failure struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func (failure *Failure) Error() string {
	return failure.Code + ": " + failure.Detail
}

func reject(code, format string, values ...any) error {
	return &Failure{Code: code, Detail: fmt.Sprintf(format, values...)}
}
