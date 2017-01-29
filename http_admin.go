package main

import (
	//"fmt"
	"net/http"
	"net/url"
)

//хендлер для /admin
func http_admin(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	get_vars, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["get_vars"] = get_vars

	r.ParseForm()
	post_vars := r.FormValue
	d["post_vars"] = post_vars

	RenderTemplate(w, r, d, "maintemplate.html", "page_admin.html")
}
