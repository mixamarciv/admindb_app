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

	"github.com/davecgh/go-spew/spew"
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
	sess := GetSess(w, r)
	//объявляем функции которые будут использоваться в шаблонах
	var funcs template.FuncMap
	{
		m := make(map[string]string) //глобальная переменная которая будет доступна во всех шаблонах через mget mset
		funcs = template.FuncMap{
			"mget": func(k string) string {
				return m[k]
			},
			"mset": func(k, v string) string {
				m[k] = v
				return ""
			},
			//возвращает значение переменной сессии
			"fsess": func(name, defaultval string) interface{} {
				return GetSessVal(sess, name, defaultval)
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
			"dump_spew": func(v interface{}) string {
				return spew.Sdump(v)
			},
			"dump_spew2": func(v interface{}, depth int, indent string) string {
				cs := &spew.ConfigState{
					Indent:                  indent,
					MaxDepth:                depth,
					SortKeys:                true,
					DisableMethods:          true,
					DisableCapacities:       true,
					DisablePointerAddresses: true,
					DisablePointerMethods:   true,
					SpewKeys:                true,
				}
				//return cs.Sprintf("%#v", v)
				return cs.Sdump(v)
			},
			"unsafeHtml": func(s string) template.HTML {
				return template.HTML(s)
			},
			"unsafeHtmlPostPreview": func(s string) template.HTML {
				i := strings.Index(s, "<code>")
				if i >= 0 {
					s = s[:i]
				}
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
			"toString": func(v interface{}) string {
				return sprintf("%v", v)
			},
		}
	}

	//--------------------------------------------------------------------------
	//в d["sess"] сохраняем значения всех переменных сессии
	if _, ok := d["sess"]; !ok {
		//копируем значения всех переменных текущей сессии
		//ещё и преобразуем их из json строки в map[string]interface{}
		allsessvars := make(map[interface{}]interface{})
		for k, v := range sess.Values {
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

		//и задаем значения для переменных сессии по умолчанию (если они не заданы)
		{
			need_save_sess := false
			for k, v := range gcfg_default_session_data {
				if _, b := allsessvars[k]; !b { //если в текущих переменных сессии значение не задано
					need_save_sess = true //то сохраняем все текущие значения переменных в текущей сессии

					var val interface{}

					f, fb := v.(func() string)
					if fb { //если это функция то вызываем её для получения значения переменной текущей сессии
						val = f()
					} else {
						val = v
					}

					sess.Values[k] = val
					allsessvars[k] = val
				}
			}
			if need_save_sess {
				LogPrint("save sess (new vars)")
				sess.Save(r, w)
			}
			d["sess"] = allsessvars
		}
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
	for ifile, file := range filenames {
		template_file := apppath + "/templates/" + file
		template_text, err := ioutil.ReadFile(template_file)
		if err != nil {
			//ShowError("RenderTemplate: read template file error", err, w, r)
			//return
			template_text = []byte(file)
			file = itoa(ifile)
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
