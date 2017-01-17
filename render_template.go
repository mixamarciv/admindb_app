package main

import (
	"bytes"
	"fmt"

	"html"
	"html/template"
	"io/ioutil"
	"net/http"

	mf "github.com/mixamarciv/gofncstd3000"
	//"reflect"
	"strings"

	"github.com/gorilla/context"
)

func init() {
	//rtr.HandleFunc("/", mf.LogreqF("/", page_main)).Methods("GET")
	fmt.Printf("")
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, d map[string]interface{}, filenames ...string) {
	if len(filenames) == 0 {
		ShowError("RenderTemplate: no filenames", fmt.Errorf("html/template: no files named in call to ParseFiles"), w, r)
		return
	}

	//--------------------------------------------------------------------------
	//объявляем функции которые будут использоваться в шаблонах
	s := GetSess(w, r)
	funcs := template.FuncMap{
		//возвращает значение переменной сессии
		"fsess": func(name, defaultval string) interface{} {
			return GetSessVal(s, name, defaultval)
		},
		//возвращает значение переменной контекста
		"fctx": func(name, defaultval string) interface{} {
			return GetCtx(r, name, defaultval)
		},
		"floadTime": func() interface{} {
			return GetLoadTime(r)
		},
		"dump": func(v interface{}) string {
			return html.EscapeString(fmt.Sprintf("%+v", v))
		},
		"dump_t": func() string {
			return fmt.Sprintf("%#v", r.URL)
			//return html.EscapeString(fmt.Sprintf("%#v", r.URL))
		},
		"unsafeHtml": func(s string) template.HTML {
			return template.HTML(s)
		},
		"unsafeHtmlPost": func(s string) template.HTML {
			i := strings.Index(s, "</code>")
			if i >= 0 {
				//LogPrint("--1-------------------------------------------------------\n" + s + "\n----------------------------------------------------------\n")
				s = mf.StrRegexpReplace(s, "<code>[ \t]*\n", "<code>")
				s = mf.StrRegexpReplace(s, "</code>[ \t]*\n", "</code>")
				//LogPrint("--2-------------------------------------------------------\n" + s + "\n----------------------------------------------------------\n")
			}
			return template.HTML(s)
		},
	}

	//--------------------------------------------------------------------------
	//в d["sess"] сохраняем значения всех переменных сессии
	if _, ok := d["sess"]; !ok {
		//копируем значения всех переменных текущей сессии
		//ещё и преобразуем их из json строки в map[string]interface{}
		allsessvars := make(map[interface{}]interface{})
		for k, v := range s.Values {
			str, ok := v.(string)
			if ok {
				t, err := mf.FromJson([]byte(str))
				if err == nil {
					allsessvars[k] = t
				} else {
					allsessvars[k] = str
				}
			} else {
				allsessvars[k] = v
			}

		}

		//и задаем значения для переменных по умолчанию (если они не заданы)
		for k, v := range gcfg_default_session_data {
			if _, b := allsessvars[k]; !b {
				allsessvars[k] = v
			}
		}
		d["sess"] = allsessvars

	}

	//в d["ctx"] сохраняем значения всех переменных контекста
	if _, ok := d["ctx"]; !ok {
		d["ctx"] = context.GetAll(r)
	}

	//в d["http_request"] передаем данные из http.Request
	if _, ok := d["http_request"]; !ok {
		//d["http_request"] = *r
		s := make(map[string]interface{})
		s["method"] = r.Method
		s["url"] = r.URL
		s["host"] = r.Host
		s["remoteAddr "] = r.RemoteAddr
		s["requestURI"] = r.RequestURI
		s["header"] = r.Header
		s["contentLength"] = r.ContentLength

		d["http_request"] = s
	}
	//--------------------------------------------------------------------------
	//собираем все файлы в один шаблон
	var t *template.Template
	for _, file := range filenames {
		template_file := apppath + "/templates/" + file
		template_text, err := ioutil.ReadFile(template_file)
		if err != nil {
			ShowError("RenderTemplate: read template file error", err, w, r)
			return
		}
		s := string(template_text)

		var tmpl *template.Template
		if t == nil {
			t = template.New(file).Funcs(funcs)
		}

		if file == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(file)
		}

		_, err = tmpl.Parse(s)
		if err != nil {
			ShowError("RenderTemplate: parse template file \""+file+"\" error", err, w, r)
			return
		}

	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if gcfg_debug["render_template"] == 1 {
		buff := new(bytes.Buffer)
		err := t.Execute(buff, d)
		if err != nil {
			ShowError("render template error", err, w, r)
			return
		}
		w.Write(buff.Bytes())
	} else {
		err := t.Execute(w, d)
		if err != nil {
			ShowError("render template error", err, w, r)
			return
		}
	}
}
