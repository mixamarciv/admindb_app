package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	//mf "github.com/mixamarciv/gofncstd3000"
)

//выводит сообщение об ошибке на странице
func RenderError(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	RenderTemplate(w, r, d, "maintemplate.html", "error_info_page.html")
	return
}

//первоначальный парсинг данных урл
func http_parse_url(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	d := make(map[string]interface{})
	q, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["get_vars"] = q
	return d
}

//получаем данные бд и права доступа пользователя к этой бд
func http_parse_url__get_db(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	q := d["get_vars"].(url.Values)
	dtype, ok := q["d"]
	if !ok {
		d["error"] = fmt.Errorf("%s", "ОШИБКА http_parse_url__get_db001: не верно указана/не указана БД")
		d["errorcode"] = "dbnotfound"
		return
	}

	db, ok := dbmap[dtype[0]]
	if !ok {
		d["error"] = fmt.Errorf("%s", "ОШИБКА http_parse_url__get_db002: не верно указана БД (бд \""+dtype[0]+"\" не существует)")
		d["errorcode"] = "dbnotfound"
		return
	}
	d["db"] = db

	if db.NeedAuth {
		u := GetSessUserData(w, r)
		d["user"] = u
		if _, b := u["error"]; b {
			d["error"] = u["error"]
			d["errorcode"] = "noauth"
			return
		}

		fdata := u["fdata"].(map[string]interface{})
		accessdb := fdata["accessdb"].(map[string]interface{})

		shortnamedb := db.ShortName

		access, ok := accessdb[shortnamedb]
		if !ok {
			d["error"] = u["error"]
			d["errorcode"] = "noauth"
			return
		}

		d["db_access"] = access.(string)
		if access.(string) == "0" {
			d["error"] = fmt.Errorf("%s", "у вас нет доступа к БД \""+db.Name+"\"")
			d["errorcode"] = "dbnoaccess"
			return
		}
	} else {
		d["db_access"] = "1" //доступ по умолчанию к бд к которой не нужна авторизация
	}

	return
}

//возвращает int значение переменной varname из url.Values или defaultval если значение не задано
func http_get_var_int(get_vars url.Values, varname string, defaultval int) int {
	vals, ok := get_vars[varname]
	if !ok {
		return defaultval
	}
	val, err := strconv.Atoi(vals[0])
	if err != nil {
		return defaultval
	}
	return val
}
func http_get_var_str(get_vars url.Values, varname string, defaultval int) int {
	vals, ok := get_vars[varname]
	if !ok {
		return defaultval
	}
	return vals[0]
}
