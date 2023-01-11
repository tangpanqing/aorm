package builder

import (
	"github.com/tangpanqing/aorm/model"
	"reflect"
	"strings"
)

func handleSelectWith(selectItem SelectItem) string {
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

	return str
}

//拼接SQL,字段相关
func (b *Builder) handleSelect(paramList []any) (string, []any) {
	fieldStr := ""
	if b.distinct {
		fieldStr += "DISTINCT "
	}

	if len(b.selectList) == 0 && len(b.selectExpList) == 0 {
		fieldStr += "*"
		return fieldStr, paramList
	}

	var strList []string

	//处理一般的参数
	for i := 0; i < len(b.selectList); i++ {
		selectItem := b.selectList[i]

		str := handleSelectWith(selectItem)

		if selectItem.FieldNew != nil {
			str += " AS "
			str += getFieldName(selectItem.FieldNew)
		}

		strList = append(strList, str)
	}

	//处理子语句
	for i := 0; i < len(b.selectExpList); i++ {
		subBuilder := *(b.selectExpList[i].Builder)
		subSql, subParamList := subBuilder.GetSqlAndParams()
		strList = append(strList, "("+subSql+") AS "+getFieldName(b.selectExpList[i].FieldName))
		paramList = append(paramList, subParamList...)
	}

	fieldStr += strings.Join(strList, ",")
	return fieldStr, paramList
}

//拼接SQL,查询条件
func (b *Builder) handleWhere(paramList []any) (string, []any) {
	if len(b.whereList) == 0 {
		return "", paramList
	}

	strList, paramList := b.whereAndHaving(b.whereList, paramList, false)

	return " WHERE " + strings.Join(strList, " AND "), paramList
}

//拼接SQL,更新信息
func (b *Builder) handleSet(typeOf reflect.Type, valueOf reflect.Value, paramList []any) (string, []any) {

	//如果没有设置表名
	if b.table == nil {
		b.table = getTableNameByReflect(typeOf, valueOf)
	}

	var keys []string
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key, _ := getFieldNameByReflect(typeOf.Elem().Field(i))

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

		str, paramList2 := genJoinConditionStr(joinItem.tableAlias, joinItem.condition, paramList)
		paramList = paramList2

		sqlItem := joinItem.joinType + " " + getTableNameByTable(joinItem.table) + " " + joinItem.tableAlias + " ON " + str
		sqlList = append(sqlList, sqlItem)
	}

	return " " + strings.Join(sqlList, " "), paramList
}

//拼接SQL,结果分组
func (b *Builder) handleGroup(paramList []any) (string, []any) {
	if len(b.groupList) == 0 {
		return "", paramList
	}

	var groupList []string
	for i := 0; i < len(b.groupList); i++ {
		groupList = append(groupList, b.groupList[i].Prefix+"."+getFieldName(b.groupList[i].Field))
	}

	return " GROUP BY " + strings.Join(groupList, ","), paramList
}

//拼接SQL,结果筛选
func (b *Builder) handleHaving(paramList []any) (string, []any) {
	if len(b.havingList) == 0 {
		return "", paramList
	}

	strList, paramList := b.whereAndHaving(b.havingList, paramList, true)

	return " Having " + strings.Join(strList, " AND "), paramList
}

//拼接SQL,结果排序
func (b *Builder) handleOrder(paramList []any) (string, []any) {
	if len(b.orderList) == 0 {
		return "", paramList
	}

	var orderList []string
	for i := 0; i < len(b.orderList); i++ {
		orderList = append(orderList, b.orderList[i].Prefix+"."+getFieldName(b.orderList[i].Field)+" "+b.orderList[i].OrderType)
	}

	return " ORDER BY " + strings.Join(orderList, ","), paramList
}

//拼接SQL,分页相关  Postgres数据库分页数量在前偏移在后，其他数据库偏移量在前分页数量在后，另外Mssql数据库的关键词是offset...next
func (b *Builder) handleLimit(paramList []any) (string, []any) {
	if 0 == b.limitItem.pageSize {
		return "", paramList
	}

	str := ""
	if b.LinkCommon.DriverName() == model.Postgres {
		paramList = append(paramList, b.limitItem.pageSize)
		paramList = append(paramList, b.limitItem.offset)

		str = " Limit ? offset ? "
	} else {
		paramList = append(paramList, b.limitItem.offset)
		paramList = append(paramList, b.limitItem.pageSize)

		str = " Limit ?,? "
		if b.LinkCommon.DriverName() == model.Mssql {
			str = " offset ? rows fetch next ? rows only "
		}
	}

	return str, paramList
}

//拼接SQL,锁
func (b *Builder) handleLockForUpdate() string {
	if b.isLockForUpdate {
		return " FOR UPDATE"
	}

	return ""
}
