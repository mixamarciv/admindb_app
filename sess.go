package main

import (
	//"log"
	"net/http"
	//"time"

	//"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	mf "github.com/mixamarciv/gofncstd3000"
)

var sess_store = sessions.NewCookieStore([]byte(gcfg_secret_cookie_key))

func GetSess(w http.ResponseWriter, r *http.Request) *sessions.Session {
	var sess *sessions.Session
	sess, err := sess_store.Get(r, gcfg_secret_cookie_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return sess
}

func SaveSess(s *sessions.Session, w http.ResponseWriter, r *http.Request) {
	s.Save(r, w)
}

func SaveSessAll(w http.ResponseWriter, r *http.Request) {
	sessions.Save(r, w)
}

func GetSessVal(s *sessions.Session, name string, defval interface{}) interface{} {
	val, ok := s.Values[name]
	if ok {
		return val
	}
	val2, ok2 := gcfg_default_session_data[name]
	if ok2 {
		return val2
	}
	return defval
}

func SetSessVal(s *sessions.Session, name string, val interface{}) {
	s.Values[name] = val
}

//выдает данные пользователя в формате json map[string]interface{}
func GetSessUserData(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	sess := GetSess(w, r)
	t, ok := sess.Values["user"]
	if !ok {
		return map[string]interface{}{"error": "no \"user\" data"}
	}
	str, ok := t.(string)
	if !ok {
		return map[string]interface{}{"error": "\"user\" data is not string"}
	}
	if str == "" {
		return map[string]interface{}{"error": "no \"user\" data(user is logout)"}
	}
	j, err := mf.FromJson([]byte(str))
	if err != nil {
		return map[string]interface{}{"error": "bad json string \"user\"! str:\"" + str + "\""}
	}
	return j
}

func SetSessUserData(w http.ResponseWriter, r *http.Request, jsonStr string) {
	sess := GetSess(w, r)
	if jsonStr == "" {
		delete(sess.Values, "user")
	} else {
		sess.Values["user"] = jsonStr
	}
	sess.Save(r, w)
}
