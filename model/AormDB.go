package model

import "database/sql"

// AormDB 数据库连接与数据库类型
type AormDB struct {
	Driver    string
	DebugMode bool
	SqlDB     *sql.DB
}

//Begin 开始一个事务
func (db *AormDB) Begin() *AormTx {
	SqlTx, _ := db.SqlDB.Begin()

	return &AormTx{
		driver:    db.Driver,
		debugMode: db.DebugMode,

		sqlTx: SqlTx,
	}
}

//SetDebugMode 获取调试模式
func (db *AormDB) SetDebugMode(debugMode bool) {
	db.DebugMode = debugMode
}

//GetDebugMode 获取调试模式
func (db *AormDB) GetDebugMode() bool {
	return db.DebugMode
}

func (db *AormDB) DriverName() string {
	return db.Driver
}

func (db *AormDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.SqlDB.Exec(query, args...)
}

func (db *AormDB) Prepare(query string) (*sql.Stmt, error) {
	return db.SqlDB.Prepare(query)
}

func (db *AormDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.SqlDB.Query(query, args...)
}

func (db *AormDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.SqlDB.QueryRow(query, args...)
}
