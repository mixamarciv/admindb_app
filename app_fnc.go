package main

import (
	"fmt"
	//"log"
	"net/http"
	//"net/url"
	"database/sql"
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
		LogPrint(mf.CurTimeStrRFC3339() + " <- " + r.URL.Scheme + " " + r.URL.Path + "?" + r.URL.RawQuery)

		if r.URL.RawQuery == "" { //проверяем что бы обязательно были заданы хотябы 1 параметр после ?
			http.Redirect(w, r, "/?main", 301)
			LogPrint("## redirect to /?main ")
			return
		}
		//context.Set(r, "nextfunc", nextfunc)

		f(w, r)

		LogPrint(fmt.Sprintf(mf.CurTimeStrRFC3339()+" -> "+r.URL.Scheme+" "+r.URL.Path+"?"+r.URL.RawQuery+"  %v ", GetLoadTime(r)))
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
	LogPrint(serr)

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
	LogPrint(serr)

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
	LogPrint(serr)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(serr))

}

//возвращает значение элемента d[args1][args2][args3]
//или возвращает defaultval если этот элемент не существует
func get_map_val(d map[string]interface{}, defaultval interface{}, args ...string) interface{} {
	var t interface{}
	t = d
	for _, v := range args {
		x := t.(map[string]interface{})
		a, ok := x[v]
		if !ok {
			return defaultval
		}
		t = a
	}
	return t
}

//запускает запрос query для d["db"].(*DBd).DB
//возвращает rows или nil в случае ошибки
func run_db_query(d map[string]interface{}, query string) *sql.Rows {
	db := d["db"].(*DBd).DB
	//tr_query := string(dbStrTr2([]byte(mf.StrTrim(query)), "UTF-8", "cp1251"))
	tr_query := query
	rows, err := db.Query(tr_query)
	if err != nil {
		d["error"] = fmtError("db.Query(query) ERROR: query:\n"+query+"\n\n", err)
		d["errorcode"] = "dbqueryerror"

		LogPrint("==ERROR:===================================================================")
		LogPrint(fmt.Sprintf("error: %#v", d["error"]))
		LogPrint(fmt.Sprintf("%#v", err))
		LogPrint("query: \n" + query)
		LogPrint("tr_query: \n" + tr_query)
		LogPrint("==ERROR/===================================================================")
		return nil
	}
	return rows
}

//запускает запрос query для d["db"].(*DBd).DB
//возвращает true или nil в случае ошибки
func run_db_exec(d map[string]interface{}, query string) interface{} {
	db := d["db"].(*DBd).DB
	//tr_query := string(dbStrTr2([]byte(mf.StrTrim(query)), "UTF-8", "cp1251"))
	tr_query := query
	_, err := db.Exec(tr_query)
	if err != nil {
		d["error"] = fmtError("db.Exec(query) ERROR: query:\n"+query+"\n\n", err)
		d["errorcode"] = "dbqueryerror"

		LogPrint("==ERROR:===================================================================")
		LogPrint(fmt.Sprintf("error: %#v", d["error"]))
		LogPrint(fmt.Sprintf("%#v", err))
		LogPrint("query: \n" + query)
		//LogPrint("tr_query: \n" + tr_query)
		LogPrint("==ERROR/===================================================================")
		return nil
	}
	return true
}
