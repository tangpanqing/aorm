package model

import "database/sql"

type LinkCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type OpinionItem struct {
	Key string
	Val string
}

const Mysql = "mysql"
const Mssql = "mssql"
const Postgres = "postgres"
const Sqlite3 = "sqlite3"
