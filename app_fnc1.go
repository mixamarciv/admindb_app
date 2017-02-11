package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

var sprintf = fmt.Sprintf

func i64toa(d int64) string {
	return sprintf("%d", d)
}

func itoa(d int) string {
	return sprintf("%d", d)
}

func floatToStr(f interface{}) string {
	return strconv.FormatFloat(f.(float64), 'f', 0, 64)
}

func fmtError(s string, err error) string {
	return s + fmt.Sprintf("\n\n%#v", err)
}

func base64Decode(s string) string {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return fmt.Sprint("error:", err)
	}
	return string(decoded)
}
