package aorm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Table struct {
	TableName    String
	Engine       String
	TableComment String
}

type Column struct {
	ColumnName    String
	ColumnDefault String
	IsNullable    String
	DataType      String //数据类型 varchar,bigint,int
	MaxLength     Int    //数据最大长度 20
	ColumnComment String
	Extra         String //扩展信息 auto_increment
}

type Index struct {
	NonUnique  int
	ColumnName string
	KeyName    string
}

type OpinionItem struct {
	Key string
	Val string
}

func (db *Executor) Opinion(key string, val string) *Executor {
	if key == "COMMENT" {
		val = "'" + val + "'"
	}

	db.opinionList = append(db.opinionList, OpinionItem{Key: key, Val: val})

	return db
}

func (db *Executor) ShowCreateTable(tableName string) string {
	var str string
	db.RawSql("show create table "+tableName).Value("Create Table", &str)
	return str
}

// AutoMigrate 迁移数据库结构,需要输入数据库名,表名自动获取
func (db *Executor) AutoMigrate(dest interface{}) {
	typeOf := reflect.TypeOf(dest)
	arr := strings.Split(typeOf.String(), ".")
	tableName := UnderLine(arr[len(arr)-1])

	db.migrateCommon(tableName, typeOf)
}

// Migrate 自动迁移数据库结构,需要输入数据库名,表名
func (db *Executor) Migrate(tableName string, dest interface{}) {
	typeOf := reflect.TypeOf(dest)
	db.migrateCommon(tableName, typeOf)
}

func (db *Executor) migrateCommon(tableName string, typeOf reflect.Type) error {
	tableFromCode := db.getTableFromCode(tableName)
	columnsFromCode := db.getColumnsFromCode(typeOf)
	indexsFromCode := db.getIndexsFromCode(typeOf, tableFromCode)

	//获取数据库名称
	var dbName string
	err := db.RawSql("SELECT DATABASE()").Value("DATABASE()", &dbName)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//查询表信息,如果找不到就新建
	sql := "SELECT TABLE_NAME,ENGINE,TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA =" + "'" + dbName + "' AND TABLE_NAME =" + "'" + tableName + "'"
	var dataList []Table
	db.RawSql(sql).GetMany(&dataList)
	for i := 0; i < len(dataList); i++ {
		dataList[i].TableComment = StringFrom("'" + dataList[i].TableComment.String + "'")
	}

	if len(dataList) != 0 {
		tableFromDb := dataList[0]
		columnsFromDb := db.getColumnsFromDb(dbName, tableName)
		indexsFromDb := db.getIndexsFromDb(tableName)

		db.modifyTable(tableFromCode, columnsFromCode, indexsFromCode, tableFromDb, columnsFromDb, indexsFromDb)
	} else {
		db.createTable(tableFromCode, columnsFromCode, indexsFromCode)
	}

	return nil
}

func (db *Executor) getTableFromCode(tableName string) Table {
	var tableFromCode Table
	tableFromCode.TableName = StringFrom(tableName)
	tableFromCode.Engine = StringFrom(db.getValFromOpinion("ENGINE", "MyISAM"))
	tableFromCode.TableComment = StringFrom(db.getValFromOpinion("COMMENT", ""))

	return tableFromCode
}

func (db *Executor) getColumnsFromCode(typeOf reflect.Type) []Column {
	var columnsFromCode []Column
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		fieldName := UnderLine(typeOf.Elem().Field(i).Name)
		fieldType := typeOf.Elem().Field(i).Type.Name()
		fieldMap := getTagMap(typeOf.Elem().Field(i).Tag.Get("aorm"))
		columnsFromCode = append(columnsFromCode, getColumnFromCode(fieldName, fieldType, fieldMap))
	}

	return columnsFromCode
}

func (db *Executor) getIndexsFromCode(typeOf reflect.Type, tableFromCode Table) []Index {
	var indexsFromCode []Index
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		fieldName := UnderLine(typeOf.Elem().Field(i).Name)
		fieldMap := getTagMap(typeOf.Elem().Field(i).Tag.Get("aorm"))

		_, primaryIs := fieldMap["primary"]
		if primaryIs {
			indexsFromCode = append(indexsFromCode, Index{
				NonUnique:  0,
				ColumnName: fieldName,
				KeyName:    "PRIMARY",
			})
		}

		_, uniqueIndexIs := fieldMap["unique"]
		if uniqueIndexIs {
			indexsFromCode = append(indexsFromCode, Index{
				NonUnique:  0,
				ColumnName: fieldName,
				KeyName:    "idx_" + tableFromCode.TableName.String + "_" + fieldName,
			})
		}

		_, indexIs := fieldMap["index"]
		if indexIs {
			indexsFromCode = append(indexsFromCode, Index{
				NonUnique:  1,
				ColumnName: fieldName,
				KeyName:    "idx_" + tableFromCode.TableName.String + "_" + fieldName,
			})
		}
	}

	return indexsFromCode
}

