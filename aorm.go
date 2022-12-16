package aorm

import (
	"database/sql" //只需导入你需要的驱动即可
)

// LinkCommon database/sql提供的库连接与事务，二者有很多方法是一致的，为了通用，抽象为该interface
type LinkCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Executor 查询记录所需要的条件
type Executor struct {
	LinkCommon      LinkCommon
	TableName       string
	SelectList      []string
	SelectExpList   []ExpItem
	GroupList       []string
	WhereList       []WhereItem
	JoinList        []string
	HavingList      []WhereItem
	OrderList       []string
	Offset          int
	PageSize        int
	IsDebug         bool
	IsLockForUpdate bool
	OpinionList     []OpinionItem
}

type ExpItem struct {
	Executor  *Executor
	FieldName string
}

// Use 使用数据库连接，或者事务
func Use(linkCommon LinkCommon) *Executor {
	executor := &Executor{
		LinkCommon: linkCommon,
	}

	return executor
}

//清空查询条件,复用对象
func (db *Executor) clear() {
	db.TableName = ""
	db.SelectList = make([]string, 0)
	db.GroupList = make([]string, 0)
	db.WhereList = make([]WhereItem, 0)
	db.JoinList = make([]string, 0)
	db.HavingList = make([]WhereItem, 0)
	db.OrderList = make([]string, 0)
	db.Offset = 0
	db.PageSize = 0
	db.IsDebug = false
	db.IsLockForUpdate = false
	db.OpinionList = make([]OpinionItem, 0)
}
