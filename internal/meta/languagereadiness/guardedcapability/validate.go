package guardedcapability

import (
	"fmt"
	"reflect"
)

func Validate(receipt Receipt) error {
	switch {
	case receipt.Schema != Schema:
		return fmt.Errorf("guarded capability schema %q is unknown", receipt.Schema)
	case len(receipt.Coordinates) != 8 || len(receipt.Indicators) != 8 || len(receipt.Proofs) != 3:
		return fmt.Errorf("guarded capability evidence dimensions are invalid")
	case receipt.ReportDigest == "":
		return fmt.Errorf("guarded capability report digest is missing")
	case !reflect.DeepEqual(receipt, Build(receipt.Source)):
		return fmt.Errorf("guarded capability receipt does not replay")
	default:
		return nil
	}
}

func ValidateForHead(receipt Receipt, expectedHeadSHA string) error {
	if err := Validate(receipt); err != nil {
		return err
	}
	if receipt.Source.CurrentHeadSHA != expectedHeadSHA {
		return fmt.Errorf("guarded capability head does not match exact subject")
	}
	return nil
}
