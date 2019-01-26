package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	//"strconv"

	"strings"

	mf "github.com/mixamarciv/gofncstd3000"

	//"github.com/gorilla/sessions"
)

//авторизация в google апи
func http_auth_google(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{}

	get_vars, _ := url.ParseQuery(r.URL.RawQuery)
	d["url_rawquery"] = r.URL.RawQuery
	d["get_vars"] = get_vars

	code := get_vars.Get("code")
	if len(code) > 0 {
		googleapi := gcfg_app["googleapi"].(map[string]interface{})

		//отправляем апи запрос на получение access_token
		{
			url := "https://www.googleapis.com/oauth2/v4/token"
			//url := "https://www.googleapis.com/oauth2/v2/token"

			str := "code=" + code
			str += "&client_id=" + googleapi["id"].(string)
			str += "&client_secret=" + googleapi["secret"].(string)
			str += "&redirect_uri=http://anykey.vrashke.net/auth_google"
			str += "&grant_type=authorization_code"
			str += ""

			var bstr = []byte(str)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(bstr))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				d["error"] = fmtError("http_auth_google ERROR001: ошибка post запроса", err)
				d["errorcode"] = "autherror"
				RenderError(w, r, d)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				d["error"] = fmtError("http_auth_google ERROR002: ошибка чтения результата post запроса", err)
				d["errorcode"] = "autherror"
				RenderError(w, r, d)
				return
			}

			d["info"] = string(body)

			j := mf.FromJsonStr(body)
			_, b_token := j["access_token"]
			_, b_err := j["error"]
			if !b_token || b_err {
				d["error"] = "http_auth_google ERROR003: ошибка, не верный формат результата post запроса (не задан access_token) \njsondata: \"" + string(body) + "\""
				d["errorcode"] = "autherror"

				RenderError(w, r, d)
				return
			}

			//d["access_token"] = j["id_token"]
			d["access_token"] = j["access_token"]
			d["id_token"] = j["id_token"]
			//d["id_token"] = base64Decode(j["id_token"].(string))
		}

		{ //отправляем запрос на получение данных пользователя
			//urlstr := "https://www.googleapis.com/oauth2/v2/userinfo?fields=family_name%2Cgiven_name%2Cid&key="
			//urlstr := "https://www.googleapis.com/userinfo/v2/me?fields=family_name%2Cgiven_name%2Cid&key="
			//urlstr := "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
			//urlstr := "https://www.googleapis.com/oauth2/v2/userinfo?key="
			//urlstr := "https://www.googleapis.com/plus/v1/people/me?fields=id%2Cname(familyName%2CgivenName)&key="
			urlstr := "https://www.googleapis.com/oauth2/v2/userinfo?fields=family_name%2Cgiven_name%2Cid&key="
			urlstr += d["access_token"].(string)

			d2 := http_auth_google_send_http_request(urlstr, d)
			_, b_err := d2["error"]
			if b_err {
				d["error"] = d2["error"]
				d["errorcode"] = "autherror"
				RenderError(w, r, d)
				return
			}

			d["resp_user_info"] = d2

			u := make(map[string]interface{})
			u["id"] = d2["id"]
			u["first_name"] = d2["given_name"]
			u["last_name"] = d2["family_name"]
			u["access_token"] = d["access_token"]

			d["user_data"] = u
		}

		user_data := d["user_data"].(map[string]interface{})

		//сохраняем авторизацию текущего юзера в бд и получаем его права доступа к бд
		http_auth_google_load_user_data(user_data)

		//если при загрузке и/или авторизации были ошибки то выводим ошибки и отменяем авторизацию
		_, b_err := user_data["error"]
		if b_err {
			RenderTemplate(w, r, user_data, "maintemplate.html", "login.html")
			return
		}

		//все прошло отлично
		//сохраняем данные юзера в текущей сессии
		//sess := GetSess(w, r)
		//sess.Values["user"] = mf.ToJsonStr(user_data)
		//sess.Save(r, w)
		SetSessUserData(w, r, mf.ToJsonStr(user_data))

		d["success"] = "авторизация пользователя " + user_data["name"].(string) + " успешно пройдена "
		RenderTemplate(w, r, d, "maintemplate.html", "login.html")
		return
	}
	RenderTemplate(w, r, d, "maintemplate.html", "login.html")
	return
}

//map[uid:1.42080324e+08 first_name:Михаил last_name:Марцив access_token:6b...9a9 expires_in:86400 user_id:1.42080324e+08]

