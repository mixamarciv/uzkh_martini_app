package main

import (
	"database/sql"
	//"strconv"

	_ "github.com/nakagami/firebirdsql"

	s "strings"

	mf "github.com/mixamarciv/gofncstd3000"
)

var db *sql.DB

func InitDb() {
	path, _ := mf.AppPath()
	path = s.Replace(path, "\\", "/", -1) + "/db/DB1.FDB"
	//path = "d/program/go/projects/test_martini_app/db/DB1.FDB"
	dbopt := "sysdba:masterkey@127.0.0.1:3050/" + path
	var err error
	db, err = sql.Open("firebirdsql", dbopt)
	LogPrintErrAndExit("ошибка подключения к базе данных "+dbopt, err)
	LogPrint("установлено подключение к БД: " + dbopt)

	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(100)

	query := `SELECT CAST(COUNT(*) AS VARCHAR(100)) FROM tpost `
	rows, err := db.Query(query)
	LogPrintErrAndExit("db.Query error: \n"+query+"\n\n", err)
	rows.Next()
	var cnt string
	err = rows.Scan(&cnt)
	LogPrintErrAndExit("rows.Scan error: \n"+query+"\n\n", err)
	LogPrint("всего записей в БД: " + cnt)
}
