package main

import (
	//"log"
	//"net/http"
	//"time"
	"fmt"
	s "strings"

	mf "github.com/mixamarciv/gofncstd3000"
)

var apppath string

func InitApp() {
	apppath, _ = mf.AppPath()
	apppath = s.Replace(apppath, "\\", "/", -1)
	fmt.Printf("apppath: " + apppath)
}
