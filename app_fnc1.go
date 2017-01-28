package main

import (
	"fmt"
)

var sprintf = fmt.Sprintf

func i64toa(d int64) string {
	return sprintf("%d", d)
}

func itoa(d int) string {
	return sprintf("%d", d)
}
