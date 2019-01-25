package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	mf "github.com/mixamarciv/gofncstd3000"
)

//хендлер для /e
func http_post_edit(w http.ResponseWriter, r *http.Request) {
	d := http_parse_url(w, r)

	http_parse_url__get_db(w, r, d)
	if d["error"] != nil {
		RenderError(w, r, d)
		return
	}

	if d["db_access"].(string) < "2" {
		d["error"] = fmt.Errorf("%s", "у вас нет доступа к редактированию записей БД \""+d["db"].(*DBd).Name+"\"")
		d["errorcode"] = "dbnoaccess"
		RenderError(w, r, d)
		return
	}

	http_post_edit__load_data(w, r, d)
	if d["error"] != nil {
		RenderError(w, r, d)
		return
	}

	RenderTemplate(w, r, d, "maintemplate.html", "post_edit.html")
}

//загрузка данных в d["data"] по заданным критериям из d["url_vars"]
//должен быть задан параметр d["get_vars"]["id"]
//возвращает:
//  d["data"] - данные поста
//
func http_post_edit__load_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	//data := make(map[string]interface{})
	get_vars := d["get_vars"].(url.Values)

	uuid_user_this := get_map_val(d, "", "user", "uuid_user").(string)
	uuid_user := http_get_var_str(get_vars, "user", "")
	if uuid_user == "" {
		uuid_user = uuid_user_this
	}

	if uuid_user != uuid_user_this { //если юзер пытается загрузить для редактирования чужую запись
		if d["db_access"].(string) < "3" { // то у него должны быть права модератора
			d["error"] = fmt.Errorf("%s", "у вас нет доступа к редактированию чужих записей в БД \""+d["db"].(*DBd).Name+"\"")
			d["errorcode"] = "dbnoaccess"
			return
		}
	}

	id := http_get_var_str(get_vars, "id", "")
	if id == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан GET параметр \"id\"")
		d["errorcode"] = "nogetparam"
		return
	}

	id = strings.Replace(id, "'", "''", -1)
	uuid_user = strings.Replace(uuid_user, "'", "''", -1)

	http_post_edit__load_data_fn(w, r, d, id, uuid_user)

	return
}

