package builder

// LeftJoin 链式操作,左联查询,例如 LeftJoin("project p", "p.project_id=o.project_id")
func (b *Builder) LeftJoin(table interface{}, condition []JoinCondition, alias ...string) *Builder {
	b.join("LEFT JOIN", table, condition, alias...)
	return b
}

// RightJoin 链式操作,右联查询,例如 RightJoin("project p", "p.project_id=o.project_id")
func (b *Builder) RightJoin(table interface{}, condition []JoinCondition, alias ...string) *Builder {
	b.join("RIGHT JOIN", table, condition, alias...)
	return b
}

// Join 链式操作,内联查询,例如 Join("project p", "p.project_id=o.project_id")
func (b *Builder) Join(table interface{}, condition []JoinCondition, alias ...string) *Builder {
	b.join("INNER JOIN", table, condition, alias...)
	return b
}

func (b *Builder) join(joinType string, table interface{}, condition []JoinCondition, alias ...string) {
	joinTableAlias := ""
	if len(alias) > 0 {
		joinTableAlias = alias[0]
	}

	b.joinList = append(b.joinList, JoinItem{
		joinType:   joinType,
		table:      table,
		tableAlias: joinTableAlias,
		condition:  condition,
	})
}
