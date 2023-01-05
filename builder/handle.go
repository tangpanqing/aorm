package builder

import (
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/model"
	"reflect"
	"strings"
)

//拼接SQL,字段相关
func (ex *Builder) handleField(paramList []any) (string, []any) {
	if len(ex.selectList) == 0 && len(ex.selectExpList) == 0 {
		return "*", paramList
	}

	//处理子语句
	//for i := 0; i < len(selectExpList); i++ {
	//	executor := *(selectExpList[i].Executor)
	//	subSql, subParamList := executor.GetSqlAndParams()
	//	selectList = append(selectList, "("+subSql+") AS "+selectExpList[i].FieldName)
	//	paramList = append(paramList, subParamList...)
	//}
	var strList []string

	for i := 0; i < len(ex.selectList); i++ {
		selectItem := ex.selectList[i]

		str := ""
		if selectItem.FuncName != "" {
			str += selectItem.FuncName
			str += "("
		}

		if selectItem.Prefix != "" {
			str += selectItem.Prefix
			str += "."
		}

		str += getFieldName(selectItem.Field)

		if selectItem.FuncName != "" {
			str += ")"
		}

		if selectItem.FieldNew != nil {
			str += " AS "
			str += getFieldName(selectItem.FieldNew)
		}

		strList = append(strList, str)
	}

	return strings.Join(strList, ","), paramList
}

//拼接SQL,查询条件
func (ex *Builder) handleWhere(paramList []any) (string, []any) {
	if len(ex.whereList) == 0 {
		return "", paramList
	}

	whereList, paramList := ex.whereAndHaving(ex.whereList, paramList)

	return " WHERE " + strings.Join(whereList, " AND "), paramList
}

//拼接SQL,更新信息
func (ex *Builder) handleSet(dest interface{}, paramList []any) (string, []any) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = getTableName(typeOf, valueOf)
	}

	var keys []string
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := helper.UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()

			keys = append(keys, key+"=?")
			paramList = append(paramList, val)
		}
	}

	return " SET " + strings.Join(keys, ","), paramList
}

//拼接SQL,关联查询
func (b *Builder) handleJoin(paramList []interface{}) (string, []interface{}) {
	if len(b.joinList) == 0 {
		return "", paramList
	}

	var sqlList []string
	for i := 0; i < len(b.joinList); i++ {
		joinItem := b.joinList[i]

		str, paramList2 := getWhereStrForJoin(joinItem.tableAlias, joinItem.condition, paramList)
		paramList = paramList2

		sqlItem := joinItem.joinType + " " + getTableNameByTable(joinItem.table) + " " + joinItem.tableAlias + " ON " + str
		sqlList = append(sqlList, sqlItem)
	}

	return " " + strings.Join(sqlList, " "), paramList
}

//拼接SQL,结果分组
func handleGroup(groupList []string) string {
	if len(groupList) == 0 {
		return ""
	}

	return " GROUP BY " + strings.Join(groupList, ",")
}

//拼接SQL,结果筛选
func (ex *Builder) handleHaving(having []WhereItem, paramList []any) (string, []any) {
	if len(having) == 0 {
		return "", paramList
	}

	whereList, paramList := ex.whereAndHaving(having, paramList)

	return " Having " + strings.Join(whereList, " AND "), paramList
}

//拼接SQL,结果排序
func (ex *Builder) handleOrder(paramList []any) (string, []any) {
	if len(ex.orderList) == 0 {
		return "", paramList
	}

	var orderList []string
	for i := 0; i < len(ex.orderList); i++ {
		orderList = append(orderList, ex.orderList[i].Prefix+"."+getFieldName(ex.orderList[i].Field)+" "+ex.orderList[i].OrderType)
	}

	return " ORDER BY " + strings.Join(orderList, ","), paramList
}

//拼接SQL,分页相关  Postgres数据库分页数量在前偏移在后，其他数据库偏移量在前分页数量在后，另外Mssql数据库的关键词是offset...next
func (ex *Builder) handleLimit(offset int, pageSize int, paramList []any) (string, []any) {
	if 0 == pageSize {
		return "", paramList
	}

	str := ""
	if ex.driverName == model.Postgres {
		paramList = append(paramList, pageSize)
		paramList = append(paramList, offset)

		str = " Limit ? offset ? "
	} else {
		paramList = append(paramList, offset)
		paramList = append(paramList, pageSize)

		str = " Limit ?,? "
		if ex.driverName == model.Mssql {
			str = " offset ? rows fetch next ? rows only "
		}
	}

	return str, paramList
}

//拼接SQL,锁
func handleLockForUpdate(isLock bool) string {
	if isLock {
		return " FOR UPDATE"
	}

	return ""
}
