package builder

// LeftJoin 链式操作,左联查询,例如 LeftJoin("project p", "p.project_id=o.project_id")
func (ex *Builder) LeftJoin(tableName string, condition string) *Builder {
	ex.joinList = append(ex.joinList, "LEFT JOIN "+tableName+" ON "+condition)
	return ex
}

// RightJoin 链式操作,右联查询,例如 RightJoin("project p", "p.project_id=o.project_id")
func (ex *Builder) RightJoin(tableName string, condition string) *Builder {
	ex.joinList = append(ex.joinList, "RIGHT JOIN "+tableName+" ON "+condition)
	return ex
}

// Join 链式操作,内联查询,例如 Join("project p", "p.project_id=o.project_id")
func (ex *Builder) Join(tableName string, condition string) *Builder {
	ex.joinList = append(ex.joinList, "INNER JOIN "+tableName+" ON "+condition)
	return ex
}
