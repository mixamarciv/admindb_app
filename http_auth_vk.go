package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	//"strconv"

	"strings"

	mf "github.com/mixamarciv/gofncstd3000"

	//"github.com/gorilla/sessions"
)

//авторизация в вк апи
func http_auth_vk(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	get_vars, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["get_vars"] = get_vars

	code := get_vars.Get("code")
	if len(code) > 0 {
		vkapi := gcfg_app["vkapi"].(map[string]interface{})

		//отправляем апи запрос на получение access_token
		urlstr := "https://oauth.vk.com/access_token?"
		urlstr += "client_id=" + vkapi["id"].(string) + "&client_secret=" + vkapi["secret"].(string)
		urlstr += "&code=" + code + "&redirect_uri=http://" + r.Host + "/auth_vk"
		d2 := http_auth_vk_send_http_request(urlstr)

		_, err1 := d2["err"]
		_, err2 := d2["error"]
		if err1 || err2 {
			RenderTemplate(w, r, d2, "maintemplate.html", "login.html")
			return
		}

		//отправляем запрос на получение данных пользователя
		urlstr = "https://api.vk.com/method/users.get?uid=" + floatToStr(d2["user_id"])
		urlstr += "&access_token=" + d2["access_token"].(string)
		d3 := http_auth_vk_send_http_request(urlstr)
		_, err1 = d3["err"]
		_, err2 = d3["error"]
		if err1 || err2 {
			RenderTemplate(w, r, d3, "maintemplate.html", "login.html")
			return
		}

		//проверяем результаты:
		resp, ok := d3["response"]
		if !ok {
			d3["error"] = fmt.Errorf("http_auth_vk ERROR001: no response")
			RenderTemplate(w, r, d3, "maintemplate.html", "login.html")
			return
		}
		tresp := fmt.Sprintf("%T", resp)
		if tresp != "[]interface {}" || fmt.Sprintf("%T", resp.([]interface{})[0]) != "map[string]interface {}" {
			d3["error"] = fmt.Errorf("http_auth_vk ERROR002: bad response")
			RenderTemplate(w, r, d3, "maintemplate.html", "login.html")
			return
		}

		//собираем все в один map
		for k, v := range resp.([]interface{})[0].(map[string]interface{}) {
			d2[k] = v
		}

		//сохраняем авторизацию текущего юзера в бд и получаем его права доступа к бд
		http_auth_vk_load_user_data(d2)
		user_data := d2

		//если при загрузке и/или авторизации были ошибки то выводи ошибки и отменяем авторизацию
		_, err1 = user_data["error"]
		if err1 {
			RenderTemplate(w, r, user_data, "maintemplate.html", "login.html")
			return
		}

		//все прошло отлично
		//сохраняем данные юзера в текущей сессии
		//sess := GetSess(w, r)
		//sess.Values["user"] = mf.ToJsonStr(user_data)
		//sess.Save(r, w)
		SetSessUserData(w, r, mf.ToJsonStr(user_data))

		//{
		//	sess, _ := sess_store.Get(r, gcfg_secret_cookie_name)
		//	sess.Values["user"] = mf.ToJsonStr(user_data)
		//	sess.Save(r, w)
		//	sessions.Save(r, w)
		//}
		//**********************************************************

		d["success"] = "авторизация пользователя " + user_data["name"].(string) + " успешно пройдена "

		RenderTemplate(w, r, d, "maintemplate.html", "login.html")
		return
	}
	RenderTemplate(w, r, d, "maintemplate.html", "login.html")
	return
}

//map[uid:1.42080324e+08 first_name:Михаил last_name:Марцив access_token:6b...9a9 expires_in:86400 user_id:1.42080324e+08]

