package builder

import (
	"fmt"
	"github.com/tangpanqing/aorm/helper"
	"reflect"
	"strings"
)

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

type OrderItem struct {
	Prefix    string
	Field     interface{}
	OrderType string
}

type LimitItem struct {
	offset   int
	pageSize int
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

//GenWhereItem 产生一个 WhereItem,用作 where 条件里
func GenWhereItem(field interface{}, opt string, val interface{}, prefix ...string) WhereItem {
	return WhereItem{getPrefixByField(field, prefix...), field, opt, val}
}

//GenHavingItem 产生一个 WhereItem,用作 having 条件里
func GenHavingItem(field interface{}, opt string, val interface{}) WhereItem {
	return WhereItem{"", field, opt, val}
}

//GenJoinCondition 产生一个 JoinCondition,用作 join 条件里
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

//getPrefixByField 获取字段前缀,如果传入则使用传入值，默认使用该字段的表名
func getPrefixByField(field interface{}, prefix ...string) string {
	str := ""
	if len(prefix) > 0 {
		str = prefix[0]
	} else {
		if field == nil {
			panic("当前field不能是nil")
		}

		valueOf := reflect.ValueOf(field)
		if reflect.Ptr == valueOf.Kind() {
			fieldPointer := valueOf.Pointer()
			tablePointer := getFieldMap(fieldPointer).TablePointer

			tableName := getTableMap(tablePointer)
			strArr := strings.Split(tableName, ".")
			str = helper.UnderLine(strArr[len(strArr)-1])
		} else {
			str = fmt.Sprintf("%v", field)
		}
	}

	return str
}

//getTableNameByTable 根据传入的表信息，获取表名
func getTableNameByTable(table interface{}) string {
	if table == nil {
		panic("当前table不能是nil")
	}

	valueOf := reflect.ValueOf(table)
	if reflect.Ptr == valueOf.Kind() {
		tableName := getTableMap(valueOf.Pointer())
		strArr := strings.Split(tableName, ".")
		return helper.UnderLine(strArr[len(strArr)-1])
	} else {
		return fmt.Sprintf("%v", table)
	}
}

//getFieldName 根据传入字段，获取字段名
func getFieldName(field interface{}) string {
	valueOf := reflect.ValueOf(field)
	if reflect.Ptr == valueOf.Kind() {
		return helper.UnderLine(getFieldMap(reflect.ValueOf(field).Pointer()).Name)
	} else {
		return fmt.Sprintf("%v", field)
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
		return helper.UnderLine(arr[len(arr)-1])
	}
}

//从结构体反射出来的属性名
func getFieldMapByReflect(destValue reflect.Value, destType reflect.Type) map[string]int {
	fieldNameMap := make(map[string]int)
	for i := 0; i < destValue.NumField(); i++ {
		fieldNameMap[destType.Field(i).Name] = i
	}

	return fieldNameMap
}

//获取赋值的地址
func getScansAddr(columnNameList []string, fieldNameMap map[string]int, destValue reflect.Value) []interface{} {
	var scans []interface{}
	for _, columnName := range columnNameList {
		fieldName := helper.CamelString(strings.ToLower(columnName))
		index, ok := fieldNameMap[fieldName]
		if ok {
			scans = append(scans, destValue.Field(index).Addr().Interface())
		} else {
			var emptyVal interface{}
			scans = append(scans, &emptyVal)
		}
	}

	return scans
}

//产生关联查询条件
func genJoinConditionStr(aliasOfCurrentTable string, joinCondition []JoinCondition, paramList []interface{}) (string, []interface{}) {
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

//将一个interface抽取成数组
func toAnyArr(val any) []any {
	var values []any
	switch val.(type) {
	case []int:
		for _, value := range val.([]int) {
			values = append(values, value)
		}
	case []int64:
		for _, value := range val.([]int64) {
			values = append(values, value)
		}
	case []float32:
		for _, value := range val.([]float32) {
			values = append(values, value)
		}
	case []float64:
		for _, value := range val.([]float64) {
			values = append(values, value)
		}
	case []string:
		for _, value := range val.([]string) {
			values = append(values, value)
		}
	}

	return values
}
