package base

import "database/sql"

type Tx struct {
	driver    string
	debugMode bool
	sqlTx     *sql.Tx
}

//GetDebugMode 获取调试状态
func (tx *Tx) GetDebugMode() bool {
	return tx.debugMode
}

func (tx *Tx) DriverName() string {
	return tx.driver
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.sqlTx.Exec(query, args...)
}

func (tx *Tx) Prepare(query string) (*sql.Stmt, error) {
	return tx.sqlTx.Prepare(query)
}

func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.sqlTx.Query(query, args...)
}

func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.sqlTx.QueryRow(query, args...)
}

func (tx *Tx) Rollback() error {
	return tx.sqlTx.Rollback()
}

func (tx *Tx) Commit() error {
	return tx.sqlTx.Commit()
}
