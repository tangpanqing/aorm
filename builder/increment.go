package builder

// Increment 某字段自增
func (b *Builder) Increment(field interface{}, step int) (int64, error) {
	var vars []any
	vars = append(vars, step)
	whereStr, vars := b.handleWhere(vars, false)
	query := "UPDATE " + getTableNameByTable(b.table) + " SET " + getFieldNameByField(field) + "=" + getFieldNameByField(field) + "+?" + whereStr

	return b.execAffected(query, vars...)
}

// Decrement 某字段自减
func (b *Builder) Decrement(field interface{}, step int) (int64, error) {
	var vars []any
	vars = append(vars, step)
	whereStr, vars := b.handleWhere(vars, false)
	query := "UPDATE " + getTableNameByTable(b.table) + " SET " + getFieldNameByField(field) + "=" + getFieldNameByField(field) + "-?" + whereStr

	return b.execAffected(query, vars...)
}
