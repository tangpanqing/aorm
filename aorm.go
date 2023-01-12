package aorm

import (
	"database/sql" //只需导入你需要的驱动即可
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/migrator"
	"github.com/tangpanqing/aorm/model"
)

//Open 开始一个数据库连接
func Open(driverName string, dataSourceName string) (*model.AormDB, error) {
	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return &model.AormDB{}, err
	}

	err2 := sqlDB.Ping()
	if err2 != nil {
		return &model.AormDB{}, err2
	}

	return &model.AormDB{
		Driver: driverName,
		SqlDB:  sqlDB,
	}, nil
}

func Store(destList ...interface{}) {
	builder.Store(destList...)
}

// Db 开始一个数据库操作
func Db(link model.AormLink) *builder.Builder {
	b := &builder.Builder{}

	b.Link = link
	b.Debug(link.GetDebugMode())

	return b
}

// Migrator 开始一个数据库迁移
func Migrator(linkCommon model.AormLink) *migrator.Migrator {
	mi := &migrator.Migrator{
		Link: linkCommon,
	}
	return mi
}
