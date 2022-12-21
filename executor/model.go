package executor

import "database/sql"

// LinkCommon database/sql提供的库连接与事务，二者有很多方法是一致的，为了通用，抽象为该interface
type LinkCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// ExpItem 将某子语句重命名为某字段
type ExpItem struct {
	Executor  **Executor
	FieldName string
}

// Executor 查询记录所需要的条件
type Executor struct {
	//数据库操作连接
	LinkCommon LinkCommon

	//查询参数
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

	//sql与参数
	sql       string
	paramList []interface{}

	//表属性
	opinionList []OpinionItem

	//驱动名字
	driverName string
}

type OpinionItem struct {
	Key string
	Val string
}
