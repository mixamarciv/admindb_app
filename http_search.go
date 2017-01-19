package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	//"strings"
	"regexp"

	mf "github.com/mixamarciv/gofncstd3000"
)

var regexp_opentag *regexp.Regexp

func init() {
	regexp_opentag_text := "<[a-zA-Z]"
	var err error
	regexp_opentag, err = mf.RegexpCompile(regexp_opentag_text)
	LogPrintErrAndExit("Ошибка компиляции регулярного выражения \""+regexp_opentag_text+"\"", err)
}

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
	query += `  p.name,p.tags,p.preview,p.text,p.uuid_user,LEFT(p.date_create,16),p.uuid FROM tpost p WHERE 1=1
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
		var name, tags, preview, text, uuid_user, date_create, uuid NullString
		if err := rows.Scan(&name, &tags, &preview, &text, &uuid_user, &date_create, &uuid); err != nil {
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
		tpreview := preview.get_trcp1251_long("")
		if tpreview == "" {
			tpreview = text.get_trcp1251_long("")
			i := len(tpreview)
			add_text := ""
			max_post_len := 1000
			if i > max_post_len {
				i = max_post_len
				add_text = "..."
			}

			tpreview = tpreview[:i]

			ja := regexp_opentag.FindStringIndex(tpreview)
			if len(ja) > 0 {
				tpreview = tpreview[:ja[0]]
				add_text = "..."
			}
			tpreview += add_text
		}
		dr["preview"] = tpreview

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
