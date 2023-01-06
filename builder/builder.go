package builder

import (
	"fmt"
	"github.com/tangpanqing/aorm/helper"
	"reflect"
	"strings"
)

var TableMap = make(map[uintptr]string)
var FieldMap = make(map[uintptr]FieldInfo)

type FieldInfo struct {
	TablePointer uintptr
	Name         string
}

type GroupItem struct {
	Prefix string
	Field  interface{}
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

type SelectExpItem struct {
	Builder   **Builder
	FieldName interface{}
}

type JoinItem struct {
	joinType   string
	table      interface{}
	tableAlias string
	condition  []JoinCondition
}

type OrderItem struct {
	Prefix    string
	Field     interface{}
	OrderType string
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

func getPrefixByField(field interface{}, alias ...string) string {
	str := ""
	if len(alias) > 0 {
		str = alias[0]
	} else {
		str = getTableNameByField(field)
	}

	return str
}

func getTableNameByTable(table interface{}) string {
	if table == nil {
		panic("当前table不能是nil")
	}

	valueOf := reflect.ValueOf(table)
	if reflect.Ptr == valueOf.Kind() {
		tableName := TableMap[valueOf.Pointer()]
		strArr := strings.Split(tableName, ".")
		return helper.UnderLine(strArr[len(strArr)-1])
	} else {
		return fmt.Sprintf("%v", table)
	}
}

func getTableNameByField(field interface{}) string {
	valueOf := reflect.ValueOf(field)
	if reflect.Ptr == valueOf.Kind() {
		tablePointer := FieldMap[reflect.ValueOf(field).Pointer()].TablePointer

		tableName := TableMap[tablePointer]
		strArr := strings.Split(tableName, ".")
		return helper.UnderLine(strArr[len(strArr)-1])
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
	valueOf := reflect.ValueOf(field)
	if reflect.Ptr == valueOf.Kind() {
		return helper.UnderLine(FieldMap[reflect.ValueOf(field).Pointer()].Name)
	} else {
		return fmt.Sprintf("%v", field)
	}
}
