package builder

// Select 链式操作-查询哪些字段,默认 *
func (ex *Builder) Select(fields ...string) *Builder {
	ex.selectList = append(ex.selectList, fields...)
	return ex
}

// SelectCount 链式操作-count(field) as field_new
func (ex *Builder) SelectCount(field string, fieldNew string) *Builder {
	ex.selectList = append(ex.selectList, "count("+field+") AS "+fieldNew)
	return ex
}

// SelectSum 链式操作-sum(field) as field_new
func (ex *Builder) SelectSum(field string, fieldNew string) *Builder {
	ex.selectList = append(ex.selectList, "sum("+field+") AS "+fieldNew)
	return ex
}

// SelectMin 链式操作-min(field) as field_new
func (ex *Builder) SelectMin(field string, fieldNew string) *Builder {
	ex.selectList = append(ex.selectList, "min("+field+") AS "+fieldNew)
	return ex
}

// SelectMax 链式操作-max(field) as field_new
func (ex *Builder) SelectMax(field string, fieldNew string) *Builder {
	ex.selectList = append(ex.selectList, "max("+field+") AS "+fieldNew)
	return ex
}

// SelectAvg 链式操作-avg(field) as field_new
func (ex *Builder) SelectAvg(field string, fieldNew string) *Builder {
	ex.selectList = append(ex.selectList, "avg("+field+") AS "+fieldNew)
	return ex
}

// SelectExp 链式操作-表达式
func (ex *Builder) SelectExp(dbSub **Builder, fieldName string) *Builder {
	ex.selectExpList = append(ex.selectExpList, &SelectItem{
		Executor:  dbSub,
		FieldName: fieldName,
	})
	return ex
}
