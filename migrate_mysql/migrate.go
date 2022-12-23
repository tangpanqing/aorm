package migrate_mysql

import (
	"fmt"
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/model"
	"github.com/tangpanqing/aorm/null"
	"reflect"
	"strconv"
	"strings"
)

type Table struct {
	TableName    null.String
	Engine       null.String
	TableComment null.String
}

type Column struct {
	ColumnName    null.String
	ColumnDefault null.String
	IsNullable    null.String
	DataType      null.String //数据类型 varchar,bigint,int
	MaxLength     null.Int    //数据最大长度 20
	ColumnComment null.String
	Extra         null.String //扩展信息 auto_increment
}

type Index struct {
	NonUnique  null.Int
	ColumnName null.String
	KeyName    null.String
}

//MigrateExecutor 定义结构
type MigrateExecutor struct {
	//驱动名字
	DriverName string

	//表属性
	OpinionList []model.OpinionItem

	//执行者
	Ex *builder.Builder
}

//ShowCreateTable 查看创建表的ddl
func (mm *MigrateExecutor) ShowCreateTable(tableName string) string {
	var str string
	mm.Ex.RawSql("show create table "+tableName).Value("Create Table", &str)
	return str
}

//MigrateCommon 迁移的主要过程
func (mm *MigrateExecutor) MigrateCommon(tableName string, typeOf reflect.Type) error {
	tableFromCode := mm.getTableFromCode(tableName)
	columnsFromCode := mm.getColumnsFromCode(typeOf)
	indexesFromCode := mm.getIndexesFromCode(typeOf, tableFromCode)

	dbName, dbErr := mm.getDbName()
	if dbErr != nil {
		return dbErr
	}

	tablesFromDb := mm.getTableFromDb(dbName, tableName)
	if len(tablesFromDb) != 0 {
		tableFromDb := tablesFromDb[0]
		columnsFromDb := mm.getColumnsFromDb(dbName, tableName)
		indexsFromDb := mm.getIndexesFromDb(tableName)

		mm.modifyTable(tableFromCode, columnsFromCode, indexesFromCode, tableFromDb, columnsFromDb, indexsFromDb)
	} else {
		mm.createTable(tableFromCode, columnsFromCode, indexesFromCode)
	}

	return nil
}

func (mm *MigrateExecutor) getTableFromCode(tableName string) Table {
	var tableFromCode Table
	tableFromCode.TableName = null.StringFrom(tableName)
	tableFromCode.Engine = null.StringFrom(mm.getOpinionVal("ENGINE", "MyISAM"))
	tableFromCode.TableComment = null.StringFrom(mm.getOpinionVal("COMMENT", ""))

	return tableFromCode
}

func (mm *MigrateExecutor) getColumnsFromCode(typeOf reflect.Type) []Column {
	var columnsFromCode []Column
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		fieldName := helper.UnderLine(typeOf.Elem().Field(i).Name)
		fieldType := typeOf.Elem().Field(i).Type.Name()
		fieldMap := getTagMap(typeOf.Elem().Field(i).Tag.Get("aorm"))
		columnsFromCode = append(columnsFromCode, Column{
			ColumnName:    null.StringFrom(fieldName),
			DataType:      null.StringFrom(getDataType(fieldType, fieldMap)),
			MaxLength:     null.IntFrom(int64(getMaxLength(getDataType(fieldType, fieldMap), fieldMap))),
			IsNullable:    null.StringFrom(getNullAble(fieldMap)),
			ColumnComment: null.StringFrom(getComment(fieldMap)),
			Extra:         null.StringFrom(getExtra(fieldMap)),
			ColumnDefault: null.StringFrom(getDefaultVal(fieldMap)),
		})
	}

	return columnsFromCode
}

