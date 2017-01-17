package main

import (
	"database/sql"
	//"strconv"

	_ "github.com/nakagami/firebirdsql"

	//s "strings"

	mf "github.com/mixamarciv/gofncstd3000"
)

type DBd struct {
	Name      string
	ShortName string
	Path      string
	DB        *sql.DB
	NeedAuth  bool
}

var dbmap map[string]*DBd

type NullString struct {
	sql.NullString
}

func (p *NullString) get(defaultval string) string {
	if p.Valid {
		return p.String
	}
	return defaultval
}

func (p *NullString) get_trcp1251(defaultval string) string {
	if p.Valid {
		return mf.StrTr(p.String, "cp1251", "UTF-8")
	}
	return defaultval
}

func (p *NullString) get_trcp1251_long(defaultval string) string {
	if p.Valid {
		return string(StrTr2([]byte(p.String), "cp1251", "UTF-8"))
	}
	return defaultval
}

func StrTr2(s []byte, from string, to string) []byte {
	cnt := 200
	cntlen := len(s) % cnt
	var s2 []byte
	for i := 0; i < len(s)-cntlen; i += cnt {
		t := string(s[i : i+cnt])
		t = mf.StrTr(t, from, to)
		s2 = append(s2, []byte(t)...)
	}
	t := string(s[len(s)-cntlen : len(s)])
	t = mf.StrTr(t, from, to)
	s2 = append(s2, []byte(t)...)
	return s2
}

func InitDb() {
	dbmap = make(map[string]*DBd)

	dbmap["users"] = conn_to_db("users", "users", apppath+"/db/DB_USERS.FDB", false)
	dbmap["a"] = conn_to_db("a", "admin", apppath+"/db/DB_ADMIN.FDB", false)
	dbmap["p"] = conn_to_db("p", "programming", apppath+"/db/DB_PROGRAMMING.FDB", false)
	dbmap["w"] = conn_to_db("w", "warez", apppath+"/db/DB_WAREZ.FDB", true)
}

func conn_to_db(shortName, name, path string, NeedAuth bool) *DBd {
	dbd := new(DBd)
	dbd.Name = name
	dbd.ShortName = shortName
	dbd.Path = path
	dbd.NeedAuth = NeedAuth
	//path = "d/program/go/projects/test_martini_app/db/DB1.FDB"
	dbopt := "sysdba:" + gcfg_db_pass + "@127.0.0.1:3050/" + path
	var err error
	db, err := sql.Open("firebirdsql", dbopt)
	LogPrintErrAndExit("ошибка подключения к базе данных "+dbopt, err)
	LogPrint("установлено подключение к БД: " + dbopt)

	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(100)

	query := `SELECT COUNT(*)-1 FROM rdb$database`
	rows, err := db.Query(query)
	LogPrintErrAndExit("db.Query error: \n"+query+"\n\n", err)
	rows.Next()
	var cnt string
	err = rows.Scan(&cnt)
	LogPrintErrAndExit("rows.Scan error: \n"+query+"\n\n", err)
	LogPrint("всего записей в БД: " + cnt)

	dbd.DB = db
	return dbd
}
