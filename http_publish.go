package main

import (
	"fmt"
	"net/http"
	"net/url"
	//"strconv"
	//"errors"
	"regexp"
	"strings"
	"time"

	mf "github.com/mixamarciv/gofncstd3000"
	//"html/template"
)

//хендлер для /publish
func http_publish(w http.ResponseWriter, r *http.Request) {
	d := http_parse_url(w, r)

	http_parse_url__get_db(w, r, d)
	if d["error"] != nil {
		RenderJson(w, r, d)
		return
	}

	if d["db_access"].(string) < "3" {
		d["error"] = fmt.Errorf("%s", "у вас нет доступа к публикации записей БД \""+d["db"].(*DBd).Name+"\"")
		d["errorcode"] = "dbnoaccess"
		RenderJson(w, r, d)
		return
	}

	get_vars := d["get_vars"].(url.Values)
	rtype := http_get_var_str(get_vars, "type", "")

	if rtype == "publish" {
		http_publish__publish(w, r, d)
		return
	}

	d["error"] = fmt.Errorf("%s", "у не верно указаны GET параметры")
	d["errorcode"] = "nogetparam"
	RenderJson(w, r, d)
}

//сохраняем слова и количество их повторений из поста в таблицы tword и tword_post
//  d["data"] - данные поста если все успешно
//  d["error"] - если ошибки
func http_publish__publish(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("%#v", r.Form))
	//LogPrint("=====================================================================")
	get_vars := d["get_vars"].(url.Values)
	uuid := http_get_var_str_sql(get_vars, "id", "")
	if uuid == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан GET параметр \"id\":\""+uuid+"\"")
		d["errorcode"] = "nopostparam"
		return
	}

	//получаем данные обновляемого поста
	query := `SELECT FIRST 1 SKIP 0 
	            p.name,p.tags,p.preview,p.text,p.edit_type,LEFT(p.date_modify,16) AS date_modify
	          FROM tpost p WHERE p.uuid='` + uuid + `' AND p.edit_type='publish' `
	rows := run_db_query(d, query)
	if rows == nil {
		return
	}

	dr := make(map[string]string)
	for rows.Next() {
		var name, tags, preview, text, edit_type, date_modify NullString
		if err := rows.Scan(&name, &tags, &preview, &text, &edit_type, &date_modify); err != nil {
			d["error"] = fmtError("http_publish__publish ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return
		}
		dr["name"] = name.get("")
		dr["tags"] = tags.get("")
		dr["preview"] = preview.get("")
		dr["text"] = text.get("")
		dr["edit_type"] = edit_type.get("")
		dr["date_modify"] = date_modify.get("")
		dr["date_modify2"] = mf.StrRegexpReplace(dr["date_modify"], "[^\\d]", "")
		//d["data"] = dr
	}
	rows.Close()

	log_file := "/public/publish_log/" + dr["date_modify2"]
	mf.MkdirAll(apppath + log_file)
	log_file += "/" + uuid + ".log"
	d["log_file"] = log_file
	d["log_file_fullpath"] = apppath + log_file

	{ //отправляем пользователю инфо где искать логи
		ret := make(map[string]interface{})
		ret["log_file"] = d["log_file"]
		RenderJson(w, r, ret)
	}

	//далее формируем эти логи:
	write_fnc_log(r, d, mf.CurTimeStr()+"  start update")
	{
		re := regexp.MustCompile("[^\\w]+")
		w_name := http_publish__get_words(re, dr["name"])
		w_preview := http_publish__get_words(re, dr["preview"])
		w_tags := http_publish__get_words(re, dr["tags"])
		w_text := http_publish__get_words(re, dr["text"])
	}
	write_fnc_log(r, d, mf.CurTimeStr()+"  end update")
	return
}

//пишем лог и выводим время каждой записи
func write_fnc_log(r *http.Request, d map[string]interface{}, s string) {
	tt := GetLoadTimeStr(r)
	a, b := d["calc_next_load_time"] //время с предыдущего вызова этой функции
	if !b {
		a = time.Now()
	}
	tl := a.(time.Time)
	d["calc_next_load_time"] = time.Now()
	mf.FileAppendStr(d["log_file_fullpath"].(string), s+fmt.Sprintf("    %v / %v\n", time.Now().Sub(tl), tt))
}

//разбиваем s на слова согласно указанному регулярному выражению
func http_publish__get_words(re regexp.Regexp, s string) map[string]int {
	s = strings.ToLower(s)
	a := re.Split(s, -1)
	words := make(map[string]int)
	for _, w := range a {
		words[w] = words[w] + 1
	}
	return words
}
