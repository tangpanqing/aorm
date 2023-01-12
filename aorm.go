package aorm

import (
	"database/sql" //只需导入你需要的驱动即可
	"github.com/tangpanqing/aorm/base"
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/migrator"
)

//Open 开始一个数据库连接
func Open(driverName string, dataSourceName string) (*base.Db, error) {
	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return &base.Db{}, err
	}

	err2 := sqlDB.Ping()
	if err2 != nil {
		return &base.Db{}, err2
	}

	return &base.Db{
		Driver: driverName,
		SqlDB:  sqlDB,
	}, nil
}

func Store(destList ...interface{}) {
	builder.Store(destList...)
}

// Db 开始一个数据库操作
func Db(link base.Link) *builder.Builder {
	b := &builder.Builder{}

	b.Link = link
	b.Debug(link.GetDebugMode())

	return b
}

// Migrator 开始一个数据库迁移
func Migrator(linkCommon base.Link) *migrator.Migrator {
	mi := &migrator.Migrator{
		Link: linkCommon,
	}
	return mi
}
