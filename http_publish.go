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
	id := http_get_var_str_sql(get_vars, "id", "")
	if id == "all" {
		http_publish__publish_all(w, r, d)
		return
	}

	rtype := http_get_var_str(get_vars, "type", "")

	if rtype == "publish" {
		http_publish__publish(w, r, d)
		return
	}

	d["error"] = fmt.Errorf("%s", "у не верно указаны GET параметры")
	d["errorcode"] = "nogetparam"
	RenderJson(w, r, d)
}

//запускает http_publish__publish для каждого поста бд
func http_publish__publish_all(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	//получаем данные обновляемых постов
	query := `SELECT p.uuid FROM tpost p WHERE p.uuid IS NOT NULL AND p.edit_type='publish' ORDER BY p.create_date`
	rows := run_db_query(d, query)
	if rows == nil {
		return
	}

	//a := make([]string, 0)
	d["publis_all"] = 1
	get_vars := d["get_vars"].(url.Values)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			d["error"] = fmtError("http_publish__publish_all ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return
		}

		get_vars.Set("id", id)
		d["get_vars"] = get_vars

		http_publish__publish(w, r, d)

	}
	rows.Close()
}

//сохраняем слова и количество их повторений из поста в таблицы tword и tword_post
//  d["data"] - данные поста если все успешно
//  d["error"] - если ошибки
func http_publish__publish(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("%#v", r.Form))
	//LogPrint("=====================================================================")
	_, publis_all := d["publis_all"]
	get_vars := d["get_vars"].(url.Values)
	uuid := http_get_var_str_sql(get_vars, "id", "")
	if uuid == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан GET параметр \"id\":\""+uuid+"\"")
		d["errorcode"] = "nopostparam"
		return
	}

	d["uuid"] = uuid
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

	curdate := mf.StrRegexpReplace(mf.CurTimeStr(), "[^\\d]", "")
	log_file := "/public/publish_log/" + curdate[0:4] + "/" + curdate[4:8]
	mf.MkdirAll(apppath + log_file)
	log_file += "/" + curdate + "_" + dr["date_modify2"] + "_" + uuid + ".log"
	d["log_file"] = log_file
	d["log_file_fullpath"] = apppath + log_file

	if !publis_all { //отправляем пользователю инфо где искать логи
		ret := make(map[string]interface{})
		ret["log_file"] = d["log_file"]
		RenderJson(w, r, ret)
	}

	//далее формируем эти логи:
	write_fnc_log(r, d, mf.CurTimeStr()+"  start update")
	{
		re := regexp.MustCompile("[^a-zа-я0-9]+")
		m := make(map[string]map[string]int)
		k := make(map[string]int)

		m["name"] = http_publish__get_words(re, dr["name"])
		k["name"] = 7

		m["preview"] = http_publish__get_words(re, dr["preview"])
		k["preview"] = 2

		m["tags"] = http_publish__get_words(re, dr["tags"])
		k["tags"] = 10

		m["text"] = http_publish__get_words(re, dr["text"])
		k["text"] = 1

		write_fnc_log(r, d, mf.CurTimeStr()+"  get words")

		all_words := make(map[string]int)
		cnt_all := 0
		for ppart, pwords := range m {
			write_fnc_log_row(r, d, ppart+fmt.Sprintf("\t\twords/cnt: %d / %d", pwords["#cnt uniq words"], pwords["#cnt words"]))
			for word, cnt := range pwords {
				if word[0:1] == "#" {
					continue
				}
				write_fnc_log_row(r, d, fmt.Sprintf("\t%d\t%s", cnt, word))
				cnt_all += cnt * k[ppart]
				all_words[word] = all_words[word] + cnt*k[ppart]
			}
		}
		write_fnc_log(r, d, mf.CurTimeStr()+"  calc words")

		write_fnc_log_row(r, d, fmt.Sprintf("all words/weight: %d / %d", len(all_words), cnt_all))
		for word, cnt := range all_words {
			write_fnc_log_row(r, d, fmt.Sprintf("\t%d\t%s", cnt, word))
		}
		write_fnc_log(r, d, mf.CurTimeStr()+"  print all words")

		d["all_words"] = all_words
		b := http_publish__db_update_post_words(r, d)
		if b < 0 {
			write_fnc_log(r, d, mf.CurTimeStr()+"  ERROR: ")
			write_fnc_log_row(r, d, sprintf("%s\n%v\n", d["errorcode"], d["error"]))
		}
		write_fnc_log(r, d, mf.CurTimeStr()+"  exec updates db sql query")
	}
	write_fnc_log(r, d, mf.CurTimeStr()+"  end update")
	write_fnc_log_row(r, d, "--==## end ##==--")
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
func write_fnc_log_row(r *http.Request, d map[string]interface{}, s string) {
	mf.FileAppendStr(d["log_file_fullpath"].(string), s+"\n")
}

//разбиваем s на слова согласно указанному регулярному выражению
func http_publish__get_words(re *regexp.Regexp, s string) map[string]int {
	s = mf.StrRegexpReplace(s, "<\\/?[a-z]+.*>", "")
	s = strings.ToLower(s)
	a := re.Split(s, -1)
	words := make(map[string]int)
	cnt := 0

	re_test := regexp.MustCompile("^[а-я]$") //кривая проверка на длину в 1 юникод символ
	for _, w := range a {
		w = strings.Trim(w, "\t\r\n\f\v ")
		if w == "" || len(w) < 2 {
			continue
		}
		if re_test.MatchString(w) {
			continue
		}
		words[w] = words[w] + 1
		cnt++
	}
	l := len(words)
	words["#cnt words"] = cnt
	words["#cnt uniq words"] = l
	return words
}

