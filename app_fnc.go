package main

import (
	"fmt"
	"log"
	"net/http"
	//"net/url"
	"time"

	"github.com/gorilla/context"
	mf "github.com/mixamarciv/gofncstd3000"
)

//------------------------------------------------------------------------------
//функции для работы с контекстом
//возвращает время выполнения запроса
func GetLoadTime(r *http.Request) time.Duration {
	startLoadTime := context.Get(r, "startLoadTime").(time.Time)
	return time.Now().Sub(startLoadTime)
}

func GetCtx(r *http.Request, varname string, defaultval interface{}) interface{} {
	val, ok := context.GetOk(r, varname)
	if !ok {
		return defaultval
	}
	return val
}

func SetCtx(r *http.Request, varname string, val interface{}) {
	context.Set(r, varname, val)
}

//------------------------------------------------------------------------------
//функция для лога всех запросов
func LogReq(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, "startLoadTime", time.Now())
		log.Print("<- " + r.URL.Scheme + " " + r.URL.Path)

		if r.URL.RawQuery == "" { //проверяем что бы обязательно были заданы хотябы 1 параметр после ?
			http.Redirect(w, r, "/?main", 301)
			log.Print("## redirect to /?main ")
			return
		}
		//context.Set(r, "nextfunc", nextfunc)

		f(w, r)

		log.Printf("-> "+r.URL.Scheme+" "+r.URL.Path+"  %v ", GetLoadTime(r))
		context.Clear(r)
	}
}

//------------------------------------------------------------------------------
//функция для преобразования "func(http.ResponseWriter, *http.Request))" в "http.Handler"
func MakeHttpHandler(f func(http.ResponseWriter, *http.Request)) http.Handler {
	a := new(struct_MakeHttpHandler)
	a.Fnc = f
	return a
}

type struct_MakeHttpHandler struct {
	Fnc func(http.ResponseWriter, *http.Request)
}

func (p *struct_MakeHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Fnc(w, r)
}

//------------------------------------------------------------------------------
//функция для вывода ошибок клиенту в читаемом виде
func ShowError(title string, err error, w http.ResponseWriter, r *http.Request) {
	if err == nil {
		return
	}
	serr := "\n\n== ERROR: ======================================\n"
	serr += title + "\n"
	serr += mf.ErrStr(err)
	serr += "\n\n== /ERROR ======================================\n"
	log.Println(serr)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(serr))

}

func ShowErrorJSON(title string, err error, w http.ResponseWriter, r *http.Request) {
	if err == nil {
		return
	}
	serr := "\n\n== ERROR: ======================================\n"
	serr += title + "\n"
	serr += mf.ErrStr(err)
	serr += "\n\n== /ERROR ======================================\n"
	log.Println(serr)

	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	serr = mf.StrRegexpReplace(serr, "\"", "\\\"")
	w.Write([]byte("{\"err\":\"" + serr + "\"}"))

}

func ShowErrors(title string, err []error, w http.ResponseWriter, r *http.Request) {
	if err == nil {
		return
	}
	serr := "\n\n== ERRORs: =====================================\n"
	//err := stacktrace.Propagate(err[0], title)
	serr += title + "\n"
	serr += fmt.Sprintf("%+v", err)
	serr += "\n\n== /ERRORs =====================================\n"
	log.Println(serr)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(serr))

}
