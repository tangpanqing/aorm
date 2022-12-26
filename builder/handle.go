package builder

import (
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/model"
	"reflect"
	"strings"
)

//拼接SQL,字段相关
func handleField(selectList []string, selectExpList []*SelectItem, paramList []any) (string, []any) {
	if len(selectList) == 0 && len(selectExpList) == 0 {
		return "*", paramList
	}

	//处理子语句
	for i := 0; i < len(selectExpList); i++ {
		executor := *(selectExpList[i].Executor)
		subSql, subParamList := executor.GetSqlAndParams()
		selectList = append(selectList, "("+subSql+") AS "+selectExpList[i].FieldName)
		paramList = append(paramList, subParamList...)
	}

	return strings.Join(selectList, ","), paramList
}

//拼接SQL,查询条件
func (ex *Builder) handleWhere(where []WhereItem, paramList []any) (string, []any) {
	if len(where) == 0 {
		return "", paramList
	}

	whereList, paramList := ex.whereAndHaving(where, paramList)

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
func handleJoin(joinList []string) string {
	if len(joinList) == 0 {
		return ""
	}

	return " " + strings.Join(joinList, " ")
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
func handleOrder(orderList []string) string {
	if len(orderList) == 0 {
		return ""
	}

	return " Order BY " + strings.Join(orderList, ",")
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
