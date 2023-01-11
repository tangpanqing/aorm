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

func (dc *DbContent) Db() *sql.DB {
	return dc.DbLink
}

func (dc *DbContent) Begin() *sql.Tx {
	tx, _ := dc.DbLink.Begin()
	return tx
}

func (dc *DbContent) Exec(query string, args ...interface{}) (sql.Result, error) {
	return dc.Exec(query, args...)
}

func (dc *DbContent) Prepare(query string) (*sql.Stmt, error) {
	return dc.Prepare(query)
}

func (dc *DbContent) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return dc.Query(query, args...)
}

func (dc *DbContent) QueryRow(query string, args ...interface{}) *sql.Row {
	return dc.QueryRow(query, args...)
}

//Open 开始一个数据库连接
func Open(driverName string, dataSourceName string) (*DbContent, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return &DbContent{}, err
	}

	err2 := db.Ping()
	if err2 != nil {
		return &DbContent{}, err2
	}

	return &DbContent{
		DriverName: driverName,
		DbLink:     db,
	}, nil
}

func Store(destList ...interface{}) {
	builder.Store(destList...)
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