func (mm *MigrateExecutor) getIndexesFromCode(typeOf reflect.Type, tableFromCode Table) []Index {
	var indexesFromCode []Index
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		fieldName := helper.UnderLine(typeOf.Elem().Field(i).Name)
		fieldMap := getTagMap(typeOf.Elem().Field(i).Tag.Get("aorm"))

		_, primaryIs := fieldMap["primary"]
		if primaryIs {
			indexesFromCode = append(indexesFromCode, Index{
				NonUnique:  null.IntFrom(0),
				ColumnName: null.StringFrom(fieldName),
				KeyName:    null.StringFrom("PRIMARY"),
			})
		}

		_, uniqueIndexIs := fieldMap["unique"]
		if uniqueIndexIs {
			indexesFromCode = append(indexesFromCode, Index{
				NonUnique:  null.IntFrom(0),
				ColumnName: null.StringFrom(fieldName),
				KeyName:    null.StringFrom("idx_" + tableFromCode.TableName.String + "_" + fieldName),
			})
		}

		_, indexIs := fieldMap["index"]
		if indexIs {
			indexesFromCode = append(indexesFromCode, Index{
				NonUnique:  null.IntFrom(1),
				ColumnName: null.StringFrom(fieldName),
				KeyName:    null.StringFrom("idx_" + tableFromCode.TableName.String + "_" + fieldName),
			})
		}
	}

	return indexesFromCode
}

func (mm *MigrateExecutor) getDbName() (string, error) {
	//获取数据库名称
	var dbName string
	err := mm.Ex.RawSql("SELECT DATABASE()").Value("DATABASE()", &dbName)
	if err != nil {
		return "", err
	}

	return dbName, nil
}

func (mm *MigrateExecutor) getTableFromDb(dbName string, tableName string) []Table {
	sql := "SELECT TABLE_NAME,ENGINE,TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA =" + "'" + dbName + "' AND TABLE_NAME =" + "'" + tableName + "'"
	var dataList []Table
	mm.Ex.RawSql(sql).GetMany(&dataList)
	for i := 0; i < len(dataList); i++ {
		dataList[i].TableComment = null.StringFrom("'" + dataList[i].TableComment.String + "'")
	}

	return dataList
}

func (mm *MigrateExecutor) getColumnsFromDb(dbName string, tableName string) []Column {
	var columnsFromDb []Column

	sqlColumn := "SELECT COLUMN_NAME,DATA_TYPE,CHARACTER_MAXIMUM_LENGTH as Max_Length,COLUMN_DEFAULT,COLUMN_COMMENT,EXTRA,IS_NULLABLE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA =" + "'" + dbName + "' AND TABLE_NAME =" + "'" + tableName + "'"
	mm.Ex.RawSql(sqlColumn).GetMany(&columnsFromDb)

	for j := 0; j < len(columnsFromDb); j++ {
		if columnsFromDb[j].DataType.String == "text" && columnsFromDb[j].MaxLength.Int64 == 65535 {
			columnsFromDb[j].MaxLength = null.IntFrom(0)
		}
	}

	return columnsFromDb
}

func (mm *MigrateExecutor) getIndexesFromDb(tableName string) []Index {
	sqlIndex := "SHOW INDEXES FROM " + tableName

	var indexsFromDb []Index
	mm.Ex.RawSql(sqlIndex).GetMany(&indexsFromDb)

	return indexsFromDb
}

