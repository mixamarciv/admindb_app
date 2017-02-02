package main

import (
	"database/sql"
	//"strconv"
	"bytes"

	_ "github.com/nakagami/firebirdsql"

	include_path "path"

	//s "strings"
	"fmt"
	"io"

	"github.com/qiniu/iconv"

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
		return p.String
		return mf.StrTr(p.String, "cp1251", "UTF-8")
	}
	return defaultval
}

func (p *NullString) get_trcp1251_long(defaultval string) string {
	if p.Valid {
		return p.String
		return string(dbStrTr2([]byte(p.String), "cp1251", "UTF-8"))
	}
	return defaultval
}

func dbStrTr2(s []byte, from string, to string) []byte {
	//s2 := make([]byte, len(s))
	// = mf.StrTr(t, from, to)
	//s2 = append(s2, []byte(t)...)
	return s

	cd, err := iconv.Open(to, from) // convert from to to
	if err != nil {
		ret := []byte(fmt.Sprintf("ERROR StrTr2: iconv.Open("+to+","+from+") failed!\n%#v", err))
		return ret
	}
	defer cd.Close()

	input := bytes.NewBuffer(s) // eg. input := os.Stdin || input, err := os.Open(file)
	bufSize := 0                // default if zero
	r := iconv.NewReader(cd, input, bufSize)

	out := bytes.NewBuffer([]byte{})

	_, err = io.Copy(out, r)
	if err != nil {
		ret := []byte(fmt.Sprintf("ERROR StrTr2: io.Copy failed!\n%#v", err))
		return ret
	}

	return out.Bytes()
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
	dbd.Path = include_path.Base(path)
	dbd.NeedAuth = NeedAuth
	//path = "d/program/go/projects/test_martini_app/db/DB1.FDB"
	dbopt := "sysdba:" + gcfg_db_pass + "@127.0.0.1:3050/" + path
	//dbopt += "?code_page=cp1251&cp=cp1251&codepage=cp1251"
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
