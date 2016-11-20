package main

import (
	"database/sql"
	//"strconv"

	_ "github.com/nakagami/firebirdsql"

	//s "strings"

	//mf "github.com/mixamarciv/gofncstd3000"
)

type DBd struct {
	Name     string
	Path     string
	DB       *sql.DB
	NeedAuth bool
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

func InitDb() {
	dbmap = make(map[string]*DBd)

	dbmap["a"] = conn_to_db("admin", apppath+"/db/DB_ADMIN.FDB", false)
	dbmap["p"] = conn_to_db("programming", apppath+"/db/DB_PROGRAMMING.FDB", false)
	dbmap["s"] = conn_to_db("warez", apppath+"/db/DB_WAREZ.FDB", true)
}

func conn_to_db(name, path string, NeedAuth bool) *DBd {
	dbd := new(DBd)
	dbd.Name = name
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
