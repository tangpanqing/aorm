package builder

import (
	"github.com/tangpanqing/aorm/utils"
	"reflect"
)

type FieldInfo struct {
	TablePointer uintptr
	Name         string
}

var TableMap = make(map[uintptr]string)
var FieldMap = make(map[uintptr]FieldInfo)

//Store 保存到缓存
func Store(destList ...interface{}) {
	for i := 0; i < len(destList); i++ {
		dest := destList[i]
		valueOf := reflect.ValueOf(dest)
		typeof := reflect.TypeOf(dest)

		tablePointer := valueOf.Pointer()
		tableName := getTableNameByReflect(typeof, valueOf)
		setTableMap(tablePointer, tableName)

		for j := 0; j < valueOf.Elem().NumField(); j++ {
			fieldPointer := valueOf.Elem().Field(j).Addr().Pointer()
			key, _ := getFieldNameByReflect(typeof.Elem().Field(j))
			setFieldMap(fieldPointer, FieldInfo{
				TablePointer: tablePointer,
				Name:         key,
			})
		}
	}
}

func setTableMap(tablePointer uintptr, name string) {
	TableMap[tablePointer] = name
}

func getTableMap(tablePointer uintptr) string {
	return TableMap[tablePointer]
}

func setFieldMap(fieldPointer uintptr, fieldInfo FieldInfo) {
	FieldMap[fieldPointer] = fieldInfo
}

func getFieldMap(fieldPointer uintptr) FieldInfo {
	return FieldMap[fieldPointer]
}

func getFieldNameByReflect(field reflect.StructField) (string, map[string]string) {
	key := utils.UnderLine(field.Name)
	tag := field.Tag.Get("aorm")
	tagMap := getTagMap(tag)
	if column, ok := tagMap["column"]; ok {
		key = column
	}

	return key, tagMap
}
