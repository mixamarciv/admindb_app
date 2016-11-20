package main

import (
	//"log"
	"net/http"
	"net/url"
)

func http_search(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	q, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_vars"] = q

	if _, ok := q["a"]; ok {
		d["db"] = dbmap["a"]
	} else if _, ok := q["p"]; ok {
		d["db"] = dbmap["p"]
	} else if _, ok := q["v"]; ok {
		d["db"] = dbmap["v"]
		err := user_check_access(w, r, d)
		if err != nil {
			d["err_text"] = "ОШИБКА: у вас нет доступа к БД:\"" + d["db"].(DBd).Name + "\""
			d["err"] = err
		}
	}

	RenderTemplate(w, r, d, "maintemplate.html", "search.html")
}
