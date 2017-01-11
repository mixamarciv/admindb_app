package main

import (
	"fmt"
	"net/http"
	"net/url"
)

//хендлер для /s
func http_search(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}
	LogPrint("http_search start")

	q, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["get_vars"] = q

	dtype, ok := q["d"]
	if !ok {
		d["error"] = fmt.Errorf("%s", "ОШИБКА http_search001: не верно указана БД")
		RenderTemplate(w, r, d, "maintemplate.html", "search.html")
		return
	}

	db, ok := dbmap[dtype[0]]
	if !ok {
		d["error"] = fmt.Errorf("%s", "ОШИБКА http_search002: не верно указана БД (возможно эта бд не существует)")
		RenderTemplate(w, r, d, "maintemplate.html", "search.html")
		return
	}
	d["db"] = db

	if db.NeedAuth {
		isauth, hasaccess := check_user_access_to_db(w, r, d)
		LogPrint(fmt.Sprintf("isauth, hasaccess: %v, %v", isauth, hasaccess))
		if !isauth {
			//отправляем на авторизацию

			var Url *url.URL
			Url, err := url.Parse("/login")
			if err != nil {
				d["error"] = fmt.Errorf("%s", "ОШИБКА http_search003: url.Parse")
				RenderTemplate(w, r, d, "maintemplate.html", "search.html")
				return
			}

			parameters := url.Values{}
			parameters.Add("f", "/"+r.URL.RawQuery)
			parameters.Add("msg", "для доступа к этой БД требуется авторизация")
			Url.RawQuery = parameters.Encode()

			LogPrint(fmt.Sprintf("URL: %s", Url.String()))

			http.Redirect(w, r, Url.String(), 301)
			return
		}
		if !hasaccess {
			d["error"] = "ОШИБКА: у вас пока нет доступа к БД \"" + d["db"].(*DBd).Name + "\", доступ выдается админом после прохождения теста"
			RenderTemplate(w, r, d, "maintemplate.html", "search.html")
			return
		}

	} else {
		d["db_access"] = "1" //доступ по умолчанию к бд
	}

	http_search__load_data(w, r, d)

	RenderTemplate(w, r, d, "maintemplate.html", "search.html")
}

//загрузка данных в d["data"] по заданным критериям из d["url_vars"]
func http_search__load_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	return
}

//проверяет авторизован ли пользователь и имеет ли достаточно прав для доступа к бд d["db"]
//если авторизован и есть право доступа то возвращает (1,1)
//если аторизован но недостаточно прав то возвращаем ошибку (1,0)
//если не авторизован то (0,0) //потом возможно редирект на http.Redirect(w, r, "/login?f="+d["url_rawquery"], 301)
func check_user_access_to_db(w http.ResponseWriter, r *http.Request, d map[string]interface{}) (isauth bool, hasaccess bool) {
	u := GetSessUserData(w, r)
	if _, b := u["error"]; b {
		isauth = false
		hasaccess = false
		return isauth, hasaccess
	}

	fdata := u["fdata"].(map[string]interface{})
	accessdb := fdata["accessdb"].(map[string]interface{})

	db := d["db"].(*DBd)
	shortnamedb := db.ShortName

	access, ok := accessdb[shortnamedb]
	if !ok {
		isauth = false
		hasaccess = false
		return isauth, hasaccess
	}

	d["db_access"] = access.(string)

	if access.(string) == "0" {
		isauth = true
		hasaccess = false
		return isauth, hasaccess
	}

	return isauth, hasaccess
}
