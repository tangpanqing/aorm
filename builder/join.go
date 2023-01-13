package builder

// LeftJoin 链式操作,左联查询,例如 LeftJoin("project p", "p.project_id=o.project_id")
func (b *Builder) LeftJoin(table interface{}, condition []JoinCondition, alias ...string) *Builder {
	return b.join("LEFT JOIN", table, condition, alias...)
}

// RightJoin 链式操作,右联查询,例如 RightJoin("project p", "p.project_id=o.project_id")
func (b *Builder) RightJoin(table interface{}, condition []JoinCondition, alias ...string) *Builder {
	return b.join("RIGHT JOIN", table, condition, alias...)
}

// Join 链式操作,内联查询,例如 Join("project p", "p.project_id=o.project_id")
func (b *Builder) Join(table interface{}, condition []JoinCondition, alias ...string) *Builder {
	return b.join("INNER JOIN", table, condition, alias...)
}

func (b *Builder) join(joinType string, table interface{}, condition []JoinCondition, alias ...string) *Builder {
	b.joinList = append(b.joinList, JoinItem{joinType, table, alias, condition})
	return b
}
