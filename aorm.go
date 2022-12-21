package aorm

import (
	"database/sql" //只需导入你需要的驱动即可
	"github.com/tangpanqing/aorm/executor"
)

// DbContent 数据库连接与数据库类型
type DbContent struct {
	DriverName string
	DbLink     *sql.DB
}

func Open(driverName string, dataSourceName string) (DbContent, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return DbContent{}, err
	}

	return DbContent{
		DriverName: driverName,
		DbLink:     db,
	}, nil
}

// Use 使用数据库连接，或者事务
func Use(linkCommon executor.LinkCommon) *executor.Executor {
	executor := &executor.Executor{
		LinkCommon: linkCommon,
	}

	return executor
}

func UseNew(dbContent DbContent) *executor.Executor {
	executor := &executor.Executor{
		LinkCommon: dbContent.DbLink,
	}

	return executor
}

// Sub 子查询
func Sub() *executor.Executor {
	executor := &executor.Executor{}
	return executor
}

//清空查询条件,复用对象
//func (ex *executor.Executor) clear() {
//	ex.tableName = ""
//	ex.selectList = make([]string, 0)
//	ex.groupList = make([]string, 0)
//	ex.whereList = make([]executor.WhereItem, 0)
//	ex.joinList = make([]string, 0)
//	ex.havingList = make([]executor.WhereItem, 0)
//	ex.orderList = make([]string, 0)
//	ex.offset = 0
//	ex.pageSize = 0
//	ex.isDebug = false
//	ex.isLockForUpdate = false
//	ex.sql = ""
//	ex.paramList = make([]interface{}, 0)
//	ex.opinionList = make([]OpinionItem, 0)
//}
