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
	linkCommon      LinkCommon
	tableName       string
	selectList      []string
	selectExpList   []*ExpItem
	groupList       []string
	whereList       []WhereItem
	joinList        []string
	havingList      []WhereItem
	orderList       []string
	offset          int
	pageSize        int
	isDebug         bool
	isLockForUpdate bool
	opinionList     []OpinionItem
}

type ExpItem struct {
	Executor  **Executor
	FieldName string
}

// Use 使用数据库连接，或者事务
func Use(linkCommon LinkCommon) *Executor {
	executor := &Executor{
		linkCommon: linkCommon,
	}

	return executor
}

// Sub 子查询
func Sub() *Executor {
	executor := &Executor{}
	return executor
}

//清空查询条件,复用对象
func (db *Executor) clear() {
	db.tableName = ""
	db.selectList = make([]string, 0)
	db.groupList = make([]string, 0)
	db.whereList = make([]WhereItem, 0)
	db.joinList = make([]string, 0)
	db.havingList = make([]WhereItem, 0)
	db.orderList = make([]string, 0)
	db.offset = 0
	db.pageSize = 0
	db.isDebug = false
	db.isLockForUpdate = false
	db.opinionList = make([]OpinionItem, 0)
}
