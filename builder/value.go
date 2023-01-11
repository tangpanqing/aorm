package builder

import "reflect"

// Value 字段值
func (b *Builder) Value(field interface{}, dest interface{}) error {
	b.Select(field).Limit(0, 1)

	fieldName := getFieldName(field)

	rows, errRows := b.GetRows()
	defer rows.Close()
	if errRows != nil {
		return errRows
	}

	destValue := reflect.ValueOf(dest).Elem()

	//从数据库中读出来的字段名字
	columnNameList, errColumns := rows.Columns()
	if errColumns != nil {
		return errColumns
	}

	for rows.Next() {
		var scans []interface{}
		for _, columnName := range columnNameList {
			if fieldName == columnName {
				scans = append(scans, destValue.Addr().Interface())
			} else {
				var emptyVal interface{}
				scans = append(scans, &emptyVal)
			}
		}

		err := rows.Scan(scans...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Pluck 获取某一列的值
func (b *Builder) Pluck(field interface{}, values interface{}) error {
	b.Select(field)
	fieldName := getFieldName(field)

	rows, errRows := b.GetRows()
	defer rows.Close()
	if errRows != nil {
		return errRows
	}

	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	destValue := reflect.New(destType).Elem()

	//从数据库中读出来的字段名字
	columnNameList, errColumns := rows.Columns()
	if errColumns != nil {
		return errColumns
	}

	for rows.Next() {
		var scans []interface{}
		for _, columnName := range columnNameList {
			if fieldName == columnName {
				scans = append(scans, destValue.Addr().Interface())
			} else {
				var emptyVal interface{}
				scans = append(scans, &emptyVal)
			}
		}

		err := rows.Scan(scans...)
		if err != nil {
			return err
		}

		destSlice.Set(reflect.Append(destSlice, destValue))
	}

	return nil
}
