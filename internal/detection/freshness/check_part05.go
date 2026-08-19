package freshness

import (
	"os"
)

func stateForRead(item Item, err error) Item {
	if item.State == StateInvalid {
		return item
	}
	if os.IsNotExist(err) {
		item.State = StateMissing
		item.Detail = "path is missing"
		return item
	}
	if os.IsPermission(err) {
		return invalid(item, "path is not readable")
	}
	return invalid(item, "cannot read path")
}
