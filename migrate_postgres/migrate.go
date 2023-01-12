package migrate_postgres

import (
	"fmt"
	"github.com/tangpanqing/aorm/builder"
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/null"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type PgIndexes struct {
	Schemaname null.String
	Tablename  null.String
	Indexname  null.String
	Tablespace null.String
	Indexdef   null.String
}

type Table struct {
	TableName    null.String
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
	//执行者
	Builder *builder.Builder
}

//ShowCreateTable 查看创建表的ddl
func (mm *MigrateExecutor) ShowCreateTable(tableName string) string {
	var str string
	mm.Builder.RawSql("show create table "+tableName).Value("Create Table", &str)
	return str
}

//MigrateCommon 迁移的主要过程
func (mm *MigrateExecutor) MigrateCommon(tableName string, typeOf reflect.Type, valueOf reflect.Value) error {
	tableFromCode := mm.getTableFromCode(tableName, typeOf, valueOf)
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
		indexesFromDb := mm.getIndexesFromDb(tableName)

		mm.modifyTable(tableFromCode, columnsFromCode, indexesFromCode, tableFromDb, columnsFromDb, indexesFromDb)
	} else {
		mm.createTable(tableFromCode, columnsFromCode, indexesFromCode)
	}

	return nil
}

func (mm *MigrateExecutor) getTableFromCode(tableName string, typeOf reflect.Type, valueOf reflect.Value) Table {
	table := Table{
		TableName:    null.StringFrom(tableName),
		TableComment: null.StringFrom("''"),
	}

	method, isSet := typeOf.MethodByName("TableOpinion")
	if isSet {
		var paramList []reflect.Value
		paramList = append(paramList, valueOf)
		valueList := method.Func.Call(paramList)
		i := valueList[0].Interface()
		m := i.(map[string]string)

		m["COMMENT"] = "'" + m["COMMENT"] + "'"
		table.TableComment = null.StringFrom(m["COMMENT"])
	}

	return table
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
	err := mm.Builder.RawSql("select current_database()").Value("current_database", &dbName)
	if err != nil {
		return "", err
	}

	return dbName, nil
}

func (mm *MigrateExecutor) getTableFromDb(dbName string, tableName string) []Table {
	sql := "select a.relname as TABLE_NAME, b.description as TABLE_COMMENT from pg_class a left join (select * from pg_description where objsubid =0) b on a.oid = b.objoid where a.relname in (select tablename from pg_tables where schemaname = 'public' and tablename = " + "'" + tableName + "') order by a.relname asc"
	var dataList []Table
	mm.Builder.RawSql(sql).GetMany(&dataList)
	for i := 0; i < len(dataList); i++ {
		dataList[i].TableComment = null.StringFrom("'" + dataList[i].TableComment.String + "'")
	}

	return dataList
}

func (mm *MigrateExecutor) getColumnsFromDb(dbName string, tableName string) []Column {
	var columnsFromDb []Column

	sqlColumn := "select column_name,data_type,character_maximum_length as max_length,column_default,'' as COLUMN_COMMENT, is_nullable from information_schema.columns where table_schema='public' and table_name=" + "'" + tableName + "'"

	mm.Builder.RawSql(sqlColumn).GetMany(&columnsFromDb)

	for j := 0; j < len(columnsFromDb); j++ {
		if columnsFromDb[j].DataType.String == "character varying" {
			columnsFromDb[j].DataType = null.StringFrom("varchar")
		}

		if columnsFromDb[j].DataType.String == "double precision" {
			columnsFromDb[j].DataType = null.StringFrom("float")
		}

		if columnsFromDb[j].DataType.String == "timestamp without time zone" {
			columnsFromDb[j].DataType = null.StringFrom("timestamp")
		}
	}

	return columnsFromDb
}

func (mm *MigrateExecutor) getIndexesFromDb(tableName string) []Index {
	sqlIndex := "select * from pg_indexes where tablename=" + "'" + tableName + "'"
	var sqliteMasterList []PgIndexes
	mm.Builder.RawSql(sqlIndex).GetMany(&sqliteMasterList)

	var indexesFromDb []Index
	for i := 0; i < len(sqliteMasterList); i++ {
		indexName := sqliteMasterList[i].Indexname.String
		sql := sqliteMasterList[i].Indexdef.String

		t := 1
		if strings.Index(sql, "UNIQUE") != -1 {
			t = 0
		}

		compileRegex := regexp.MustCompile("INDEX\\s(.*?)\\sON.*?\\((.*?)\\)")
		matchArr := compileRegex.FindAllStringSubmatch(sql, -1)

		//主键索引
		if indexName == tableName+"_pkey" {
			indexesFromDb = append(indexesFromDb, Index{
				NonUnique:  null.IntFrom(int64(t)),
				ColumnName: null.StringFrom(matchArr[0][2]),
				KeyName:    null.StringFrom("PRIMARY"),
			})
		} else {
			indexesFromDb = append(indexesFromDb, Index{
				NonUnique:  null.IntFrom(int64(t)),
				ColumnName: null.StringFrom(matchArr[0][2]),
				KeyName:    null.StringFrom(matchArr[0][1]),
			})
		}
	}

	return indexesFromDb
}