//только загрузка данных tpost в d["data"] по id и uuid_user
//возвращает:
//  d["data"] - данные поста
// или d["error"] - в случае ошибки
func http_post_edit__load_data_fn(w http.ResponseWriter, r *http.Request, d map[string]interface{}, id, uuid_user string) {
	if id == "new" {
		//если создается новая запись
		//то в любом случае создаем её от имени текущего пользователя
		uuid_user_this := get_map_val(d, "", "user", "uuid_user").(string)

		dr := make(map[string]interface{})
		dr["name"] = ""
		dr["tags"] = ""
		dr["preview"] = ""
		dr["text"] = ""
		dr["uuid_user"] = uuid_user_this
		dr["uuid_user_create"] = uuid_user_this
		dr["uuid_user_publish"] = ""
		dr["date_create"] = ""
		dr["date_modify"] = ""
		dr["edit_type"] = "new"
		dr["uuid"] = strings.Replace(mf.StrUuid(), "-", "", -1)
		LogPrint("gen new uuid: " + dr["uuid"].(string))
		d["data"] = dr
		return
	}

	query := `SELECT * FROM
	          (SELECT FIRST 1 SKIP 0 
	            0 AS n,p.name,p.tags,p.preview,p.text,p.uuid_user,LEFT(p.date_create,16) AS date_create,p.uuid,p.edit_type,p.uuid_user_create,p.uuid_user_publish,LEFT(p.date_modify,16) AS date_modify
	          FROM tpost p WHERE p.uuid='` + id + `'
			  UNION ALL
			  SELECT FIRST 1 SKIP 0 
	            1 AS n,p.name,p.tags,p.preview,p.text,p.uuid_user,LEFT(p.date_create,16) AS date_create,p.uuid,p.edit_type,p.uuid_user_create,p.uuid_user_publish,LEFT(p.date_modify,16) AS date_modify
	          FROM tpost p WHERE p.uuid='` + id + `' AND p.uuid_user='` + uuid_user + `' AND p.edit_type!='publish'
			  )
			  ORDER BY n
			`
	rows := run_db_query(d, query)
	if rows == nil {
		return
	}

	cnt := 0
	for rows.Next() {
		var n, name, tags, preview, text, uuid_user, date_create, uuid, edit_type, uuid_user_create, uuid_user_publish, date_modify NullString
		if err := rows.Scan(&n, &name, &tags, &preview, &text, &uuid_user, &date_create, &uuid, &edit_type, &uuid_user_create, &uuid_user_publish, &date_modify); err != nil {
			d["error"] = fmtError("http_post_edit__load_data ERROR002 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return
		}
		cnt++
		dr := make(map[string]interface{})
		dr["name"] = name.get("")
		dr["tags"] = tags.get("")
		dr["preview"] = preview.get("")
		dr["text"] = text.get("")
		/*dr["text_html"] = template.HTML(text.get_trcp1251_long(""))*/
		dr["uuid_user"] = uuid_user.get("")
		dr["uuid_user_create"] = uuid_user_create.get("")
		dr["uuid_user_publish"] = uuid_user_publish.get("")
		dr["date_create"] = date_create.get("")
		dr["date_modify"] = date_modify.get("")
		dr["edit_type"] = edit_type.get("")
		dr["uuid"] = uuid.get("")
		d["data"] = dr
	}
	rows.Close()

	if cnt == 0 {
		d["error"] = fmtError("ERROR: пост не найден db/id: "+d["db"].(*DBd).ShortName+"/"+id, errors.New("пост не найден"))
		d["errorcode"] = "notfoundpost"
		return
	}
}

//хендлер для /e_ajax
func http_post_edit_ajax(w http.ResponseWriter, r *http.Request) {
	d := http_parse_url(w, r)

	http_parse_url__get_db(w, r, d)
	if d["error"] != nil {
		RenderJson(w, r, d)
		return
	}

	if d["db_access"].(string) < "2" {
		d["error"] = fmt.Errorf("%s", "у вас нет доступа к редактированию записей БД \""+d["db"].(*DBd).Name+"\"")
		d["errorcode"] = "dbnoaccess"
		RenderJson(w, r, d)
		return
	}

	get_vars := d["get_vars"].(url.Values)
	rtype := http_get_var_str(get_vars, "type", "")

	if rtype == "load" {
		http_post_edit__load_data(w, r, d)
		RenderJson(w, r, d)
		return
	}

	if rtype == "save" {
		http_post_edit__save_data(w, r, d)
		RenderJson(w, r, d)
		return
	}

	if rtype == "delete" {
		http_post_edit__delete_data(w, r, d)
		RenderJson(w, r, d)
		return
	}

	if rtype == "publish" {
		http_post_edit__publish(w, r, d)
		RenderJson(w, r, d)
		return
	}

	d["error"] = fmt.Errorf("%s", "у не верно указаны GET параметры")
	d["errorcode"] = "nogetparam"
	RenderJson(w, r, d)
}

//должены быть заданы post данные
//  d["data"]["uuid"] - uuid поста если все успешно
//  d["error"] - если ошибки
func http_post_edit__save_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	r.ParseForm()
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("%#v", r.Form))
	//LogPrint("=====================================================================")

	uuid_user_this := get_map_val(d, "", "user", "uuid_user").(string)
	uuid_user := http_get_var_str(r.Form, "uuid_user", "")
	if uuid_user == "" {
		uuid_user = uuid_user_this
	}

	if uuid_user != uuid_user_this { //если юзер пытается загрузить для редактирования чужую запись
		if d["db_access"].(string) < "3" { // то у него должны быть права модератора
			d["error"] = fmt.Errorf("%s", "у вас нет доступа к редактированию чужих записей в БД \""+d["db"].(*DBd).Name+"\"")
			d["errorcode"] = "dbnoaccess"
			return
		}
	}
	uuid_user = strings.Replace(uuid_user, "'", "''", -1)

	uuid := http_get_var_str_sql(r.Form, "uuid", "")
	if uuid == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан POST параметр \"uuid\":\""+uuid+"\"")
		d["errorcode"] = "nopostparam"
		return
	}

	//проверяем есть ли обновляемые записи в бд
	query := `SELECT 
				(SELECT COUNT(*) FROM tpost p WHERE p.uuid='` + uuid + `') AS c1,
				(SELECT COUNT(*) FROM tpost p WHERE p.uuid='` + uuid + `' AND p.uuid_user='` + uuid_user + `' AND p.edit_type!='publish') AS c2
			  FROM rdb$database`
	rows := run_db_query(d, query)
	if rows == nil {
		return
	}
	var c1, c2 int
	for rows.Next() {
		if err := rows.Scan(&c1, &c2); err != nil {
			d["error"] = fmtError("http_post_edit__save_data ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return
		}
	}
	rows.Close()

	data := make(map[string]interface{})
	d["data"] = data

	name := http_get_var_str_sql(r.Form, "name", "")
	text := http_get_var_str_sql(r.Form, "text", "")
	tags := http_get_var_str_sql(r.Form, "tags", "")
	preview := http_get_var_str_sql(r.Form, "preview", "")

	query = "hz che delatb nax"

	if c2 > 0 {
		//если пользователь ранее редактировал эту запись и она не публиковалась то просто обновляем его версию его записи
		data["edit_type"] = "update"
		query = `UPDATE TPOST p SET 
					NAME='` + name + `',
					TAGS='` + tags + `',
					TEXT='` + text + `',
					PREVIEW='` + preview + `',
					UUID_USER='` + uuid_user_this + `',
				    /*UUID_USER_PUBLISH='` + uuid_user_this + `',*/
					EDIT_TYPE='` + data["edit_type"].(string) + `'
			    WHERE p.uuid='` + uuid + `' 
				  AND p.uuid_user='` + uuid_user + `' 
				  AND p.edit_type!='publish'`
	} else if c1 > 0 {
		//если пользователь первый раз редактирует эту запись то создаем его версию этой записи
		data["edit_type"] = "update"
		query = `INSERT INTO TPOST(NAME,TAGS,TEXT,PREVIEW,UUID_USER,
				   DATE_CREATE,uuid_user_create,
				   UUID/*,UUID_USER_PUBLISH*/,EDIT_TYPE
				   )
			    VALUES('` + name + `','` + tags + `','` + text + `','` + preview + `','` + uuid_user_this + `',
				   (SELECT MIN(date_create) FROM tpost WHERE uuid='` + uuid + `'),(SELECT MIN(uuid_user_create) FROM tpost WHERE uuid='` + uuid + `'),
				   '` + uuid + `'/*,'` + uuid_user_this + `'*/,'` + data["edit_type"].(string) + `'
				   )`
	} else {
		//если пользователь создает новую запись то создаем его версию этой записи
		data["edit_type"] = "create"
		query = `INSERT INTO TPOST(NAME,TAGS,TEXT,PREVIEW,UUID_USER,
				   DATE_CREATE,
				   UUID/*,UUID_USER_PUBLISH*/,EDIT_TYPE,UUID_USER_CREATE
				   )
			    VALUES('` + name + `','` + tags + `','` + text + `','` + preview + `','` + uuid_user_this + `',
				   (SELECT MIN(date_create) FROM tpost WHERE uuid='` + uuid + `'),
				   '` + uuid + `'/*,'` + uuid_user_this + `'*/,'` + data["edit_type"].(string) + `','` + uuid_user_this + `'
				   )`
	}
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("edit_type: %#v", data["edit_type"]))
	//LogPrint(fmt.Sprintf("%#v", query))
	//LogPrint("=====================================================================")
	run_db_exec(d, query)

	http_post_edit__load_data_fn(w, r, d, uuid, uuid_user_this)

	return
}

