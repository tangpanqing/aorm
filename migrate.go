package aorm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Table struct {
	TableName string
	Engine    string
	Comment   string
}

type Column struct {
	ColumnName    string
	ColumnDefault string
	IsNullable    string
	DataType      string //数据类型 varchar,bigint,int
	MaxLength     int    //数据最大长度 20
	ColumnType    string //列类型 varchar(20)
	ColumnComment string
	Extra         string //扩展信息 auto_increment
	DefaultVal    string //默认值
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

	db.OpinionList = append(db.OpinionList, OpinionItem{Key: key, Val: val})

	return db
}

func (db *Executor) ShowCreateTable(tableName string) string {
	list, _ := db.Query("show create table " + tableName)
	return list[0]["Create Table"].(string)
}

// Migrate 迁移数据库结构,需要输入数据库名,表名自动获取
func (db *Executor) AutoMigrate(dest interface{}) {
	typeOf := reflect.TypeOf(dest)
	arr := strings.Split(typeOf.String(), ".")
	tableName := UnderLine(arr[len(arr)-1])

	db.migrateCommon(tableName, typeOf)
}

// AutoMigrate 自动迁移数据库结构,需要输入数据库名,表名
func (db *Executor) Migrate(tableName string, dest interface{}) {
	typeOf := reflect.TypeOf(dest)
	db.migrateCommon(tableName, typeOf)
}

func (db *Executor) migrateCommon(tableName string, typeOf reflect.Type) {
	tableFromCode := db.getTableFromCode(tableName)
	columnsFromCode := db.getColumnsFromCode(typeOf)
	indexsFromCode := db.getIndexsFromCode(typeOf, tableFromCode)

	//获取数据库名称
	dbNameRows, _ := db.Query("SELECT DATABASE()")
	dbName := dbNameRows[0]["DATABASE()"].(string)

	//查询表信息,如果找不到就新建
	sql := "SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA =" + "'" + dbName + "' AND TABLE_NAME =" + "'" + tableName + "'"
	dataList, _ := db.Query(sql)
	if len(dataList) != 0 {
		tableFromDb := getTableFromDb(dataList)
		columnsFromDb := db.getColumnsFromDb(dbName, tableName)
		indexsFromDb := db.getIndexsFromDb(tableName)

		db.modifyTable(tableFromCode, columnsFromCode, indexsFromCode, tableFromDb, columnsFromDb, indexsFromDb)
	} else {
		db.createTable(tableFromCode, columnsFromCode, indexsFromCode)
	}
}

func (db *Executor) getTableFromCode(tableName string) Table {
	var tableFromCode Table
	tableFromCode.TableName = tableName
	tableFromCode.Engine = db.getValFromOpinion("ENGINE", "MyISAM")
	tableFromCode.Comment = db.getValFromOpinion("COMMENT", "")

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
				KeyName:    "idx_" + tableFromCode.TableName + "_" + fieldName,
			})
		}

		_, indexIs := fieldMap["index"]
		if indexIs {
			indexsFromCode = append(indexsFromCode, Index{
				NonUnique:  1,
				ColumnName: fieldName,
				KeyName:    "idx_" + tableFromCode.TableName + "_" + fieldName,
			})
		}
	}

	return indexsFromCode
}

func getTableFromDb(dataList []map[string]interface{}) Table {
	var tableFromDb Table
	tableFromDb.TableName = fmt.Sprintf("%v", dataList[0]["TABLE_NAME"])
	tableFromDb.Engine = fmt.Sprintf("%v", dataList[0]["ENGINE"])
	tableFromDb.Comment = "'" + fmt.Sprintf("%v", dataList[0]["TABLE_COMMENT"]) + "'"

	return tableFromDb
}

func (db *Executor) getColumnsFromDb(dbName string, tableName string) []Column {
	var columnsFromDb []Column

	sqlColumn := "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA =" + "'" + dbName + "' AND TABLE_NAME =" + "'" + tableName + "'"
	dataColumn, _ := db.Query(sqlColumn)

	for j := 0; j < len(dataColumn); j++ {
		dataType := dataColumn[j]["DATA_TYPE"].(string)
		maxLength, _ := strconv.Atoi(fmt.Sprintf("%v", dataColumn[j]["CHARACTER_MAXIMUM_LENGTH"]))
		if dataType == "text" && maxLength == 65535 {
			maxLength = 0
		}

		defaultVal := ""
		if dataColumn[j]["COLUMN_DEFAULT"] != nil {
			defaultVal = dataColumn[j]["COLUMN_DEFAULT"].(string)
		}

		columnsFromDb = append(columnsFromDb, Column{
			ColumnName:    dataColumn[j]["COLUMN_NAME"].(string),
			DataType:      dataType,
			IsNullable:    dataColumn[j]["IS_NULLABLE"].(string),
			MaxLength:     maxLength,
			ColumnType:    dataColumn[j]["COLUMN_TYPE"].(string),
			ColumnComment: dataColumn[j]["COLUMN_COMMENT"].(string),
			Extra:         dataColumn[j]["EXTRA"].(string),
			DefaultVal:    defaultVal,
		})
	}

	return columnsFromDb
}

