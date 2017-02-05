package main

import (
	"fmt"
	"net/http"
	"net/url"
	//"strconv"
	"strings"

	//mf "github.com/mixamarciv/gofncstd3000"
	//"html/template"
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
	if d["error"] != nil {
		RenderError(w, r, d)
		return
	}

	RenderTemplate(w, r, d, "maintemplate.html", "post_view.html")
}

//загрузка данных в d["data"] по заданным критериям из d["url_vars"]
//должен быть задан параметр d["get_vars"]["id"]
//возвращает:
//  d["data"] - данные поста
//
func http_post_view__load_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	//data := make(map[string]interface{})
	get_vars := d["get_vars"].(url.Values)
	id := http_get_var_str(get_vars, "id", "")
	if id == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан GET параметр \"id\"")
		d["errorcode"] = "nogetparam"
		return
	}

	id = strings.Replace(id, "'", "''", -1)
	query := "SELECT FIRST 1 SKIP 0 "
	query += `  p.name,p.tags,p.preview,p.text,p.uuid_user,LEFT(p.date_create,16),p.uuid
	          FROM tpost p WHERE p.uuid='` + id + `'
			  ORDER BY p.date_create DESC
			`
	db := d["db"].(*DBd).DB
	rows, err := db.Query(query)
	if err != nil {
		d["error"] = fmtError("http_post_view__load_data ERROR001 db.Query(query): query:\n"+query+"\n\n", err)
		d["errorcode"] = "dbqueryerror"
		return
	}

	cnt := 0
	for rows.Next() {
		var name, tags, preview, text, uuid_user, date_create, uuid NullString
		if err := rows.Scan(&name, &tags, &preview, &text, &uuid_user, &date_create, &uuid); err != nil {
			d["error"] = fmtError("http_post_view__load_data ERROR002 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return
		}
		cnt++
		dr := make(map[string]interface{})
		dr["name"] = name.get_trcp1251("")
		dr["tags"] = tags.get_trcp1251("")
		dr["preview"] = preview.get_trcp1251("")
		dr["text"] = text.get_trcp1251_long("")
		/*dr["text_html"] = template.HTML(text.get_trcp1251_long(""))*/
		dr["uuid_user"] = uuid_user.get("")
		dr["date_create"] = date_create.get("")
		dr["uuid"] = uuid.get("")
		d["data"] = dr
	}
	rows.Close()

	if cnt == 0 {
		d["error"] = "запрашиваемая запись \"" + id + "\" не найдена"
		d["errorcode"] = "postnotfound"
		return
	}

	return
}