//должены быть заданы post данные
//  d["data"]["uuid"] - uuid поста если все успешно
//  d["error"] - если ошибки
func http_post_edit__delete_data(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	r.ParseForm()
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("%#v", r.Form))
	//LogPrint("=====================================================================")

	uuid_user_this := get_map_val(d, "", "user", "uuid_user").(string)
	uuid_user := http_get_var_str(r.Form, "uuid_user", "")
	if uuid_user == "" {
		uuid_user = uuid_user_this
	}

	if uuid_user != uuid_user_this { //если юзер пытается загрузить для редактирования чужую запись
		if d["db_access"].(string) < "3" { // то у него должны быть права модератора
			d["error"] = fmt.Errorf("%s", "у вас нет доступа к редактированию чужих записей в БД \""+d["db"].(*DBd).Name+"\"")
			d["errorcode"] = "dbnoaccess"
			return
		}
	}
	uuid_user = strings.Replace(uuid_user, "'", "''", -1)

	uuid := http_get_var_str_sql(r.Form, "uuid", "")
	if uuid == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан POST параметр \"uuid\":\""+uuid+"\"")
		d["errorcode"] = "nopostparam"
		return
	}

	//проверяем есть ли обновляемые записи в бд
	query := `SELECT 
				(SELECT COUNT(*) FROM tpost p WHERE p.uuid='` + uuid + `') AS c1,
				(SELECT COUNT(*) FROM tpost p WHERE p.uuid='` + uuid + `' AND p.uuid_user='` + uuid_user + `' AND p.edit_type!='publish') AS c2
			  FROM rdb$database`
	rows := run_db_query(d, query)
	if rows == nil {
		return
	}
	var c1, c2 int
	for rows.Next() {
		if err := rows.Scan(&c1, &c2); err != nil {
			d["error"] = fmtError("http_post_edit__delete_data ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return
		}
	}
	rows.Close()

	data := make(map[string]interface{})
	d["data"] = data
	name := http_get_var_str_sql(r.Form, "name", "")
	text := http_get_var_str_sql(r.Form, "text", "")
	tags := http_get_var_str_sql(r.Form, "tags", "")
	preview := http_get_var_str_sql(r.Form, "preview", "")

	query = "hz che delatb nax"
	if c2 > 0 {
		//если пользователь ранее редактировал эту запись и она не публиковалась то просто обновляем его версию его записи
		data["edit_type"] = "delete"
		query = `UPDATE TPOST p SET 
					UUID_USER='` + uuid_user_this + `',
					EDIT_TYPE='` + data["edit_type"].(string) + `'
			    WHERE p.uuid='` + uuid + `' 
				  AND p.uuid_user='` + uuid_user + `' 
				  AND p.edit_type!='publish'`
	} else if c1 > 0 {
		//если пользователь первый раз редактирует эту запись то создаем его версию этой записи
		data["edit_type"] = "delete"
		query = `INSERT INTO TPOST(NAME,TAGS,TEXT,PREVIEW,UUID_USER,
				   DATE_CREATE,uuid_user_create,
				   UUID/*,UUID_USER_PUBLISH*/,EDIT_TYPE
				   )
			    VALUES('` + name + `','` + tags + `','` + text + `','` + preview + `','` + uuid_user_this + `',
				   (SELECT MIN(date_create) FROM tpost WHERE uuid='` + uuid + `'),(SELECT MIN(uuid_user_create) FROM tpost WHERE uuid='` + uuid + `'),
				   '` + uuid + `'/*,'` + uuid_user_this + `'*/,'` + data["edit_type"].(string) + `'
				   )`
	} else {
		//если запись не существует
		d["error"] = ("запись не найдена (возможно она уже удалена): uuid: " + uuid + " / uuid_user: " + uuid_user + " ")
		d["errorcode"] = "recordnotfound"
		return
	}
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("edit_type: %#v", data["edit_type"]))
	//LogPrint(fmt.Sprintf("%#v", query))
	//LogPrint("=====================================================================")
	run_db_exec(d, query)

	http_post_edit__load_data_fn(w, r, d, uuid, uuid_user_this)

	return
}

