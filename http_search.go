package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	//mf "github.com/mixamarciv/gofncstd3000"
)

//хендлер для /s
func http_search(w http.ResponseWriter, r *http.Request) {
	d := http_parse_url(w, r)

	http_parse_url__get_db(w, r, d)
	if d["error"] != nil {
		RenderError(w, r, d)
		return
	}

	if d["db_access"].(string) < "1" {
		d["error"] = fmt.Errorf("%s", "у вас нет доступа к БД \""+d["db"].(*DBd).Name+"\"")
		d["errorcode"] = "dbnoaccess"
		RenderError(w, r, d)
		return
	}

	http_search__load_data(w, r, d)

	RenderTemplate(w, r, d, "maintemplate.html", "search.html", "search_data.html")
}

//хендлер для /s
func http_search_old(w http.ResponseWriter, r *http.Request) {
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

			u, err := url.Parse("/login")
			if err != nil {
				d["error"] = fmt.Errorf("%s", "ОШИБКА http_search003: url.Parse")
				RenderTemplate(w, r, d, "maintemplate.html", "search.html")
				return
			}

			parameters := url.Values{}
			parameters.Add("f", "/"+r.URL.RawQuery)
			parameters.Add("msg", "для доступа к этой БД требуется авторизация")
			u.RawQuery = parameters.Encode()

			LogPrint(fmt.Sprintf("URL: %s", u.String()))

			http.Redirect(w, r, u.String(), 301)
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

	RenderTemplate(w, r, d, "maintemplate.html", "search.html", "search_data.html")
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

//загрузка данных в d["data"] по заданным критериям из d["url_vars"]
func http_search__load_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	data := make(map[string]interface{})
	get_vars := d["get_vars"].(url.Values)

	page := http_get_var_int(get_vars, "p", 1)
	data["page"] = page
	data["has_next_page"] = 0

	{ //задаем url_next_page и url_prev_page
		turlstr := r.URL.Path + "?" + r.URL.RawQuery
		u, err := url.Parse(turlstr)
		if err != nil {
			d["error"] = fmt.Errorf("%s", "http_search__load_data ERROR003: url.Parse("+turlstr+")")
			return
		}
		q := u.Query()

		q.Set("p", strconv.Itoa(page+1))
		u.RawQuery = q.Encode()
		data["url_next_page"] = u.String()

		q.Set("p", strconv.Itoa(page-1))
		u.RawQuery = q.Encode()
		data["url_prev_page"] = u.String()
	}

	//получаем список записей по заданным критериям
	str_first := strconv.Itoa(gcfg_cnt_posts_on_page + 1)
	str_skip := strconv.Itoa(gcfg_cnt_posts_on_page * (page - 1))

	query := "SELECT FIRST " + str_first + " SKIP " + str_skip + " "
	query += `  p.name,p.tags,COALESCE(p.preview,LEFT(p.text,1200)),p.uuid_user,LEFT(p.date_create,16),p.uuid FROM tpost p WHERE 1=1
				ORDER BY p.date_create DESC
			`
	db := d["db"].(*DBd).DB
	rows, err := db.Query(query)
	if err != nil {
		d["error"] = fmtError("http_search__load_data ERROR001 db.Query(query): query:\n"+query+"\n\n", err)
		return
	}

	data_rows := make([]map[string]string, 0)
	cnt := 0
	for rows.Next() {
		var name, tags, preview, uuid_user, date_create, uuid NullString
		if err := rows.Scan(&name, &tags, &preview, &uuid_user, &date_create, &uuid); err != nil {
			d["error"] = fmtError("http_search__load_data ERROR002 rows.Scan: query:\n"+query+"\n\n", err)
			return
		}
		cnt++
		if cnt > gcfg_cnt_posts_on_page {
			data["has_next_page"] = 1
			break
		}
		dr := make(map[string]string)
		dr["name"] = name.get_trcp1251("")
		dr["tags"] = tags.get_trcp1251("")
		dr["preview"] = preview.get_trcp1251("")
		dr["uuid_user"] = uuid_user.get("")
		dr["date_create"] = date_create.get("")
		dr["uuid"] = uuid.get("")
		data_rows = append(data_rows, dr)
	}
	rows.Close()
	data["rows"] = data_rows
	d["data"] = data
	return
}
