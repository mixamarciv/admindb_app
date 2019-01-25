package main

import (
	//"fmt"
	//"io/ioutil"
	"net/http"
	"net/url"

	//"strconv"

	//"strings"

	//mf "github.com/mixamarciv/gofncstd3000"

	//"github.com/gorilla/sessions"
)

//авторизация в вк апи
func http_auth_oauth(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	get_vars, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["get_vars"] = get_vars

	RenderTemplate(w, r, d, "maintemplate.html", "login.html")
	return
}