func (db *Executor) getIndexsFromDb(tableName string) []Index {
	sqlIndex := "SHOW INDEXES FROM " + tableName
	dataIndex, _ := db.Query(sqlIndex)

	var indexsFromDb []Index
	for j := 0; j < len(dataIndex); j++ {
		nonUnique, _ := strconv.Atoi(fmt.Sprintf("%v", dataIndex[j]["Non_unique"]))
		indexsFromDb = append(indexsFromDb, Index{
			ColumnName: fmt.Sprintf("%v", dataIndex[j]["Column_name"]),
			KeyName:    fmt.Sprintf("%v", dataIndex[j]["Key_name"]),
			NonUnique:  nonUnique,
		})
	}

	return indexsFromDb
}

// 修改表
func (db *Executor) modifyTable(tableFromCode Table, columnsFromCode []Column, indexsFromCode []Index, tableFromDb Table, columnsFromDb []Column, indexsFromDb []Index) {
	if tableFromCode.Engine != tableFromDb.Engine {
		sql := "ALTER TABLE " + tableFromCode.TableName + " Engine " + tableFromCode.Engine
		_, err := db.Exec(sql)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("修改表:" + sql)
		}
	}

	if tableFromCode.Comment != tableFromDb.Comment {
		sql := "ALTER TABLE " + tableFromCode.TableName + " Comment " + tableFromCode.Comment
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
				if columnCode.DataType != columnDb.DataType || columnCode.MaxLength != columnDb.MaxLength || columnCode.ColumnComment != columnDb.ColumnComment || columnCode.Extra != columnDb.Extra || columnCode.DefaultVal != columnDb.DefaultVal {
					sql := "ALTER TABLE " + tableFromCode.TableName + " MODIFY " + getColumnStr(columnCode)
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
			sql := "ALTER TABLE " + tableFromCode.TableName + " ADD " + getColumnStr(columnCode)
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
					sql := "ALTER TABLE " + tableFromCode.TableName + " MODIFY " + getIndexStr(indexCode)
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
			sql := "ALTER TABLE " + tableFromCode.TableName + " ADD " + getIndexStr(indexCode)
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

	sqlStr := "CREATE TABLE `" + tableFromCode.TableName + "` (\n" + strings.Join(fieldArr, ",\n") + "\n) " + getTableInfoFromCode(tableFromCode) + ";"
	_, err := db.Exec(sqlStr)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("创建表:" + tableFromCode.TableName)
	}
}

//
func (db *Executor) getValFromOpinion(key string, def string) string {
	for i := 0; i < len(db.OpinionList); i++ {
		opinionItem := db.OpinionList[i]
		if opinionItem.Key == key {
			def = opinionItem.Val
		}
	}
	return def
}

func getTableInfoFromCode(tableFromCode Table) string {
	return " ENGINE " + tableFromCode.Engine + " COMMENT  " + tableFromCode.Comment
}

// 获得某列的结构
func getColumnFromCode(fieldName string, fieldType string, fieldMap map[string]string) Column {
	var column Column
	//字段名
	column.ColumnName = fieldName
	//字段数据类型
	column.DataType = getDataType(fieldType, fieldMap)
	//字段数据长度
	maxLength := getMaxLength(column.DataType, fieldMap)
	columnType := column.DataType
	if maxLength > 0 {
		columnType = columnType + "(" + strconv.Itoa(maxLength) + ")"
	}
	column.MaxLength = maxLength
	//字段是否可以为空
	column.IsNullable = getNullAble(fieldMap)
	//字段注释
	column.ColumnComment = getComment(fieldMap)
	//字段类型
	column.ColumnType = columnType
	//扩展信息
	column.Extra = getExtra(fieldMap)
	//默认信息
	column.DefaultVal = getDefaultVal(fieldMap)

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
	strArr = append(strArr, column.ColumnName)
	if column.MaxLength == 0 {
		if column.DataType == "varchar" {
			strArr = append(strArr, column.DataType+"(255)")
		} else {
			strArr = append(strArr, column.DataType)
		}
	} else {
		strArr = append(strArr, column.DataType+"("+strconv.Itoa(column.MaxLength)+")")
	}

	if column.DefaultVal != "" {
		strArr = append(strArr, "DEFAULT '"+column.DefaultVal+"'")
	}

	if column.IsNullable == "NO" {
		strArr = append(strArr, "NOT NULL")
	}

	if column.ColumnComment != "" {
		strArr = append(strArr, "COMMENT '"+column.ColumnComment+"'")
	}

	if column.Extra != "" {
		strArr = append(strArr, column.Extra)
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