func (mm *MigrateExecutor) modifyTable(tableFromCode Table, columnsFromCode []Column, indexesFromCode []Index, tableFromDb Table, columnsFromDb []Column, indexesFromDb []Index) {
	if tableFromCode.Engine != tableFromDb.Engine {
		mm.modifyTableEngine(tableFromCode)
	}

	if tableFromCode.TableComment != tableFromDb.TableComment {
		mm.modifyTableComment(tableFromCode)
	}

	for i := 0; i < len(columnsFromCode); i++ {
		isFind := 0
		columnCode := columnsFromCode[i]

		for j := 0; j < len(columnsFromDb); j++ {
			columnDb := columnsFromDb[j]
			if columnCode.ColumnName == columnDb.ColumnName {
				isFind = 1
				if columnCode.DataType.String != columnDb.DataType.String ||
					columnCode.MaxLength.Int64 != columnDb.MaxLength.Int64 ||
					columnCode.ColumnComment.String != columnDb.ColumnComment.String ||
					columnCode.Extra.String != columnDb.Extra.String ||
					columnCode.ColumnDefault.String != columnDb.ColumnDefault.String {
					sql := "ALTER TABLE " + tableFromCode.TableName.String + " MODIFY " + getColumnStr(columnCode)
					_, err := mm.Ex.Exec(sql)
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("修改属性:" + sql)
					}
				}
			}
		}

		if isFind == 0 {
			sql := "ALTER TABLE " + tableFromCode.TableName.String + " ADD " + getColumnStr(columnCode)
			_, err := mm.Ex.Exec(sql)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("增加属性:" + sql)
			}
		}
	}

	for i := 0; i < len(indexesFromCode); i++ {
		isFind := 0
		indexCode := indexesFromCode[i]

		for j := 0; j < len(indexesFromDb); j++ {
			indexDb := indexesFromDb[j]
			if indexCode.ColumnName == indexDb.ColumnName {
				isFind = 1
				if indexCode.KeyName != indexDb.KeyName || indexCode.NonUnique != indexDb.NonUnique {
					sql := "ALTER TABLE " + tableFromCode.TableName.String + " MODIFY " + getIndexStr(indexCode)
					_, err := mm.Ex.Exec(sql)
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("修改索引:" + sql)
					}
				}
			}
		}

		if isFind == 0 {
			sql := "ALTER TABLE " + tableFromCode.TableName.String + " ADD " + getIndexStr(indexCode)
			_, err := mm.Ex.Exec(sql)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("增加索引:" + sql)
			}
		}
	}
}

func (mm *MigrateExecutor) modifyTableEngine(tableFromCode Table) {
	sql := "ALTER TABLE " + tableFromCode.TableName.String + " Engine " + tableFromCode.Engine.String
	_, err := mm.Ex.Exec(sql)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("修改表:" + sql)
	}
}

func (mm *MigrateExecutor) modifyTableComment(tableFromCode Table) {
	sql := "ALTER TABLE " + tableFromCode.TableName.String + " Comment " + tableFromCode.TableComment.String
	_, err := mm.Ex.Exec(sql)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("修改表:" + sql)
	}
}

func (mm *MigrateExecutor) createTable(tableFromCode Table, columnsFromCode []Column, indexesFromCode []Index) {
	var fieldArr []string

	for i := 0; i < len(columnsFromCode); i++ {
		column := columnsFromCode[i]
		fieldArr = append(fieldArr, getColumnStr(column))
	}

	for i := 0; i < len(indexesFromCode); i++ {
		index := indexesFromCode[i]
		fieldArr = append(fieldArr, getIndexStr(index))
	}

	sqlStr := "CREATE TABLE `" + tableFromCode.TableName.String + "` (\n" + strings.Join(fieldArr, ",\n") + "\n) " + " ENGINE " + tableFromCode.Engine.String + " COMMENT  " + tableFromCode.TableComment.String + ";"
	_, err := mm.Ex.Exec(sqlStr)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("创建表:" + tableFromCode.TableName.String)
	}
}

func (mm *MigrateExecutor) getOpinionVal(key string, def string) string {
	opinions := mm.OpinionList
	for i := 0; i < len(opinions); i++ {
		opinionItem := opinions[i]
		if opinionItem.Key == key {
			def = opinionItem.Val
		}
	}
	return def
}

func getTagMap(fieldTag string) map[string]string {
	var fieldMap = make(map[string]string)
	if "" != fieldTag {
		tagArr := strings.Split(fieldTag, ";")
		for j := 0; j < len(tagArr); j++ {
			tagArrArr := strings.Split(tagArr[j], ":")
			fieldMap[tagArrArr[0]] = ""
			if len(tagArrArr) > 1 {
				fieldMap[tagArrArr[0]] = tagArrArr[1]
			}
		}
	}
	return fieldMap
}

