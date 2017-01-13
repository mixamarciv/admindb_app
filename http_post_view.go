package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	//mf "github.com/mixamarciv/gofncstd3000"
)

//хендлер для /p
func http_post_view(w http.ResponseWriter, r *http.Request) {
	d := http_parse_url(w, r)

	http_parse_url__get_db(w, r, d)
	if d["error"] != nil {
		RenderError(w, r, d)
		return
	}

	http_post_view__load_data(w, r, d)

	RenderTemplate(w, r, d, "maintemplate.html", "post_view.html")
}

//загрузка данных в d["data"] по заданным критериям из d["url_vars"]
func http_post_view__load_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	data := make(map[string]interface{})
	get_vars := d["get_vars"].(url.Values)
	id := http_get_var_str(get_vars, "i", "")
	if id == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан GET параметр \"i\"")
		d["errorcode"] = "nogetparam"
		RenderError(w, r, d)
		return
	}

	query := "SELECT FIRST 1 SKIP 0 "
	query += `  p.name,p.tags,COALESCE(p.preview,LEFT(p.text,1200)),p.uuid_user,LEFT(p.date_create,16),p.uuid,p.text FROM tpost p WHERE 1=1
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