//загружаем данные пользователя из бд или задаем значения по умолчанию и сохраняем нового юзера в бд
func http_auth_google_load_user_data(d map[string]interface{}) {

	//d["id"] = floatToStr(d["user_id"])
	name := d["first_name"].(string) + " " + d["last_name"].(string)
	default_fdata := `{"accessdb":{"a":"1","p":"1","w":"0"},"first_name":"` + d["first_name"].(string) + `","last_name":"` + d["last_name"].(string) + `"}`
	//права по умолчанию 0-нет доступа,
	//1-только чтение, 2-запись, 3-модерация чужих записей(подтверждение и пубдикация записей)

	db := dbmap["users"].DB

	//получаем текущие данны пользователя в бд
	query := "SELECT uuid,name,fdata FROM tuser WHERE id='" + d["id"].(string) + "' AND type='google'"
	rows, err := db.Query(query)
	if err != nil {
		d["error"] = fmtError("http_auth_google_load_user_data ERROR001 db.Query(query): query:\n"+query+"\n\n", err)
		d["errorcode"] = "dbqueryerror"
		return
	}

	d["uuid_user"] = ""
	for rows.Next() {
		var uuid_user, name, fdata NullString
		if err := rows.Scan(&uuid_user, &name, &fdata); err != nil {
			d["error"] = fmtError("http_auth_google_load_user_data ERROR002 rows.Scan: query:\n"+query+"\n\n", err)
			d["errorcode"] = "dbqueryerror"
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
			",'" + name + "','" + fdata + "','google')"
		_, err := db.Exec(query)
		if err != nil {
			d["error"] = fmtError("http_auth_google_load_user_data ERROR003 db.Exec(query): query:\n"+query+"\n\n", err)
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
		query := "UPDATE tuser SET fdata='" + fdata + "',name='" + name + "' WHERE uuid='" + d["uuid_user"].(string) + "'"
		_, err := db.Exec(query)
		if err != nil {
			d["error"] = fmtError("http_auth_google_load_user_data ERROR004 db.Exec(query): query:\n"+query+"\n\n", err)
			return
		}
	}

	{ //и в любом случае регистрируем его регистрацию в системе с теми правами которые он получает при регистрации
		d["uuid_auth"] = mf.StrUuid()
		fdata := strings.Replace(d["fdata"].(string), "'", "''", -1)

		query := "INSERT INTO TUSER_AUTH_GOOGLE(uuid,uuid_tuser,fdata,access_token) VALUES('" + d["uuid_auth"].(string) + "'" +
			",'" + d["uuid_user"].(string) + "','" + fdata + "','" + d["access_token"].(string) + "')"
		_, err := db.Exec(query)
		if err != nil {
			d["error"] = fmtError("http_auth_google_load_user_data ERROR005 db.Exec(query): query:\n"+query+"\n\n", err)
			return
		}
	}

	d["fdata"] = mf.FromJsonStr([]byte(d["fdata"].(string)))
	d["type"] = "google"

	//удаляем лишние данные, все необходимое уже сохранено в бд, далее при внесении изменений сверяем данные сессии с бд
	delete(d, "first_name")
	delete(d, "last_name")
	delete(d, "access_token")

	return
}

//отправляем запрос
func http_auth_google_send_http_request(urlStr string, d map[string]interface{}) map[string]interface{} {

	LogPrint("http_auth_google_send_http_request urlStr: " + urlStr + "\n")

	//data := url.Values{}
	ret := make(map[string]interface{})
	client := &http.Client{}

	r, err := http.NewRequest("GET", urlStr, nil)
	LogPrintErrAndExit("http_auth_google_send_http_request error001: \n urlStr: "+urlStr+"\n\n", err)

	r.Header.Add("Authorization", "Bearer "+d["access_token"].(string))
	//r.Header.Add("path", "/")
	//r.Header.Add("scheme", "https")
	//r.Header.Add("accept", "text/json")

	resp, err := client.Do(r)
	if err != nil {
		ret["error"] = fmt.Sprintf("http_auth_google_send_http_request ERROR002 client.Do: %#v", err)
		return ret
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ret["error"] = fmt.Sprintf("http_auth_google_send_http_request ERROR003 ioutil.ReadAll(resp.Body): %#v", err)
		return ret
	}

	ret, err = mf.FromJson(body)
	if err != nil {
		ret["errorcode"] = "hz"
		ret["error"] = fmt.Sprintf("http_auth_google_send_http_request ERROR004 Unmarshal json error: %#v", err) + "\njsondata: \"" + string(body) + "\"\nurlStr: \"" + urlStr + "\"\n"
		return ret
	}
	return ret
}
