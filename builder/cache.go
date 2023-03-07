package builder

import (
	"reflect"
)

type FieldInfo struct {
	TablePointer uintptr
	Name         string
	TagMap       map[string]string
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
			key, _ := getFieldNameByStructField(typeof.Elem().Field(j))
			tag := typeof.Elem().Field(j).Tag.Get("aorm")

			setFieldMap(fieldPointer, FieldInfo{
				TablePointer: tablePointer,
				Name:         key,
				TagMap:       getTagMap(tag),
			})
		}
	}
}

func Comment(field interface{}) string {
	fieldPointer := reflect.ValueOf(field).Pointer()
	tagMap := getFieldMap(fieldPointer).TagMap
	val, ok := tagMap["comment"]
	if ok {
		return val
	} else {
		return ""
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