//загружаем данные пользователя из бд или задаем значения по умолчанию и сохраняем нового юзера в бд
func http_auth_vk_load_user_data(d map[string]interface{}) {
	_, err1 := d["err"]
	_, err2 := d["error"]
	if err1 || err2 {
		return
	}

	d["id"] = floatToStr(d["user_id"])
	name := d["first_name"].(string) + " " + d["last_name"].(string)
	default_fdata := `{"accessdb":{"a":"1","p":"1","w":"0"},"first_name":"` + d["first_name"].(string) + `","last_name":"` + d["last_name"].(string) + `"}`
	//права по умолчанию 0-нет доступа,
	//1-только чтение, 2-запись, 3-модерация чужих записей(подтверждение и пубдикация записей)

	db := dbmap["users"].DB

	//получаем текущие данны пользователя в бд
	query := "SELECT uuid,name,fdata FROM tuser WHERE id=" + d["id"].(string)
	rows, err := db.Query(query)
	if err != nil {
		d["error"] = fmtError("http_auth_vk_load_user_data ERROR001 db.Query(query): query:\n"+query+"\n\n", err)
		return
	}

	d["uuid_user"] = ""
	for rows.Next() {
		var uuid_user, name, fdata NullString
		if err := rows.Scan(&uuid_user, &name, &fdata); err != nil {
			d["error"] = fmtError("http_auth_vk_load_user_data ERROR002 rows.Scan: query:\n"+query+"\n\n", err)
			return
		}
		d["uuid_user"] = uuid_user.get("")
		d["name"] = name.get("")
		//d["fdata"] = mf.FromJsonStr([]byte(fdata.get(default_fdata)))
		d["fdata"] = fdata.get(default_fdata)
	}
	rows.Close()

	//если пользователя ещё нет в системе то регистрируем его в бд
	if d["uuid_user"].(string) == "" {
		d["uuid_user"] = mf.StrUuid()
		d["name"] = name
		d["fdata"] = default_fdata

		name := strings.Replace(d["name"].(string), "'", "''", -1)
		fdata := strings.Replace(d["fdata"].(string), "'", "''", -1)

		query := "INSERT INTO tuser(uuid,id,name,fdata,type) VALUES('" + d["uuid_user"].(string) + "','" + d["id"].(string) + "'" +
			",'" + name + "','" + fdata + "','vk')"
		_, err := db.Exec(query)
		if err != nil {
			d["error"] = fmtError("http_auth_vk_load_user_data ERROR003 db.Exec(query): query:\n"+query+"\n\n", err)
			return
		}
	}

	//если имя пользователя изменилось
	if d["name"].(string) != name {
		t := mf.FromJsonStr([]byte(d["fdata"].(string)))
		t["last_name"] = d["last_name"].(string)
		t["first_name"] = d["first_name"].(string)
		d["fdata"] = mf.ToJsonStr(t)
		name := strings.Replace(name, "'", "''", -1)
		fdata := strings.Replace(d["fdata"].(string), "'", "''", -1)
		query := "UPDATE tuser SET fdata='" + fdata + "',name='" + name + "' WHERE uuid='" + d["uuid_user"].(string) + "' AND type='vk'"
		_, err := db.Exec(query)
		if err != nil {
			d["error"] = fmtError("http_auth_vk_load_user_data ERROR004 db.Exec(query): query:\n"+query+"\n\n", err)
			return
		}
	}

	{ //и в любом случае регистрируем его регистрацию в системе с теми правами которые он получает при регистрации
		d["uuid_auth"] = mf.StrUuid()
		fdata := strings.Replace(d["fdata"].(string), "'", "''", -1)

		query := "INSERT INTO TUSER_AUTH_VK(uuid,uuid_tuser,fdata,access_token) VALUES('" + d["uuid_auth"].(string) + "'" +
			",'" + d["uuid_user"].(string) + "','" + fdata + "','" + d["access_token"].(string) + "')"
		_, err := db.Exec(query)
		if err != nil {
			d["error"] = fmtError("http_auth_vk_load_user_data ERROR005 db.Exec(query): query:\n"+query+"\n\n", err)
			return
		}
	}

	d["fdata"] = mf.FromJsonStr([]byte(d["fdata"].(string)))
	d["type"] = "vk"

	//удаляем лишние данные, все необходимое уже сохранено в бд, далее при внесении изменений сверяем данные сессии с бд
	delete(d, "access_token")
	delete(d, "first_name")
	delete(d, "last_name")
	delete(d, "expires_in")
	delete(d, "user_id")
	delete(d, "uid")

	return
}

//отправляем запрос
func http_auth_vk_send_http_request(urlStr string) map[string]interface{} {

	LogPrint("http_auth_vk_send_http_request urlStr: " + urlStr + "\n")

	//data := url.Values{}
	ret := make(map[string]interface{})
	client := &http.Client{}

	r, err := http.NewRequest("GET", urlStr, nil)
	LogPrintErrAndExit("http_auth_vk_send_http_request error001: \n urlStr: "+urlStr+"\n\n", err)

	r.Header.Add("method", "GET")
	r.Header.Add("path", "/")
	r.Header.Add("scheme", "https")
	r.Header.Add("accept", "text/json")

	resp, err := client.Do(r)
	if err != nil {
		ret["err"] = fmt.Sprintf("http_auth_vk_send_http_request ERROR002 client.Do: %#v", err)
		return ret
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ret["err"] = fmt.Sprintf("http_auth_vk_send_http_request ERROR003 ioutil.ReadAll(resp.Body): %#v", err)
		return ret
	}

	ret, err = mf.FromJson(body)
	if err != nil {
		ret["err"] = fmt.Sprintf("http_auth_vk_send_http_request ERROR004 Unmarshal json error: %#v", err)
		return ret
	}
	return ret
}
