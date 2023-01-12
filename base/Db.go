package base

import (
	"database/sql"
	"time"
)

type Db struct {
	Driver    string
	DebugMode bool
	SqlDB     *sql.DB
}

//Close 关闭
func (db *Db) Close() error {
	return db.SqlDB.Close()
}

//Begin 开始一个事务
func (db *Db) Begin() *Tx {
	SqlTx, _ := db.SqlDB.Begin()

	return &Tx{
		driver:    db.Driver,
		debugMode: db.DebugMode,

		sqlTx: SqlTx,
	}
}

//SetDebugMode 获取调试模式
func (db *Db) SetDebugMode(debugMode bool) {
	db.DebugMode = debugMode
}

func (db *Db) SetConnMaxLifetime(d time.Duration) {
	db.SqlDB.SetConnMaxLifetime(d)
}

func (db *Db) SetConnMaxIdleTime(d time.Duration) {
	db.SqlDB.SetConnMaxIdleTime(d)
}

func (db *Db) SetMaxIdleConns(n int) {
	db.SqlDB.SetMaxIdleConns(n)
}

func (db *Db) SetMaxOpenConns(n int) {
	db.SqlDB.SetMaxOpenConns(n)
}

func (db *Db) Stats() sql.DBStats {
	return db.SqlDB.Stats()
}

//GetDebugMode 获取调试模式
func (db *Db) GetDebugMode() bool {
	return db.DebugMode
}

func (db *Db) DriverName() string {
	return db.Driver
}

func (db *Db) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.SqlDB.Exec(query, args...)
}

func (db *Db) Prepare(query string) (*sql.Stmt, error) {
	return db.SqlDB.Prepare(query)
}

func (db *Db) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.SqlDB.Query(query, args...)
}

func (db *Db) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.SqlDB.QueryRow(query, args...)
}