func (db *Executor) getColumnsFromDb(dbName string, tableName string) []Column {
	var columnsFromDb []Column

	sqlColumn := "SELECT COLUMN_NAME,DATA_TYPE,CHARACTER_MAXIMUM_LENGTH as Max_Length,COLUMN_DEFAULT,COLUMN_COMMENT,EXTRA,IS_NULLABLE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA =" + "'" + dbName + "' AND TABLE_NAME =" + "'" + tableName + "'"
	db.RawSql(sqlColumn).GetMany(&columnsFromDb)

	for j := 0; j < len(columnsFromDb); j++ {
		if columnsFromDb[j].DataType.String == "text" && columnsFromDb[j].MaxLength.Int64 == 65535 {
			columnsFromDb[j].MaxLength = IntFrom(0)
		}
	}

	return columnsFromDb
}

func (db *Executor) getIndexsFromDb(tableName string) []Index {
	sqlIndex := "SHOW INDEXES FROM " + tableName

	var indexsFromDb []Index
	db.RawSql(sqlIndex).GetMany(&indexsFromDb)

	return indexsFromDb
}

// 修改表
func (db *Executor) modifyTable(tableFromCode Table, columnsFromCode []Column, indexsFromCode []Index, tableFromDb Table, columnsFromDb []Column, indexsFromDb []Index) {
	if tableFromCode.Engine != tableFromDb.Engine {
		sql := "ALTER TABLE " + tableFromCode.TableName.String + " Engine " + tableFromCode.Engine.String
		_, err := db.Exec(sql)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("修改表:" + sql)
		}
	}

	if tableFromCode.TableComment != tableFromDb.TableComment {
		sql := "ALTER TABLE " + tableFromCode.TableName.String + " Comment " + tableFromCode.TableComment.String
		_, err := db.Exec(sql)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("修改表:" + sql)
		}
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
					_, err := db.Exec(sql)
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
			_, err := db.Exec(sql)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("增加属性:" + sql)
			}
		}
	}

	for i := 0; i < len(indexsFromCode); i++ {
		isFind := 0
		indexCode := indexsFromCode[i]

		for j := 0; j < len(indexsFromDb); j++ {
			indexDb := indexsFromDb[j]
			if indexCode.ColumnName == indexDb.ColumnName {
				isFind = 1
				if indexCode.KeyName != indexDb.KeyName || indexCode.NonUnique != indexDb.NonUnique {
					sql := "ALTER TABLE " + tableFromCode.TableName.String + " MODIFY " + getIndexStr(indexCode)
					_, err := db.Exec(sql)
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
			_, err := db.Exec(sql)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("增加索引:" + sql)
			}
		}
	}
}

// 创建表
func (db *Executor) createTable(tableFromCode Table, columnsFromCode []Column, indexsFromCode []Index) {
	var fieldArr []string

	for i := 0; i < len(columnsFromCode); i++ {
		column := columnsFromCode[i]
		fieldArr = append(fieldArr, getColumnStr(column))
	}

	for i := 0; i < len(indexsFromCode); i++ {
		index := indexsFromCode[i]
		fieldArr = append(fieldArr, getIndexStr(index))
	}

	sqlStr := "CREATE TABLE `" + tableFromCode.TableName.String + "` (\n" + strings.Join(fieldArr, ",\n") + "\n) " + getTableInfoFromCode(tableFromCode) + ";"
	_, err := db.Exec(sqlStr)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("创建表:" + tableFromCode.TableName.String)
	}
}

//
func (db *Executor) getValFromOpinion(key string, def string) string {
	for i := 0; i < len(db.opinionList); i++ {
		opinionItem := db.opinionList[i]
		if opinionItem.Key == key {
			def = opinionItem.Val
		}
	}
	return def
}

func getTableInfoFromCode(tableFromCode Table) string {
	return " ENGINE " + tableFromCode.Engine.String + " COMMENT  " + tableFromCode.TableComment.String
}

// 获得某列的结构
func getColumnFromCode(fieldName string, fieldType string, fieldMap map[string]string) Column {
	var column Column
	//字段名
	column.ColumnName = StringFrom(fieldName)
	//字段数据类型
	column.DataType = StringFrom(getDataType(fieldType, fieldMap))
	//字段数据长度
	column.MaxLength = IntFrom(int64(getMaxLength(column.DataType.String, fieldMap)))
	//字段是否可以为空
	column.IsNullable = StringFrom(getNullAble(fieldMap))
	//字段注释
	column.ColumnComment = StringFrom(getComment(fieldMap))
	//扩展信息
	column.Extra = StringFrom(getExtra(fieldMap))
	//默认信息
	column.ColumnDefault = StringFrom(getDefaultVal(fieldMap))

	return column
}

// 转换tag成map
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

	if "PRIMARY" == index.KeyName {
		strArr = append(strArr, index.KeyName)
		strArr = append(strArr, "KEY")
		strArr = append(strArr, "(`"+index.ColumnName+"`)")
	} else {
		if 0 == index.NonUnique {
			strArr = append(strArr, "Unique")
			strArr = append(strArr, index.KeyName)
			strArr = append(strArr, "(`"+index.ColumnName+"`)")
		} else {
			strArr = append(strArr, "Index")
			strArr = append(strArr, index.KeyName)
			strArr = append(strArr, "(`"+index.ColumnName+"`)")
		}
	}

	return strings.Join(strArr, " ")
}

//将对象属性类型转换数据库字段数据类型
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