func getColumnStr(column Column) string {
	var strArr []string
	strArr = append(strArr, column.ColumnName.String)
	if column.MaxLength.Int64 == 0 {
		if column.DataType.String == "varchar" {
			strArr = append(strArr, column.DataType.String+"(255)")
		} else {
			strArr = append(strArr, column.DataType.String)
		}
	} else {
		strArr = append(strArr, column.DataType.String+"("+strconv.Itoa(int(column.MaxLength.Int64))+")")
	}

	if column.ColumnDefault.String != "" {
		strArr = append(strArr, "DEFAULT '"+column.ColumnDefault.String+"'")
	}

	if column.IsNullable.String == "NO" {
		strArr = append(strArr, "NOT NULL")
	}

	if column.ColumnComment.String != "" {
		strArr = append(strArr, "COMMENT '"+column.ColumnComment.String+"'")
	}

	if column.Extra.String != "" {
		strArr = append(strArr, column.Extra.String)
	}

	return strings.Join(strArr, " ")
}

func getIndexStr(index Index) string {
	var strArr []string

	if "PRIMARY" == index.KeyName.String {
		strArr = append(strArr, index.KeyName.String)
		strArr = append(strArr, "KEY")
		strArr = append(strArr, "(`"+index.ColumnName.String+"`)")
	} else {
		if 0 == index.NonUnique.Int64 {
			strArr = append(strArr, "Unique")
			strArr = append(strArr, index.KeyName.String)
			strArr = append(strArr, "(`"+index.ColumnName.String+"`)")
		} else {
			strArr = append(strArr, "Index")
			strArr = append(strArr, index.KeyName.String)
			strArr = append(strArr, "(`"+index.ColumnName.String+"`)")
		}
	}

	return strings.Join(strArr, " ")
}

func getDataType(fieldType string, fieldMap map[string]string) string {
	var DataType string

	dataTypeVal, dataTypeOk := fieldMap["type"]
	if dataTypeOk {
		DataType = dataTypeVal
	} else {
		if "Int" == fieldType {
			DataType = "int"
		}
		if "String" == fieldType {
			DataType = "varchar"
		}
		if "Bool" == fieldType {
			DataType = "tinyint"
		}
		if "Time" == fieldType {
			DataType = "datetime"
		}
		if "Float" == fieldType {
			DataType = "float"
		}
	}

	return DataType
}

func getMaxLength(DataType string, fieldMap map[string]string) int {
	var MaxLength int

	maxLengthVal, maxLengthOk := fieldMap["size"]
	if maxLengthOk {
		num, _ := strconv.Atoi(maxLengthVal)
		MaxLength = num
	} else {
		MaxLength = 0
		if "varchar" == DataType {
			MaxLength = 255
		}
	}

	return MaxLength
}

func getNullAble(fieldMap map[string]string) string {
	var IsNullable string

	_, primaryOk := fieldMap["primary"]
	if primaryOk {
		IsNullable = "NO"
	} else {
		_, ok := fieldMap["not null"]
		if ok {
			IsNullable = "NO"
		} else {
			IsNullable = "YES"
		}
	}

	return IsNullable
}

func getComment(fieldMap map[string]string) string {
	commentVal, commentIs := fieldMap["comment"]
	if commentIs {
		return commentVal
	}

	return ""
}

func getExtra(fieldMap map[string]string) string {
	_, commentIs := fieldMap["auto_increment"]
	if commentIs {
		return "auto_increment"
	}

	return ""
}

func getDefaultVal(fieldMap map[string]string) string {
	defaultVal, defaultIs := fieldMap["default"]
	if defaultIs {
		return defaultVal
	}

	return ""
}
