package cache

import (
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/model"
	"reflect"
	"strings"
)

var TableMap = make(map[uintptr]string)
var FieldMap = make(map[uintptr]model.FieldInfo)

//Store 保存到缓存
func Store(destList ...interface{}) {
	for i := 0; i < len(destList); i++ {
		dest := destList[i]
		valueOf := reflect.ValueOf(dest)
		typeof := reflect.TypeOf(dest)

		tablePointer := valueOf.Pointer()
		SetTableMap(tablePointer, getTableNameByReflect(typeof, valueOf))

		for j := 0; j < valueOf.Elem().NumField(); j++ {
			addr := valueOf.Elem().Field(j).Addr().Pointer()
			name := typeof.Elem().Field(j).Name

			SetFieldMap(addr, model.FieldInfo{
				TablePointer: tablePointer,
				Name:         name,
			})
		}
	}
}

func SetTableMap(tablePointer uintptr, name string) {
	TableMap[tablePointer] = name
}

func GetTableMap(tablePointer uintptr) string {
	return TableMap[tablePointer]
}

func SetFieldMap(fieldPointer uintptr, fieldInfo model.FieldInfo) {
	FieldMap[fieldPointer] = fieldInfo
}

func GetFieldMap(fieldPointer uintptr) model.FieldInfo {
	return FieldMap[fieldPointer]
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
