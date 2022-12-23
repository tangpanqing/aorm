package migrator

import (
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/migrate_mssql"
	"github.com/tangpanqing/aorm/migrate_mysql"
	"github.com/tangpanqing/aorm/migrate_postgres"
	"github.com/tangpanqing/aorm/migrate_sqlite3"
	"github.com/tangpanqing/aorm/model"
	"reflect"
	"strings"
)

type Migrator struct {
	//数据库操作连接
	LinkCommon model.LinkCommon

	//驱动名字
	driverName string

	//表属性
	opinionList []model.OpinionItem
}

func (mi *Migrator) Driver(driverName string) *Migrator {
	mi.driverName = driverName
	return mi
}

func (mi *Migrator) Opinion(key string, val string) *Migrator {
	if key == "COMMENT" {
		val = "'" + val + "'"
	}

	mi.opinionList = append(mi.opinionList, model.OpinionItem{Key: key, Val: val})

	return mi
}

//ShowCreateTable 获取创建表的ddl
func (mi *Migrator) ShowCreateTable(tableName string) string {
	if mi.driverName == "mysql" {
		me := migrate_mysql.MigrateExecutor{
			DriverName:  mi.driverName,
			OpinionList: mi.opinionList,
			Ex: &builder.Builder{
				LinkCommon: mi.LinkCommon,
			},
		}
		return me.ShowCreateTable(tableName)
	}
	return ""
}

// AutoMigrate 迁移数据库结构,需要输入数据库名,表名自动获取
func (mi *Migrator) AutoMigrate(dest interface{}) {
	typeOf := reflect.TypeOf(dest)
	arr := strings.Split(typeOf.String(), ".")
	tableName := helper.UnderLine(arr[len(arr)-1])

	mi.migrateCommon(tableName, typeOf)
}

// Migrate 自动迁移数据库结构,需要输入数据库名,表名
func (mi *Migrator) Migrate(tableName string, dest interface{}) {
	typeOf := reflect.TypeOf(dest)
	mi.migrateCommon(tableName, typeOf)
}

func (mi *Migrator) migrateCommon(tableName string, typeOf reflect.Type) {
	if mi.driverName == "mssql" {
		me := migrate_mssql.MigrateExecutor{
			DriverName:  mi.driverName,
			OpinionList: mi.opinionList,
			Ex: &builder.Builder{
				LinkCommon: mi.LinkCommon,
			},
		}
		me.MigrateCommon(tableName, typeOf)
	}

	if mi.driverName == "mysql" {
		me := migrate_mysql.MigrateExecutor{
			DriverName:  mi.driverName,
			OpinionList: mi.opinionList,
			Ex: &builder.Builder{
				LinkCommon: mi.LinkCommon,
			},
		}
		me.MigrateCommon(tableName, typeOf)
	}

	if mi.driverName == "sqlite3" {
		me := migrate_sqlite3.MigrateExecutor{
			DriverName:  mi.driverName,
			OpinionList: mi.opinionList,
			Ex: &builder.Builder{
				LinkCommon: mi.LinkCommon,
			},
		}
		me.MigrateCommon(tableName, typeOf)
	}

	if mi.driverName == "postgres" {
		me := migrate_postgres.MigrateExecutor{
			DriverName:  mi.driverName,
			OpinionList: mi.opinionList,
			Ex: &builder.Builder{
				LinkCommon: mi.LinkCommon,
			},
		}
		me.MigrateCommon(tableName, typeOf)
	}
}

func (mi *Migrator) GetOpinionList() []model.OpinionItem {
	return mi.opinionList
}
