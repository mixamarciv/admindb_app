package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

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

//хендлер для /sq
func http_searchq(w http.ResponseWriter, r *http.Request) {
	d := http_parse_url(w, r)

	http_parse_url__get_db(w, r, d)
	if d["error"] != nil {
		RenderTemplate(w, r, d, "block_content.html")
		return
	}

	if d["db_access"].(string) < "1" {
		d["error"] = fmt.Errorf("%s", "у вас нет доступа к БД \""+d["db"].(*DBd).Name+"\"")
		d["errorcode"] = "dbnoaccess"
		RenderTemplate(w, r, d, "block_content.html")
		return
	}

	http_search__load_data(w, r, d)

	RenderTemplate(w, r, d, "block_content.html", "{{define \"content\"}}{{block \"search_data\" .}}search_data{{end}}{{end}}", "search_data.html")
}

//загрузка данных в d["data"] по заданным критериям из d["url_vars"]
func http_search__load_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	data := make(map[string]interface{})
	get_vars := d["get_vars"].(url.Values)

	page := http_get_var_int(get_vars, "p", 1)
	data["page"] = page
	data["has_next_page"] = 0

	{ //задаем url_next_page и url_prev_page
		turlstr := "/s?" + r.URL.RawQuery
		u, err := url.Parse(turlstr)
		if err != nil {
			d["error"] = fmt.Errorf("%s", "http_search__load_data ERROR003: url.Parse("+turlstr+")")
			d["errorcode"] = "parseurlerror"
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
	cnt_select, join_filter := http_search__get_filter(w, r, d)

	query :=
		`SELECT FIRST ` + str_first + ` SKIP ` + str_skip +
			` p.name,p.tags,p.preview,p.text,p.uuid_user,LEFT(p.date_create,16),p.uuid,` +
			` ` + cnt_select + ` AS cnt` +
			` FROM tpost p` + "\n" +
			` ` + join_filter + "\n" +
			` WHERE p.edit_type='publish'` +
			` ORDER BY p.date_create DESC`

	/******************
	//в итоге запрос должен быть вида:
	SELECT
	  --FIRST 20 SKIP 0
	  p.name,p.tags,p.preview,p.text,p.uuid_user,LEFT(p.date_create,16) AS dt,p.uuid,
	  c1.cnt+c2.cnt+c3.cnt AS cnt
	FROM tpost p
	  INNER JOIN (
	    SELECT wp.uuid_post,SUM(wp.cnt) AS cnt FROM tword_post wp INNER JOIN tword w ON w.id=wp.id_word
	    WHERE w.word STARTS WITH 'word1'  GROUP BY wp.uuid_post
	    ) AS c1 ON c1.uuid_post=p.uuid
	  INNER JOIN (
	    SELECT wp.uuid_post,SUM(wp.cnt) AS cnt FROM tword_post wp INNER JOIN tword w ON w.id=wp.id_word
	    WHERE w.word STARTS WITH 'word2'  GROUP BY wp.uuid_post
	    ) AS c2 ON c2.uuid_post=p.uuid
	  INNER JOIN (
	    SELECT wp.uuid_post,SUM(wp.cnt) AS cnt FROM tword_post wp INNER JOIN tword w ON w.id=wp.id_word
	    WHERE w.word STARTS WITH 'word3'  GROUP BY wp.uuid_post
	    ) AS c3 ON c3.uuid_post=p.uuid
	WHERE  p.edit_type='publish'
	ORDER BY cnt,p.date_create DESC
	******************/
	d["query"] = query

	db := d["db"].(*DBd).DB
	rows, err := db.Query(query)
	if err != nil {
		d["error"] = fmtError("http_search__load_data ERROR001 db.Query(query): query:\n"+query+"\n\n", err)
		d["errorcode"] = "dbqueryerror"
		return
	}

	data_rows := make([]map[string]string, 0)
	cnt := 0
	for rows.Next() {
		var name, tags, preview, text, uuid_user, date_create, uuid NullString
		var sum_weight_words int
		if err := rows.Scan(&name, &tags, &preview, &text, &uuid_user, &date_create, &uuid, &sum_weight_words); err != nil {
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

//возвращает условие where для запроса
func http_search__get_filter(w http.ResponseWriter, r *http.Request, d map[string]interface{}) (string, string) {
	get_vars := d["get_vars"].(url.Values)
	s := http_get_var_str_sql(get_vars, "s", "")
	s = mf.StrTrim(s)
	if len(s) == 0 {
		return "0", ""
	}
	s = strings.ToLower(s)

	//разбиваем текст на искомые выражения
	re := regexp.MustCompile("[^a-zа-я0-9_\\!=]+")
	a := re.Split(s, -1)
	sw := make(map[string]int) //список искомых выражений
	cnt := 0

	re_test := regexp.MustCompile("[a-zа-я0-9_]{2,}$") //проверка на длину минимум в 2 символа

	for _, w := range a {
		w = strings.Trim(w, "\t\r\n\f\v ")
		if w == "" || len(w) < 2 {
			continue
		}
		if !re_test.MatchString(w) {
			continue
		}
		if sw[w] > 0 { //если такое слово уже есть в списке поиска
			continue
		}
		sw[w] = 1
		cnt++
	}

	if cnt == 0 {
		return "0", ""
	}

	cnt_select, join_filter := "", ""
	i := 0
	for w, _ := range sw {
		if i > 0 {
			cnt_select += "+"
		}
		i++
		alias := "c" + itoa(i)
		cnt_select += alias + ".cnt"

		filter := " STARTS WITH '" + w + "'"
		if w[0:1] == "=" {
			filter = " = '" + w[1:] + "'"
		}
		join_filter += "\nINNER JOIN (" +
			`SELECT wp.uuid_post,SUM(wp.cnt) AS cnt ` +
			`FROM tword_post wp ` +
			`  INNER JOIN tword w ON w.id=wp.id_word ` +
			`WHERE w.word` + filter + ` GROUP BY wp.uuid_post` +
			`) AS ` + alias + ` ON ` + alias + `.uuid_post=p.uuid `

	}

	return cnt_select, join_filter
}
