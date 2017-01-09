package main

import (
	"fmt"
	"net/http"
	"net/url"
)

//хендлер для /s
func http_login(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	get_vars, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["get_vars"] = get_vars

	r.ParseForm()
	post_vars := map[string]interface{}{}
	post_vars["login"] = r.FormValue("login")
	post_vars["pass"] = r.FormValue("pass")
	d["post_vars"] = post_vars

	_, ok := get_vars["d"]
	if !ok {
		d["err"] = fmt.Errorf("%s", "ОШИБКА db001: не верно указана БД")
		RenderTemplate(w, r, d, "maintemplate.html", "login.html")
		return
	}

	RenderTemplate(w, r, d, "maintemplate.html", "login.html")
}
