package builder

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

var RawEq = "rawEq"
var TableMap = make(map[uintptr]string)
var FieldMap = make(map[uintptr]FieldInfo)

type FieldInfo struct {
	TablePointer uintptr
	Name         string
}

type WhereItem struct {
	Prefix string
	Field  interface{}
	Opt    string
	Val    interface{}
}

type SelectItem struct {
	FuncName string
	Prefix   string
	Field    interface{}
	FieldNew interface{}
}

type JoinItem struct {
	joinType   string
	table      interface{}
	tableAlias string
	condition  []JoinCondition
}

type JoinCondition struct {
	FieldOfCurrentTable interface{}
	Opt                 string
	FieldOfOtherTable   interface{}
	AliasOfOtherTable   string
}

func Store(destList ...interface{}) {
	for i := 0; i < len(destList); i++ {
		dest := destList[i]
		valueOf := reflect.ValueOf(dest)
		typeof := reflect.TypeOf(dest)

		tablePointer := valueOf.Pointer()
		TableMap[tablePointer] = typeof.String()
		for j := 0; j < valueOf.Elem().NumField(); j++ {
			addr := valueOf.Elem().Field(j).Addr().Pointer()
			name := typeof.Elem().Field(j).Name
			FieldMap[addr] = FieldInfo{
				TablePointer: tablePointer,
				Name:         name,
			}
		}
	}
}

func GenJoinCondition(fieldOfCurrentTable interface{}, opt string, fieldOfOtherTable interface{}, aliasOfOtherTable ...string) JoinCondition {
	alias := ""
	if len(aliasOfOtherTable) > 0 {
		alias = aliasOfOtherTable[0]
	}

	return JoinCondition{
		FieldOfCurrentTable: fieldOfCurrentTable,
		Opt:                 opt,
		FieldOfOtherTable:   fieldOfOtherTable,
		AliasOfOtherTable:   alias,
	}
}

//func (b *Builder) GetOne(dest interface{}) {
//	var paramList []interface{}
//
//	table := getTableName(b.table)
//	selectStr, paramList := b.handleSelect(paramList)
//	joinStr, paramList := b.handleJoin(paramList)
//	whereStr, paramList := b.handleWhere(paramList)
//
//	var sql = "SELECT " + selectStr + " FROM " + table + " " + b.tableAlias + joinStr + whereStr
//
//	fmt.Println(sql)
//	//fmt.Println(paramList)
//}

func (b *Builder) handleSelect(paramList []interface{}) (string, []interface{}) {
	if len(b.selectList) == 0 {
		return "*", paramList
	}

	var sqlList []string
	for i := 0; i < len(b.selectList); i++ {
		nameOfField := ""
		typeOfField := reflect.TypeOf(b.selectList[i].Field)
		valueOfField := reflect.ValueOf(b.selectList[i].Field)

		if reflect.String == typeOfField.Kind() {
			nameOfField = fmt.Sprintf("%v", b.selectList[i].Field)
		} else if reflect.Ptr == typeOfField.Kind() {
			nameOfField = UnderLine(FieldMap[valueOfField.Pointer()].Name)
		} else {
			panic("其他类型")
		}

		nameOfFieldNew := ""
		if b.selectList[i].FieldNew != nil {
			typeOfFieldNew := reflect.TypeOf(b.selectList[i].FieldNew)
			valueOfFieldNew := reflect.ValueOf(b.selectList[i].FieldNew)
			if reflect.String == typeOfFieldNew.Kind() {
				nameOfFieldNew = fmt.Sprintf("%v", b.selectList[i].FieldNew)
			} else if reflect.Ptr == typeOfFieldNew.Kind() {
				nameOfFieldNew = UnderLine(FieldMap[valueOfFieldNew.Pointer()].Name)
			} else {
				panic("其他类型")
			}

			if nameOfFieldNew != "" {
				nameOfFieldNew = " AS " + nameOfFieldNew
			}
		}

		sqlList = append(sqlList, b.selectList[i].Prefix+"."+nameOfField+nameOfFieldNew)
	}

	return strings.Join(sqlList, ","), paramList
}

//func (b *Builder) handleWhere(paramList []interface{}) (string, []interface{}) {
//	if len(b.whereList) == 0 {
//		return "", paramList
//	}
//
//	str, paramList := getWhereStr(b.whereList, paramList)
//
//	return " WHERE " + str, paramList
//}

func getPrefixByField(field interface{}, alias ...string) string {
	str := ""
	if len(alias) > 0 {
		str = alias[0]
	} else {
		str = getTableNameByField(field)
	}

	return str
}

func UnderLine(s string) string {
	var output []rune
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToLower(r))
			continue
		}
		if unicode.IsUpper(r) {
			output = append(output, '_')
		}
		output = append(output, unicode.ToLower(r))
	}
	return string(output)
}

func getTableNameByTable(table interface{}) string {
	if table == nil {
		panic("当前table不能是nil")
	}
	tableName := TableMap[reflect.ValueOf(table).Pointer()]
	strArr := strings.Split(tableName, ".")
	return UnderLine(strArr[len(strArr)-1])
}

func getTableNameByField(field interface{}) string {
	valueOf := reflect.ValueOf(field)
	if reflect.Ptr == valueOf.Kind() {
		tablePointer := FieldMap[reflect.ValueOf(field).Pointer()].TablePointer

		tableName := TableMap[tablePointer]
		strArr := strings.Split(tableName, ".")
		return UnderLine(strArr[len(strArr)-1])
	} else {
		return fmt.Sprintf("%v", field)
	}
}

func getWhereStrForJoin(aliasOfCurrentTable string, joinCondition []JoinCondition, paramList []interface{}) (string, []interface{}) {
	var sqlList []string
	for i := 0; i < len(joinCondition); i++ {
		fieldNameOfCurrentTable := getFieldName(joinCondition[i].FieldOfCurrentTable)
		if joinCondition[i].Opt == RawEq {
			if aliasOfCurrentTable == "" {
				aliasOfCurrentTable = getPrefixByField(joinCondition[i].FieldOfCurrentTable)
			}

			aliasOfOtherTable := joinCondition[i].AliasOfOtherTable
			if aliasOfOtherTable == "" {
				aliasOfOtherTable = getPrefixByField(joinCondition[i].FieldOfOtherTable)
			}

			fieldNameOfOtherTable := getFieldName(joinCondition[i].FieldOfOtherTable)
			sqlList = append(sqlList, aliasOfCurrentTable+"."+fieldNameOfCurrentTable+"="+aliasOfOtherTable+"."+fieldNameOfOtherTable)
		}
	}

	return strings.Join(sqlList, " AND "), paramList
}

func getFieldName(field interface{}) string {
	return UnderLine(FieldMap[reflect.ValueOf(field).Pointer()].Name)
}

func getWhereStr(whereList []WhereItem, paramList []interface{}) (string, []interface{}) {
	var sqlList []string
	for i := 0; i < len(whereList); i++ {
		prefix := whereList[i].Prefix
		field := getFieldName(whereList[i].Field)
		if whereList[i].Opt == "=" {
			sqlList = append(sqlList, prefix+"."+field+"="+"?")
			paramList = append(paramList, whereList[i].Val)
		}

		if whereList[i].Opt == "rawEq" {
			value := getFieldName(whereList[i].Val)
			sqlList = append(sqlList, prefix+"."+field+"="+value)
		}
	}

	return strings.Join(sqlList, " AND "), paramList
}
