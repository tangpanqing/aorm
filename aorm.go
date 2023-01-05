package aorm

import (
	"database/sql" //只需导入你需要的驱动即可
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/migrator"
	"github.com/tangpanqing/aorm/model"
)

// DbContent 数据库连接与数据库类型
type DbContent struct {
	DriverName string
	DbLink     *sql.DB
}

func Store(destList ...interface{}) {
	builder.Store(destList...)
}

//Open 开始一个数据库连接
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

// Use 开始一个数据库操作
func Use(linkCommon model.LinkCommon) *builder.Builder {
	executor := &builder.Builder{
		LinkCommon: linkCommon,
	}

	return executor
}

// Sub 子查询
func Sub() *builder.Builder {
	executor := &builder.Builder{}
	return executor
}

// Migrator 开始一个数据库迁移
func Migrator(linkCommon model.LinkCommon) *migrator.Migrator {
	mi := &migrator.Migrator{
		LinkCommon: linkCommon,
	}
	return mi
}

//清空查询条件,复用对象
//func (ex *builder.Executor) clear() {
//	ex.tableName = ""
//	ex.selectList = make([]string, 0)
//	ex.groupList = make([]string, 0)
//	ex.whereList = make([]builder.WhereItem, 0)
//	ex.joinList = make([]string, 0)
//	ex.havingList = make([]builder.WhereItem, 0)
//	ex.orderList = make([]string, 0)
//	ex.offset = 0
//	ex.pageSize = 0
//	ex.isDebug = false
//	ex.isLockForUpdate = false
//	ex.sql = ""
//	ex.paramList = make([]interface{}, 0)
//	ex.opinionList = make([]OpinionItem, 0)
//}
