package integrationprogress

import "strconv"

func itoa(value int) string { return strconv.Itoa(value) }

func i64toa(value int64) string { return strconv.FormatInt(value, 10) }
