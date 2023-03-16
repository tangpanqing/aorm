package builder

import (
	"fmt"
	"github.com/tangpanqing/aorm/utils"
	"reflect"
	"strings"
)

type GroupItem struct {
	Prefix []string
	Field  interface{}
}

type WhereItem struct {
	Prefix []string
	Field  interface{}
	Opt    string
	Val    interface{}
}

type SelectItem struct {
	FuncName string
	Prefix   []string
	Field    interface{}
	FieldNew interface{}
}

type SelectExpItem struct {
	Builder   **Builder
	FieldName interface{}
}

type OrderItem struct {
	Prefix    []string
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
	tableAlias []string
	condition  []JoinCondition
}

type JoinCondition struct {
	FieldOfCurrentTable interface{}
	Opt                 string
	FieldOfOtherTable   interface{}
	AliasOfOtherTable   []string
}

//GenWhereItem 产生一个 WhereItem,用作 where 条件里
func GenWhereItem(field interface{}, opt string, val interface{}, prefix ...string) WhereItem {
	return WhereItem{prefix, field, opt, val}
}

//GenHavingItem 产生一个 WhereItem,用作 having 条件里
func GenHavingItem(field interface{}, opt string, val interface{}) WhereItem {
	return WhereItem{[]string{}, field, opt, val}
}

//GenJoinCondition 产生一个 JoinCondition,用作 join 条件里
func GenJoinCondition(fieldOfCurrentTable interface{}, opt string, fieldOfOtherTable interface{}, aliasOfOtherTable ...string) JoinCondition {
	return JoinCondition{
		FieldOfCurrentTable: fieldOfCurrentTable,
		Opt:                 opt,
		FieldOfOtherTable:   fieldOfOtherTable,
		AliasOfOtherTable:   aliasOfOtherTable,
	}
}

//getPrefixByField 获取字段前缀,如果传入则使用传入值，默认使用该字段的表名
func getPrefixByField(valueOf reflect.Value, prefix ...string) string {
	str := ""
	if len(prefix) > 0 {
		str = prefix[0]
	} else {
		if reflect.Ptr == valueOf.Kind() {
			fieldPointer := valueOf.Pointer()
			tablePointer := getFieldMap(fieldPointer).TablePointer

			tableName := getTableMap(tablePointer)
			strArr := strings.Split(tableName, ".")
			str = utils.UnderLine(strArr[len(strArr)-1])
		} else {
			//str = fmt.Sprintf("%v", valueOf.Interface())
		}
	}

	return str
}

//getTableNameByTable 根据传入的表信息，获取表名
func getTableNameByTable(table interface{}) string {
	valueOf := reflect.ValueOf(table)
	if reflect.Ptr == valueOf.Kind() {
		return getTableMap(valueOf.Pointer())
	} else {
		return fmt.Sprintf("%v", table)
	}
}

//getTableNameByReflect 反射表名,优先从方法获取,没有方法则从名字获取
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

//getFieldNameByField 根据传入字段，获取字段名
func getFieldNameByField(field interface{}) string {
	return getFieldNameByReflectValue(reflect.ValueOf(field))
}

//getFieldNameByReflectNew 根据传入字段，获取字段名
func getFieldNameByReflectValue(valueOfField reflect.Value) string {
	if reflect.Ptr == valueOfField.Kind() {
		return getFieldMap(valueOfField.Pointer()).Name
	} else {
		return fmt.Sprintf("%v", valueOfField)
	}
}

//getFieldNameByStructField
func getFieldNameByStructField(field reflect.StructField) (string, map[string]string) {
	key := utils.UnderLine(field.Name)
	tag := field.Tag.Get("aorm")
	tagMap := getTagMap(tag)
	if column, ok := tagMap["column"]; ok {
		key = column
	}
	return key, tagMap
}

//getFieldMapByReflect 从结构体反射出来的属性名
func getFieldMapByReflect(destType reflect.Type) map[string][]int {
	fieldNameMap := make(map[string][]int)
	for i := 0; i < destType.NumField(); i++ {
		isMultiLevel := false
		if "struct" == destType.Field(i).Type.Kind().String() && (destType.Field(i).Type.Name() != "Int" &&
			destType.Field(i).Type.Name() != "Float" &&
			destType.Field(i).Type.Name() != "Time" &&
			destType.Field(i).Type.Name() != "String" &&
			destType.Field(i).Type.Name() != "Bool") {
			isMultiLevel = true
		}

		//fmt.Println(isMore, destType, destType.Field(i).Name, destType.Field(i).Type.Name(), destType.Field(i).Type.Kind().String())
		if isMultiLevel {
			for j := 0; j < destType.Field(i).Type.NumField(); j++ {
				fieldNameMap[destType.Field(i).Type.Field(j).Name] = []int{i, j}
			}
		} else {
			fieldNameMap[destType.Field(i).Name] = []int{i}
		}
	}
	return fieldNameMap
}

//getScansAddr 获取赋值的地址
func getScansAddr(columnNameList []string, fieldNameMap map[string][]int, destValue reflect.Value) []interface{} {
	var scans []interface{}
	for _, columnName := range columnNameList {
		fieldName := utils.CamelString(strings.ToLower(columnName))
		index, ok := fieldNameMap[fieldName]
		if ok {
			t := destValue
			for j := 0; j < len(index); j++ {
				t = t.Field(index[j])
			}
			scans = append(scans, t.Addr().Interface())
		} else {
			var emptyVal interface{}
			scans = append(scans, &emptyVal)
		}
	}
	return scans
}

//genJoinConditionStr 产生关联查询条件
func genJoinConditionStr(aliasOfCurrentTable string, joinCondition []JoinCondition) (string, []interface{}) {
	var paramList []interface{}
	var sqlList []string
	for i := 0; i < len(joinCondition); i++ {
		fieldNameOfCurrentTable := getFieldNameByField(joinCondition[i].FieldOfCurrentTable)

		if aliasOfCurrentTable == "" {
			aliasOfCurrentTable = getPrefixByField(reflect.ValueOf(joinCondition[i].FieldOfCurrentTable))
		}

		fieldNameOfOtherTable := getFieldNameByField(joinCondition[i].FieldOfOtherTable)

		if joinCondition[i].Opt == RawEq {
			aliasOfOtherTable := getPrefixByField(reflect.ValueOf(joinCondition[i].FieldOfOtherTable), joinCondition[i].AliasOfOtherTable...)
			if aliasOfOtherTable != "" {
				aliasOfOtherTable += "."
			}

			sqlList = append(sqlList, aliasOfCurrentTable+"."+fieldNameOfCurrentTable+"="+aliasOfOtherTable+fieldNameOfOtherTable)
		}

		if joinCondition[i].Opt == Eq {
			sqlList = append(sqlList, aliasOfCurrentTable+"."+fieldNameOfCurrentTable+"=?")
			paramList = append(paramList, fieldNameOfOtherTable)
		}
	}

	return strings.Join(sqlList, " AND "), paramList
}

//toAnyArr 将一个interface抽取成数组
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
