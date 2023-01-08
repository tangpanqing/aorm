package builder

var TableMap = make(map[uintptr]string)
var FieldMap = make(map[uintptr]FieldInfo)

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
