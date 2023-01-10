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

// Db 开始一个数据库操作
func Db(linkCommon ...model.LinkCommon) *builder.Builder {
	b := &builder.Builder{}

	if len(linkCommon) > 0 {
		b.LinkCommon = linkCommon[0]
	}

	return b
}

// Migrator 开始一个数据库迁移
func Migrator(linkCommon model.LinkCommon) *migrator.Migrator {
	mi := &migrator.Migrator{
		LinkCommon: linkCommon,
	}
	return mi
}