type publish_post_word struct {
	Word           string
	Id_word        int64
	Cnt            int64 //количество повторений этого слова в посте
	Cnt_new        int64 //новое количество повторений в этом посте
	Cnt_word_total int64 //всего количество повторений этого слова в бд
}

//сохраняем текущее количество повторений слов в посте
func http_publish__db_update_post_words(r *http.Request, d map[string]interface{}) int {
	new_words := d["all_words"].(map[string]int)
	uuid := d["uuid"].(string)

	upd_words := make(map[string]*publish_post_word)

	//получаем данные обновляемого поста
	query := `SELECT w.id,w.word,p.cnt,(SELECT SUM(tp.cnt) FROM tword_post tp WHERE tp.id_word=w.id) AS wtotal_cnt
	          FROM tword_post p 
			    LEFT JOIN tword w ON w.id=p.id_word
			  WHERE p.uuid_post='` + uuid + `' `
	rows := run_db_query(d, query)
	if rows == nil {
		return -1
	}
	for rows.Next() {
		p := new(publish_post_word)
		if err := rows.Scan(&p.Id_word, &p.Word, &p.Cnt, &p.Cnt_word_total); err != nil {
			d["error"] = fmtError("http_publish__db_update_post_words ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return -1
		}

		new_cnt := int64(new_words[p.Word])
		if new_cnt == p.Cnt {
			continue
		}

		p.Cnt_new = new_cnt
		upd_words[p.Word] = p
		delete(new_words, p.Word)
	}
	rows.Close()

	//получаем список новых слов добавляемых в словарь
	for word, cnt := range new_words {
		p := new(publish_post_word)
		p.Word = word
		p.Cnt_new = int64(cnt)

		new_id_word, cnt_total := http_publish__db_update_post_words__get_id_word_and_cnt(d, p.Word)
		if new_id_word < 0 {
			return -1
		}
		p.Id_word = new_id_word
		p.Cnt_word_total = cnt_total

		upd_words[p.Word] = p
		delete(new_words, p.Word)
	}

	//загружаем все слова в бд
	for _, p := range upd_words {
		if p.Cnt_new == 0 { //если слова больше нет в этом посте
			//удаляем слово из слов поста
			b := run_db_exec(d, `DELETE FROM tword_post p WHERE p.uuid_post='`+uuid+`' AND p.id_word=`+i64toa(p.Id_word))
			if !b.(bool) {
				return -1
			}
			//удаляем слово вообще из бд
			if p.Cnt == p.Cnt_word_total {
				b := run_db_exec(d, `DELETE FROM tword w WHERE w.id=`+i64toa(p.Id_word))
				if !b.(bool) {
					return -1
				}
			}
			continue
		}
		if p.Cnt > 0 { //если слово уже было в этом посте
			//то меняем количество его повторений на новое
			b := run_db_exec(d, `UPDATE tword_post p SET p.cnt=`+i64toa(p.Cnt_new)+` WHERE p.uuid_post='`+uuid+`' AND p.id_word=`+i64toa(p.Id_word))
			if !b.(bool) {
				return -1
			}
			continue
		}
		if p.Cnt == 0 { //если этого слова небыло в этом посте
			//то создаем новую запись слова количества с количеством повторений для этого поста
			b := run_db_exec(d, `INSERT INTO tword_post(uuid_post,id_word,cnt) VALUES('`+uuid+`',`+i64toa(p.Id_word)+`,`+i64toa(p.Cnt_new)+`)`)
			if !b.(bool) {
				return -1
			}
			continue
		}
	}
	return 1
}

//возвращает новый id для tword.id
func http_publish__db_update_post_words__gen_id_word(d map[string]interface{}) int64 {
	query := `SELECT gen_id(id_word,1) FROM rdb$database`
	rows := run_db_query(d, query)
	if rows == nil {
		return -1
	}
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			d["error"] = fmtError("http_publish__db_update_post_words__gen_id_word ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return -1
		}
	}
	rows.Close()
	return id
}

//получаем id слова и количество его повторений в бд всего
func http_publish__db_update_post_words__get_id_word_and_cnt(d map[string]interface{}, word string) (int64, int64) {
	query := `SELECT w.id,COALESCE((SELECT SUM(tp.cnt) FROM tword_post tp WHERE tp.id_word=w.id),0) AS wtotal_cnt
	          FROM tword w WHERE w.word = '` + word + `'`
	rows := run_db_query(d, query)
	if rows == nil {
		return -1, -1
	}
	var id, cnt int64
	for rows.Next() {
		if err := rows.Scan(&id, &cnt); err != nil {
			d["error"] = fmtError("http_publish__db_update_post_words__get_id_word_and_cnt ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return -1, -1
		}
	}
	rows.Close()

	if id == 0 { //если такого слова нет в бд то создаем его
		id = http_publish__db_update_post_words__gen_id_word(d)
		if id < 0 {
			return -1, -1
		}
		b := run_db_exec(d, `INSERT INTO tword(id,word) VALUES(`+i64toa(id)+`,'`+word+`')`)
		if !b.(bool) {
			return -1, -1
		}
	}
	return id, cnt
}
