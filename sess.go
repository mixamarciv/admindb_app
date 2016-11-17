package main

import (
	//"log"
	"net/http"
	//"time"

	//"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var sess_store = sessions.NewCookieStore([]byte(gcfg_secret_cookie_key))

func GetSess(w http.ResponseWriter, r *http.Request) *sessions.Session {
	var sess *sessions.Session
	sess, _ = sess_store.Get(r, gcfg_secret_cookie_name)
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
