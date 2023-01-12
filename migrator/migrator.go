package migrator

import (
	"github.com/tangpanqing/aorm/base"
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/driver"
	"github.com/tangpanqing/aorm/migrate_mssql"
	"github.com/tangpanqing/aorm/migrate_mysql"
	"github.com/tangpanqing/aorm/migrate_postgres"
	"github.com/tangpanqing/aorm/migrate_sqlite3"
	"github.com/tangpanqing/aorm/utils"
	"reflect"
	"strings"
)

type Migrator struct {
	//数据库操作连接
	Link base.Link
}

//ShowCreateTable 获取创建表的ddl
func (mi *Migrator) ShowCreateTable(tableName string) string {
	if mi.Link.DriverName() == driver.Mysql {
		me := migrate_mysql.MigrateExecutor{
			Builder: &builder.Builder{
				Link: mi.Link,
			},
		}
		return me.ShowCreateTable(tableName)
	}
	return ""
}

// AutoMigrate 迁移数据库结构,需要输入数据库名,表名自动获取
func (mi *Migrator) AutoMigrate(destList ...interface{}) {
	for i := 0; i < len(destList); i++ {
		dest := destList[i]
		typeOf := reflect.TypeOf(dest)
		valueOf := reflect.ValueOf(dest)
		tableName := getTableNameByReflect(typeOf, valueOf)
		mi.migrateCommon(tableName, typeOf, valueOf)
	}
}

// Migrate 自动迁移数据库结构,需要输入数据库名,表名
func (mi *Migrator) Migrate(tableName string, dest interface{}) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)
	mi.migrateCommon(tableName, typeOf, valueOf)
}

func (mi *Migrator) migrateCommon(tableName string, typeOf reflect.Type, valueOf reflect.Value) {
	if mi.Link.DriverName() == driver.Mssql {
		me := migrate_mssql.MigrateExecutor{
			Builder: &builder.Builder{
				Link: mi.Link,
			},
		}
		me.MigrateCommon(tableName, typeOf)
	}

	if mi.Link.DriverName() == driver.Mysql {
		me := migrate_mysql.MigrateExecutor{
			Builder: &builder.Builder{
				Link: mi.Link,
			},
		}
		me.MigrateCommon(tableName, typeOf, valueOf)
	}

	if mi.Link.DriverName() == driver.Sqlite3 {
		me := migrate_sqlite3.MigrateExecutor{
			Builder: &builder.Builder{
				Link: mi.Link,
			},
		}
		me.MigrateCommon(tableName, typeOf)
	}

	if mi.Link.DriverName() == driver.Postgres {
		me := migrate_postgres.MigrateExecutor{
			Builder: &builder.Builder{
				Link: mi.Link,
			},
		}
		me.MigrateCommon(tableName, typeOf, valueOf)
	}
}

//反射表名,优先从方法获取,没有方法则从名字获取
func getTableNameByReflect(typeOf reflect.Type, valueOf reflect.Value) string {
	method, isSet := typeOf.MethodByName("TableName")
	if isSet {
		var paramList []reflect.Value
		paramList = append(paramList, valueOf)
		res := method.Func.Call(paramList)
		return res[0].String()
	} else {
		arr := strings.Split(typeOf.String(), ".")
		return utils.UnderLine(arr[len(arr)-1])
	}
}
