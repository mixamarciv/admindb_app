package main

import (
	//"fmt"

	"bufio"
	//"io"
	"os"
	//s "strings"

	mf "github.com/mixamarciv/gofncstd3000"

	"strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

func InitMinify() {
	path := apppath + "/public"
	mf.MkdirAll(path)

	m := minify.New()
	m.AddFunc("css", css.Minify)

	//minify css files:
	files := []string{
		path + "/css/bootstrap.dark.min.css",
		path + "/css/blueimp-gallery.min.css",
		path + "/css/bootstrap-image-gallery.min.css",
		path + "/css/stickyfooter.css",
		path + "/css/fonts.css",
		path + "/css/styles.css",
	}
	run_minify(files, path+"/css/minify.dark.css", "css", m)

	files[0] = path + "/css/bootstrap.white.min.css"
	run_minify(files, path+"/css/minify.white.css", "css", m)

	//minify js files:
	m.AddFunc("js", js.Minify)
	files = []string{
		path + "/js/jquery.min.js",
		path + "/js/bootstrap.min.js",
		path + "/js/jquery.blueimp-gallery.min.js",
		path + "/js/bootstrap-image-gallery.min.js",
		path + "/js/scripts.js",
	}
	run_minify(files, path+"/js/minify.all.js", "js", m)
}

func run_minify(files []string, outfile string, ftype string, m *minify.M) {
	LogPrint("minify " + strconv.Itoa(len(files)) + " " + ftype + " files to " + outfile)
	filename := outfile
	fo, err := os.Create(filename)
	LogPrintErrAndExit("ERROR  os.Open(\""+filename+"\"):\n", err)
	defer func() {
		err := fo.Close()
		LogPrintErrAndExit("ERROR fo.Close:\n", err)
	}()
	w := bufio.NewWriter(fo)

	for _, file := range files {
		fi, err := os.Open(file)
		LogPrintErrAndExit("ERROR  os.Open(\""+file+"\"):\n", err)
		defer func() {
			err := fi.Close()
			LogPrintErrAndExit("ERROR fi.Close:\n", err)
		}()
		r := bufio.NewReader(fi)
		err = m.Minify(ftype, w, r)
		LogPrintErrAndExit("ERROR  m.Minify("+ftype+"):\nfile: "+file+"\n", err)
		_, err = w.Write([]byte("\n"))
		LogPrintErrAndExit("ERROR  w.Write:\n", err)
	}
	err = w.Flush()
	LogPrintErrAndExit("ERROR  w.Flush:\n", err)
}