func (mm *MigrateExecutor) modifyTable(tableFromCode Table, columnsFromCode []Column, indexesFromCode []Index, tableFromDb Table, columnsFromDb []Column, indexesFromDb []Index) {
	//if tableFromCode.TableComment != tableFromDb.TableComment {
	//	mm.modifyTableComment(tableFromCode)
	//}

	for i := 0; i < len(columnsFromCode); i++ {
		isFind := 0
		columnCode := columnsFromCode[i]

		for j := 0; j < len(columnsFromDb); j++ {
			columnDb := columnsFromDb[j]
			if columnCode.ColumnName.String == columnDb.ColumnName.String {
				isFind = 1
				if columnCode.DataType.String != columnDb.DataType.String {
					fmt.Println(columnCode.ColumnName.String, columnCode.DataType.String, columnDb.DataType.String)

					sql := "ALTER TABLE " + tableFromCode.TableName.String + " alter COLUMN " + getColumnStr(columnCode, "driver")
					//fmt.Println(model)

					_, err := mm.Builder.RawSql(sql).Exec()
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("修改属性:" + sql)
					}
				}
			}
		}

		if isFind == 0 {
			sql := "ALTER TABLE " + tableFromCode.TableName.String + " ADD " + getColumnStr(columnCode, "")
			_, err := mm.Builder.RawSql(sql).Exec()
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
					_, err := mm.Builder.RawSql(sql).Exec()
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("修改索引:" + sql)
					}
				}
			}
		}

		if isFind == 0 {
			mm.createIndex(tableFromCode.TableName.String, indexCode)
		}
	}
}

func (mm *MigrateExecutor) modifyTableComment(tableFromCode Table) {
	sql := "ALTER TABLE " + tableFromCode.TableName.String + " Comment " + tableFromCode.TableComment.String
	_, err := mm.Builder.RawSql(sql).Exec()
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
		fieldArr = append(fieldArr, getColumnStr(column, ""))
	}

	for i := 0; i < len(indexesFromCode); i++ {
		index := indexesFromCode[i]
		if index.KeyName.String == "PRIMARY" {
			fieldArr = append(fieldArr, "PRIMARY KEY ("+index.ColumnName.String+")")
		}
	}

	sql := "CREATE TABLE " + tableFromCode.TableName.String + " (\n" + strings.Join(fieldArr, ",\n") + "\n) " + ";"

	_, err := mm.Builder.RawSql(sql).Exec()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("创建表:" + tableFromCode.TableName.String)
	}

	//创建其他索引
	for i := 0; i < len(indexesFromCode); i++ {
		index := indexesFromCode[i]
		if index.KeyName.String != "PRIMARY" {
			mm.createIndex(tableFromCode.TableName.String, index)
		}
	}
}

func (mm *MigrateExecutor) createIndex(tableName string, index Index) {
	keyType := ""
	if index.NonUnique.Int64 == 0 {
		keyType = "UNIQUE"
	}

	sql := "CREATE " + keyType + " INDEX " + index.KeyName.String + " on " + tableName + " (" + index.ColumnName.String + ")"
	_, err := mm.Builder.RawSql(sql).Exec()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("增加索引:" + sql)
	}
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

func getColumnStr(column Column, f string) string {
	var strArr []string
	strArr = append(strArr, column.ColumnName.String)

	//类型
	if column.Extra.String == "auto_increment" {
		strArr = append(strArr, "serial")
	} else {
		if column.MaxLength.Int64 == 0 {
			if column.DataType.String == "varchar" {
				strArr = append(strArr, column.DataType.String+"(255)")
			} else {
				strArr = append(strArr, f+" "+column.DataType.String)
			}
		} else {
			strArr = append(strArr, column.DataType.String+"("+strconv.Itoa(int(column.MaxLength.Int64))+")")
		}
	}

	if column.ColumnDefault.String != "" {
		strArr = append(strArr, "DEFAULT '"+column.ColumnDefault.String+"'")
	}

	if column.IsNullable.String == "NO" {
		//strArr = append(strArr, "NOT NULL")
	}

	if column.ColumnComment.String != "" {
		//strArr = append(strArr, "COMMENT '"+column.ColumnComment.String+"'")
	}

	if column.Extra.String != "" {
		//strArr = append(strArr, column.Extra.String)
	}

	return strings.Join(strArr, " ")
}

func getIndexStr(index Index) string {
	var strArr []string

	if "PRIMARY" == index.KeyName.String {
		strArr = append(strArr, index.KeyName.String)
		strArr = append(strArr, "KEY")
		strArr = append(strArr, "("+index.ColumnName.String+")")
	} else {
		if 0 == index.NonUnique.Int64 {
			strArr = append(strArr, "Unique")
			strArr = append(strArr, index.KeyName.String)
			strArr = append(strArr, "("+index.ColumnName.String+")")
		} else {
			strArr = append(strArr, "Index")
			strArr = append(strArr, index.KeyName.String)
			strArr = append(strArr, "("+index.ColumnName.String+")")
		}
	}

	return strings.Join(strArr, " ")
}

func getDataType(fieldType string, fieldMap map[string]string) string {
	var DataType string

	dataTypeVal, dataTypeOk := fieldMap["driver"]
	if dataTypeOk {
		DataType = dataTypeVal
		if "tinyint" == DataType {
			DataType = "integer"
		}
		if "double" == DataType {
			DataType = "float"
		}
	} else {
		if "Int" == fieldType {
			DataType = "integer"
		}
		if "String" == fieldType {
			DataType = "varchar"
		}
		if "Bool" == fieldType {
			//DataType = "tinyint"
			DataType = "boolean"
		}
		if "Time" == fieldType {
			DataType = "date"
			DataType = "timestamp"
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
