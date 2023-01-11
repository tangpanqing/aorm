package model

import "database/sql"

type AormTx struct {
	driver    string
	debugMode bool
	sqlTx     *sql.Tx
}

//GetDebugMode 获取调试状态
func (tx *AormTx) GetDebugMode() bool {
	return tx.debugMode
}

func (tx *AormTx) DriverName() string {
	return tx.driver
}

func (tx *AormTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.sqlTx.Exec(query, args...)
}

func (tx *AormTx) Prepare(query string) (*sql.Stmt, error) {
	return tx.sqlTx.Prepare(query)
}

func (tx *AormTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.sqlTx.Query(query, args...)
}

func (tx *AormTx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.sqlTx.QueryRow(query, args...)
}

func (tx *AormTx) Rollback() error {
	return tx.sqlTx.Rollback()
}

func (tx *AormTx) Commit() error {
	return tx.sqlTx.Commit()
}
