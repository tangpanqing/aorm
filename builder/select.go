package builder

func (b *Builder) SelectAll(table interface{}) *Builder {
	return b.selectCommon("", "*", nil, getTableNameByTable(table))
}

// Select 链式操作-查询哪些字段,默认 *
func (b *Builder) Select(field interface{}, prefix ...string) *Builder {
	return b.selectCommon("", field, nil, prefix...)
}

func (b *Builder) SelectAs(field interface{}, fieldNew interface{}, prefix ...string) *Builder {
	return b.selectCommon("", field, fieldNew, prefix...)
}

// SelectCount 链式操作-count(field) as field_new
func (b *Builder) SelectCount(field interface{}, fieldNew interface{}, prefix ...string) *Builder {
	return b.selectCommon("count", field, fieldNew, prefix...)
}

// SelectSum 链式操作-sum(field) as field_new
func (b *Builder) SelectSum(field interface{}, fieldNew interface{}, prefix ...string) *Builder {
	return b.selectCommon("sum", field, fieldNew, prefix...)
}

// SelectMin 链式操作-min(field) as field_new
func (b *Builder) SelectMin(field interface{}, fieldNew interface{}, prefix ...string) *Builder {
	return b.selectCommon("min", field, fieldNew, prefix...)
}

// SelectMax 链式操作-max(field) as field_new
func (b *Builder) SelectMax(field interface{}, fieldNew interface{}, prefix ...string) *Builder {
	return b.selectCommon("max", field, fieldNew, prefix...)
}

// SelectAvg 链式操作-avg(field) as field_new
func (b *Builder) SelectAvg(field interface{}, fieldNew interface{}, prefix ...string) *Builder {
	return b.selectCommon("avg", field, fieldNew, prefix...)
}

func (b *Builder) selectCommon(funcName string, field interface{}, fieldNew interface{}, prefix ...string) *Builder {
	b.selectList = append(b.selectList, SelectItem{
		funcName,
		getPrefixByField(field, prefix...),
		field,
		fieldNew,
	})

	return b
}

// SelectExp 链式操作-表达式
func (b *Builder) SelectExp(dbSub **Builder, fieldName interface{}) *Builder {
	b.selectExpList = append(b.selectExpList, &SelectExpItem{
		Executor:  dbSub,
		FieldName: fieldName,
	})
	return b
}
