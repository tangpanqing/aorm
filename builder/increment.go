package builder

// Increment 某字段自增
func (b *Builder) Increment(field interface{}, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := b.handleWhere(paramList)
	sqlStr := "UPDATE " + getTableNameByTable(b.table) + " SET " + getFieldName(field) + "=" + getFieldName(field) + "+?" + whereStr

	return b.execAffected(sqlStr, paramList...)
}

// Decrement 某字段自减
func (b *Builder) Decrement(field interface{}, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := b.handleWhere(paramList)
	sqlStr := "UPDATE " + getTableNameByTable(b.table) + " SET " + getFieldName(field) + "=" + getFieldName(field) + "-?" + whereStr

	return b.execAffected(sqlStr, paramList...)
}
