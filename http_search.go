package main

import (
	//"log"
	"net/http"
	"net/url"
)

//хендлер для /s
func http_search(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	q, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["url_vars"] = q

	dtype, ok := q["d"]
	if !ok {
		d["err"] = fmt.Errorf("%s", "ОШИБКА: не верно указана БД")
		return RenderTemplate(w, r, d, "maintemplate.html", "search.html")
	}

	db := dbmap[dtype]
	d["db"] = db

	if db.NeedAuth {
		redirect, err := check_user_access_to_db(w, r, d)
		if err != nil {
			d["err_text"] = "ОШИБКА: у вас нет доступа к БД:\"" + d["db"].(DBd).Name + "\""
			d["err"] = err
			return RenderTemplate(w, r, d, "maintemplate.html", "search.html")
		}
		if redirect {
			return
		}
	}
	
	http_search__load_data(w, r, d)
	
	RenderTemplate(w, r, d, "maintemplate.html", "search.html")
}

//загрузка данных в d["data"] по заданным критериям из d["url_vars"]
func http_search__load_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	return
}

//проверяет авторизован ли пользователь и имеет ли достаточно прав для доступа к бд d["db"]
//если авторизован и есть право доступа то возвращает (0,nil)
//если аторизован но недостаточно прав то возвращаем ошибку (0,err)
//если не авторизован то редиректит на http.Redirect(w, r, "/login?f="+d["url_rawquery"], 301) и возвращает (1,nil)
func check_user_access_to_db(w http.ResponseWriter, r *http.Request, d map[string]interface{}) bool,error {
	return (0,nil)
}