//обновляем данные поста в таблице tpost и выставляем edit_type == 'publish'
//должены быть заданы post данные
//  d["data"]["uuid"] - uuid поста если все успешно
//  d["error"] - если ошибки
func http_post_edit__publish(w http.ResponseWriter, r *http.Request, d map[string]interface{}) {
	r.ParseForm()
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("%#v", r.Form))
	//LogPrint("=====================================================================")

	if d["db_access"].(string) < "3" { //  у юзера должны быть права модератора
		d["error"] = fmt.Errorf("%s", "у вас нет доступа к редактированию чужих записей в БД \""+d["db"].(*DBd).Name+"\"")
		d["errorcode"] = "dbnoaccess"
		return
	}
	uuid_user_this := get_map_val(d, "", "user", "uuid_user").(string)
	uuid_user := http_get_var_str(r.Form, "uuid_user", "")
	if uuid_user == "" {
		uuid_user = uuid_user_this
	}
	uuid_user = strings.Replace(uuid_user, "'", "''", -1)

	uuid := http_get_var_str_sql(r.Form, "uuid", "")
	if uuid == "" {
		d["error"] = fmt.Errorf("%s", "не верно указан/не указан POST параметр \"uuid\":\""+uuid+"\"")
		d["errorcode"] = "nopostparam"
		return
	}

	//проверяем есть ли обновляемые записи в бд
	query := `SELECT 
				(SELECT COUNT(*) FROM tpost p WHERE p.uuid='` + uuid + `' AND p.edit_type='publish') AS c1,
				(SELECT COUNT(*) FROM tpost p WHERE p.uuid='` + uuid + `' AND p.uuid_user='` + uuid_user + `' AND p.edit_type!='publish') AS c2,
				(SELECT COUNT(*) FROM tpost p WHERE p.uuid='` + uuid + `' AND p.uuid_user='` + uuid_user + `' AND p.edit_type='delete') AS c3
			  FROM rdb$database`
	rows := run_db_query(d, query)
	if rows == nil {
		return
	}
	var c1, c2, c3 int
	for rows.Next() {
		if err := rows.Scan(&c1, &c2, &c3); err != nil {
			d["error"] = fmtError("http_post_edit__publish ERROR001 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
			return
		}
	}
	rows.Close()

	data := make(map[string]interface{})
	d["data"] = data

	if c3 == 0 {
		if c2 > 0 && c1 > 0 {
			//если пост существует и существует ранее не публиковавшаяся отредактированная запись этого поста указанным юзером
			//то обновляем текущий опубликованный пост из отредактированной записи и удаляем её за ненадобностью

			//обновляем существующую опубликованную запись (именно не удаляем а меняем значения полей - это надо для логов)
			query = `UPDATE tpost p SET 
					p.name =         (SELECT t.name FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.text =         (SELECT t.text FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.tags =         (SELECT t.tags FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.preview =      (SELECT t.preview FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.uuid_user =    (SELECT t.uuid_user FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					
					p.date_modify   = current_timestamp,
					p.date_publish  = current_timestamp,
					p.uuid_user_publish = '` + uuid_user_this + `',
					p.edit_type = 'publish'
				--FROM tpost p 
				--  LEFT JOIN tpost t ON t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'
			    WHERE p.uuid='` + uuid + `' 
				  AND p.edit_type='publish'
				`
			b := run_db_exec(d, query)
			if b == nil {
				return
			}

			//удаляем запись которую пользователь редактировал (она больше не нужна)
			query = `DELETE FROM tpost p 
			    WHERE p.uuid='` + uuid + `' 
				  AND p.uuid_user='` + uuid_user + `' 
				  AND p.edit_type!='publish'`
			b = run_db_exec(d, query)
			if b == nil {
				return
			}

			LogPrint(fmt.Sprintf("publish update record id: %s; user: %s", uuid, uuid_user))

		} else if c1 > 0 && c2 == 0 {
			//если пользователь пытается опубликовать запись но до этого её не редактировал (не сохранял изменения)
			d["error"] = "нет обновленных(сохраненных) данных для публикации записи"
			d["errorcode"] = "nochanges"
			return
		} else if c1 == 0 && c2 > 0 {
			//если пользователь создает новую запись и до этого она не публиковалась или была удалена
			query = `INSERT INTO TPOST(NAME,TAGS,TEXT,PREVIEW,UUID_USER,
				   		DATE_CREATE,date_modify,uuid_user_create,
				   		UUID,UUID_USER_PUBLISH,EDIT_TYPE,date_publish
				    )
				 SELECT NAME,TAGS,TEXT,PREVIEW,UUID_USER,
				   		DATE_CREATE,date_modify,uuid_user_create,
				   		UUID,'` + uuid_user_this + `','publish',current_timestamp
				 FROM tpost 
				 WHERE uuid='` + uuid + `' AND uuid_user='` + uuid_user + `'
				`
			b := run_db_exec(d, query)
			if b == nil {
				return
			}
			//удаляем запись которую пользователь редактировал (она больше не нужна)
			query = `DELETE FROM tpost p 
			    WHERE p.uuid='` + uuid + `' 
				  AND p.uuid_user='` + uuid_user + `' 
				  AND p.edit_type!='publish'
				`
			b = run_db_exec(d, query)
			if b == nil {
				return
			}
			LogPrint(fmt.Sprintf("publish new record id: %s; user: %s", uuid, uuid_user))
		}
	} else if c3 > 0 { //если удаляем запись
		if c1 > 0 {
			//если пост существует то удаляем текущий опубликованный пост

			//обновляем существующую опубликованную запись (именно не удаляем а меняем значения полей - это надо для логов)
			query = `UPDATE tpost p SET 
					p.name =         (SELECT t.name FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.text =         (SELECT t.text FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.tags =         (SELECT t.tags FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.preview =      (SELECT t.preview FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					p.uuid_user =    (SELECT t.uuid_user FROM tpost t WHERE t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'),
					
					p.date_modify   = current_timestamp,
					p.date_publish  = current_timestamp,
					p.uuid_user_publish = '` + uuid_user_this + `',
					p.edit_type = 'delete'
				--FROM tpost p 
				--  LEFT JOIN tpost t ON t.uuid=p.uuid AND t.uuid_user='` + uuid_user + `' AND t.edit_type!='publish'
			    WHERE p.uuid='` + uuid + `' 
				  AND p.edit_type='publish'
				`
			b := run_db_exec(d, query)
			if b == nil {
				return
			}

			//удаляем опубликованную запись
			query = `DELETE FROM tpost p 
			    WHERE p.uuid='` + uuid + `'
				  AND p.uuid_user = '` + uuid_user + `'
				  AND p.uuid_user_publish IS NULL
				  AND p.date_publish IS NULL
				  AND p.edit_type='delete'`
			b = run_db_exec(d, query)
			if b == nil {
				return
			}

			LogPrint(fmt.Sprintf("delete record id: %s; user: %s", uuid, uuid_user))

		} else if c1 == 0 {
			//если пользователь пытается удалить запись которая не публиковалась
			d["error"] = "указанная запись не опубликована (её не удалить из опубликованных записей)"
			d["errorcode"] = "nochanges"
			return
		}
	}
	//LogPrint("=====================================================================")
	//LogPrint(fmt.Sprintf("edit_type: %#v", data["edit_type"]))
	//LogPrint(fmt.Sprintf("%#v", query))
	//LogPrint("=====================================================================")
	//run_db_exec(d, query)

	http_post_edit__load_data_fn(w, r, d, uuid, uuid_user_this)

	return
}
