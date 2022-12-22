package executor

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/null"
	"reflect"
	"strings"
	"unsafe"
)

const Desc = "DESC"
const Asc = "ASC"

const Eq = "="
const Ne = "!="
const Gt = ">"
const Ge = ">="
const Lt = "<"
const Le = "<="

const In = "IN"
const NotIn = "NOT IN"
const Like = "LIKE"
const NotLike = "NOT LIKE"
const Between = "BETWEEN"
const NotBetween = "NOT BETWEEN"

const Raw = "Raw"

type WhereItem struct {
	Field string
	Opt   string
	Val   any
}

type IntStruct struct {
	C null.Int
}

type FloatStruct struct {
	C null.Float
}

// Insert 增加记录
func (ex *Executor) Insert(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = reflectTableName(typeOf, valueOf)
	}

	var keys []string
	var paramList []any
	var place []string
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := helper.UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			keys = append(keys, key)
			paramList = append(paramList, val)
			place = append(place, "?")
		}
	}

	sqlStr := "INSERT INTO " + ex.tableName + " (" + strings.Join(keys, ",") + ") VALUES (" + strings.Join(place, ",") + ")"

	res, err := ex.Exec(sqlStr, paramList...)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastId, nil
}

// InsertBatch 批量增加记录
func (ex *Executor) InsertBatch(values interface{}) (int64, error) {

	var keys []string
	var paramList []any
	var place []string

	valueOf := reflect.ValueOf(values).Elem()
	if valueOf.Len() == 0 {
		return 0, errors.New("the data list for insert batch not found")
	}
	typeOf := reflect.TypeOf(values).Elem().Elem()

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = reflectTableName(typeOf, valueOf.Index(0))
	}

	for j := 0; j < valueOf.Len(); j++ {
		var placeItem []string

		for i := 0; i < valueOf.Index(j).NumField(); i++ {
			isNotNull := valueOf.Index(j).Field(i).Field(0).Field(1).Bool()
			if isNotNull {
				if j == 0 {
					key := helper.UnderLine(typeOf.Field(i).Name)
					keys = append(keys, key)
				}

				val := valueOf.Index(j).Field(i).Field(0).Field(0).Interface()
				paramList = append(paramList, val)
				placeItem = append(placeItem, "?")
			}
		}

		place = append(place, "("+strings.Join(placeItem, ",")+")")
	}

	sqlStr := "INSERT INTO " + ex.tableName + " (" + strings.Join(keys, ",") + ") VALUES " + strings.Join(place, ",")

	res, err := ex.Exec(sqlStr, paramList...)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetRows 获取行操作
func (ex *Executor) GetRows() (*sql.Rows, error) {
	sqlStr, paramList := ex.GetSqlAndParams()

	smt, errSmt := ex.LinkCommon.Prepare(sqlStr)
	if errSmt != nil {
		return nil, errSmt
	}
	//defer smt.Close()

	rows, errRows := smt.Query(paramList...)
	if errRows != nil {
		return nil, errRows
	}

	return rows, nil
}

// GetMany 查询记录(新)
func (ex *Executor) GetMany(values interface{}) error {
	rows, errRows := ex.GetRows()
	defer rows.Close()
	if errRows != nil {
		return errRows
	}

	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	destValue := reflect.New(destType).Elem()

	//从数据库中读出来的字段名字
	columnNameList, errColumns := rows.Columns()
	if errColumns != nil {
		return errColumns
	}

	//从结构体反射出来的属性名
	fieldNameMap := getFieldNameMap(destValue, destType)

	for rows.Next() {
		scans := getScans(columnNameList, fieldNameMap, destValue)

		errScan := rows.Scan(scans...)
		if errScan != nil {
			return errScan
		}

		destSlice.Set(reflect.Append(destSlice, destValue))
	}

	return nil
}

// GetOne 查询某一条记录
func (ex *Executor) GetOne(obj interface{}) error {
	ex.Limit(0, 1)

	rows, errRows := ex.GetRows()
	defer rows.Close()
	if errRows != nil {
		return errRows
	}

	destType := reflect.TypeOf(obj).Elem()
	destValue := reflect.ValueOf(obj).Elem()

	//从数据库中读出来的字段名字
	columnNameList, errColumns := rows.Columns()
	if errColumns != nil {
		return errColumns
	}

	//从结构体反射出来的属性名
	fieldNameMap := getFieldNameMap(destValue, destType)

	for rows.Next() {
		scans := getScans(columnNameList, fieldNameMap, destValue)
		err := rows.Scan(scans...)
		if err != nil {
			return err
		}
	}

	return nil
}

// RawSql 执行原始的sql语句
func (ex *Executor) RawSql(sql string, paramList ...interface{}) *Executor {
	ex.sql = sql
	ex.paramList = paramList
	return ex
}

func (ex *Executor) GetSqlAndParams() (string, []interface{}) {
	if ex.sql != "" {
		return ex.sql, ex.paramList
	}

	var paramList []interface{}

	fieldStr, paramList := handleField(ex.selectList, ex.selectExpList, paramList)
	whereStr, paramList := handleWhere(ex.whereList, paramList)
	joinStr := handleJoin(ex.joinList)
	groupStr := handleGroup(ex.groupList)
	havingStr, paramList := handleHaving(ex.havingList, paramList)
	orderStr := handleOrder(ex.orderList)
	limitStr, paramList := handleLimit(ex.offset, ex.pageSize, paramList)
	lockStr := handleLockForUpdate(ex.isLockForUpdate)

	sqlStr := "SELECT " + fieldStr + " FROM " + ex.tableName + joinStr + whereStr + groupStr + havingStr + orderStr + limitStr + lockStr

	if ex.isDebug {
		fmt.Println(sqlStr)
		fmt.Println(paramList...)
	}

	return sqlStr, paramList
}

// Update 更新记录
func (ex *Executor) Update(dest interface{}) (int64, error) {
	var paramList []any
	setStr, paramList := ex.handleSet(dest, paramList)
	whereStr, paramList := handleWhere(ex.whereList, paramList)
	sqlStr := "UPDATE " + ex.tableName + setStr + whereStr

	return ex.ExecAffected(sqlStr, paramList...)
}

// Delete 删除记录
func (ex *Executor) Delete() (int64, error) {
	var paramList []any
	whereStr, paramList := handleWhere(ex.whereList, paramList)
	sqlStr := "DELETE FROM " + ex.tableName + whereStr

	return ex.ExecAffected(sqlStr, paramList...)
}

// Truncate 清空记录
func (ex *Executor) Truncate() (int64, error) {
	sqlStr := "TRUNCATE TABLE  " + ex.tableName

	return ex.ExecAffected(sqlStr)
}

// Count 聚合函数-数量
func (ex *Executor) Count(fieldName string) (int64, error) {
	var obj []IntStruct
	err := ex.Select("count(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Int64, nil
}

// Sum 聚合函数-合计
func (ex *Executor) Sum(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := ex.Select("sum(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Avg 聚合函数-平均值
func (ex *Executor) Avg(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := ex.Select("avg(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Max 聚合函数-最大值
func (ex *Executor) Max(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := ex.Select("max(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Min 聚合函数-最小值
func (ex *Executor) Min(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := ex.Select("min(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Value 字段值
func (ex *Executor) Value(fieldName string, dest interface{}) error {
	ex.Select(fieldName).Limit(0, 1)

	rows, errRows := ex.GetRows()
	defer rows.Close()
	if errRows != nil {
		return errRows
	}

	destValue := reflect.ValueOf(dest).Elem()

	//从数据库中读出来的字段名字
	columnNameList, errColumns := rows.Columns()
	if errColumns != nil {
		return errColumns
	}

	for rows.Next() {
		var scans []interface{}
		for _, columnName := range columnNameList {
			if fieldName == columnName {
				scans = append(scans, destValue.Addr().Interface())
			} else {
				var emptyVal interface{}
				scans = append(scans, &emptyVal)
			}
		}

		err := rows.Scan(scans...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Pluck 获取某一列的值
func (ex *Executor) Pluck(fieldName string, values interface{}) error {
	ex.Select(fieldName)

	rows, errRows := ex.GetRows()
	defer rows.Close()
	if errRows != nil {
		return errRows
	}

	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	destValue := reflect.New(destType).Elem()

	//从数据库中读出来的字段名字
	columnNameList, errColumns := rows.Columns()
	if errColumns != nil {
		return errColumns
	}

	for rows.Next() {
		var scans []interface{}
		for _, columnName := range columnNameList {
			if fieldName == columnName {
				scans = append(scans, destValue.Addr().Interface())
			} else {
				var emptyVal interface{}
				scans = append(scans, &emptyVal)
			}
		}

		err := rows.Scan(scans...)
		if err != nil {
			return err
		}

		destSlice.Set(reflect.Append(destSlice, destValue))
	}

	return nil
}

// Increment 某字段自增
func (ex *Executor) Increment(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := handleWhere(ex.whereList, paramList)
	sqlStr := "UPDATE " + ex.tableName + " SET " + fieldName + "=" + fieldName + "+?" + whereStr

	return ex.ExecAffected(sqlStr, paramList...)
}

// Decrement 某字段自减
func (ex *Executor) Decrement(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := handleWhere(ex.whereList, paramList)
	sqlStr := "UPDATE " + ex.tableName + " SET " + fieldName + "=" + fieldName + "-?" + whereStr

	return ex.ExecAffected(sqlStr, paramList...)
}

// Exec 通用执行-新增,更新,删除
func (ex *Executor) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	if ex.isDebug {
		fmt.Println(sqlStr)
		fmt.Println(args...)
	}

	smt, err1 := ex.LinkCommon.Prepare(sqlStr)
	if err1 != nil {
		return nil, err1
	}
	defer smt.Close()

	res, err2 := smt.Exec(args...)
	if err2 != nil {
		return nil, err2
	}

	//ex.clear()
	return res, nil
}

// ExecAffected 通用执行-更新,删除
func (ex *Executor) ExecAffected(sqlStr string, args ...interface{}) (int64, error) {
	res, err := ex.Exec(sqlStr, args...)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Debug 链式操作-是否开启调试,打印sql
func (ex *Executor) Debug(isDebug bool) *Executor {
	ex.isDebug = isDebug
	return ex
}

// Select 链式操作-查询哪些字段,默认 *
func (ex *Executor) Select(fields ...string) *Executor {
	ex.selectList = append(ex.selectList, fields...)
	return ex
}

// SelectCount 链式操作-count(field) as field_new
func (ex *Executor) SelectCount(field string, fieldNew string) *Executor {
	ex.selectList = append(ex.selectList, "count("+field+") AS "+fieldNew)
	return ex
}

// SelectSum 链式操作-sum(field) as field_new
func (ex *Executor) SelectSum(field string, fieldNew string) *Executor {
	ex.selectList = append(ex.selectList, "sum("+field+") AS "+fieldNew)
	return ex
}

// SelectMin 链式操作-min(field) as field_new
func (ex *Executor) SelectMin(field string, fieldNew string) *Executor {
	ex.selectList = append(ex.selectList, "min("+field+") AS "+fieldNew)
	return ex
}

// SelectMax 链式操作-max(field) as field_new
func (ex *Executor) SelectMax(field string, fieldNew string) *Executor {
	ex.selectList = append(ex.selectList, "max("+field+") AS "+fieldNew)
	return ex
}

// SelectAvg 链式操作-avg(field) as field_new
func (ex *Executor) SelectAvg(field string, fieldNew string) *Executor {
	ex.selectList = append(ex.selectList, "avg("+field+") AS "+fieldNew)
	return ex
}

// SelectExp 链式操作-表达式
func (ex *Executor) SelectExp(dbSub **Executor, fieldName string) *Executor {
	ex.selectExpList = append(ex.selectExpList, &ExpItem{
		Executor:  dbSub,
		FieldName: fieldName,
	})
	return ex
}

// Table 链式操作-从哪个表查询,允许直接写别名,例如 person p
func (ex *Executor) Table(tableName string) *Executor {
	ex.tableName = tableName
	return ex
}

// LeftJoin 链式操作,左联查询,例如 LeftJoin("project p", "p.project_id=o.project_id")
func (ex *Executor) LeftJoin(tableName string, condition string) *Executor {
	ex.joinList = append(ex.joinList, "LEFT JOIN "+tableName+" ON "+condition)
	return ex
}

// RightJoin 链式操作,右联查询,例如 RightJoin("project p", "p.project_id=o.project_id")
func (ex *Executor) RightJoin(tableName string, condition string) *Executor {
	ex.joinList = append(ex.joinList, "RIGHT JOIN "+tableName+" ON "+condition)
	return ex
}

// Join 链式操作,内联查询,例如 Join("project p", "p.project_id=o.project_id")
func (ex *Executor) Join(tableName string, condition string) *Executor {
	ex.joinList = append(ex.joinList, "INNER JOIN "+tableName+" ON "+condition)
	return ex
}

// Where 链式操作,以对象作为查询条件
func (ex *Executor) Where(dest interface{}) *Executor {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = reflectTableName(typeOf, valueOf)
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := helper.UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			ex.whereList = append(ex.whereList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return ex
}

// WhereArr 链式操作,以数组作为查询条件
func (ex *Executor) WhereArr(whereList []WhereItem) *Executor {
	ex.whereList = append(ex.whereList, whereList...)
	return ex
}

func (ex *Executor) WhereEq(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Eq,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereNe(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Ne,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereGt(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Gt,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereGe(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Ge,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereLt(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Lt,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereLe(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Le,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereIn(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   In,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereNotIn(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   NotIn,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereBetween(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Between,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereNotBetween(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   NotBetween,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereLike(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Like,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereNotLike(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   NotLike,
		Val:   val,
	})
	return ex
}

func (ex *Executor) WhereRaw(field string, val interface{}) *Executor {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Raw,
		Val:   val,
	})
	return ex
}

// GroupBy 链式操作,以某字段进行分组
func (ex *Executor) GroupBy(fieldName string) *Executor {
	ex.groupList = append(ex.groupList, fieldName)
	return ex
}

// Having 链式操作,以对象作为筛选条件
func (ex *Executor) Having(dest interface{}) *Executor {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = reflectTableName(typeOf, valueOf)
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := helper.UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			ex.havingList = append(ex.havingList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return ex
}

// HavingArr 链式操作,以数组作为筛选条件
func (ex *Executor) HavingArr(havingList []WhereItem) *Executor {
	ex.havingList = append(ex.havingList, havingList...)
	return ex
}

func (ex *Executor) HavingEq(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Eq,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingNe(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Ne,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingGt(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Gt,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingGe(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Ge,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingLt(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Lt,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingLe(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Le,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingIn(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   In,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingNotIn(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   NotIn,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingBetween(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Between,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingNotBetween(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   NotBetween,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingLike(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Like,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingNotLike(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   NotLike,
		Val:   val,
	})
	return ex
}

func (ex *Executor) HavingRaw(field string, val interface{}) *Executor {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Raw,
		Val:   val,
	})
	return ex
}

// OrderBy 链式操作,以某字段进行排序
func (ex *Executor) OrderBy(field string, orderType string) *Executor {
	ex.orderList = append(ex.orderList, field+" "+orderType)
	return ex
}

// Limit 链式操作,分页
func (ex *Executor) Limit(offset int, pageSize int) *Executor {
	ex.offset = offset
	ex.pageSize = pageSize
	return ex
}

// Page 链式操作,分页
func (ex *Executor) Page(pageNum int, pageSize int) *Executor {
	ex.offset = (pageNum - 1) * pageSize
	ex.pageSize = pageSize
	return ex
}

// LockForUpdate 加锁
func (ex *Executor) LockForUpdate(isLockForUpdate bool) *Executor {
	ex.isLockForUpdate = isLockForUpdate
	return ex
}

//拼接SQL,字段相关
func handleField(selectList []string, selectExpList []*ExpItem, paramList []any) (string, []any) {
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
func handleWhere(where []WhereItem, paramList []any) (string, []any) {
	if len(where) == 0 {
		return "", paramList
	}

	whereList, paramList := whereAndHaving(where, paramList)

	return " WHERE " + strings.Join(whereList, " AND "), paramList
}

//拼接SQL,更新信息
func (ex *Executor) handleSet(dest interface{}, paramList []any) (string, []any) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = reflectTableName(typeOf, valueOf)
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
func handleHaving(having []WhereItem, paramList []any) (string, []any) {
	if len(having) == 0 {
		return "", paramList
	}

	whereList, paramList := whereAndHaving(having, paramList)

	return " Having " + strings.Join(whereList, " AND "), paramList
}

//拼接SQL,结果排序
func handleOrder(orderList []string) string {
	if len(orderList) == 0 {
		return ""
	}

	return " Order BY " + strings.Join(orderList, ",")
}

//拼接SQL,分页相关
func handleLimit(offset int, pageSize int, paramList []any) (string, []any) {
	if 0 == pageSize {
		return "", paramList
	}

	paramList = append(paramList, offset)
	paramList = append(paramList, pageSize)

	return " Limit ?,? ", paramList
}

//拼接SQL,锁
func handleLockForUpdate(isLock bool) string {
	if isLock {
		return " FOR UPDATE"
	}

	return ""
}

//拼接SQL,查询与筛选通用操作
func whereAndHaving(where []WhereItem, paramList []any) ([]string, []any) {
	var whereList []string
	for i := 0; i < len(where); i++ {
		if "**executor.Executor" == reflect.TypeOf(where[i].Val).String() {
			executor := *(**Executor)(unsafe.Pointer(reflect.ValueOf(where[i].Val).Pointer()))
			subSql, subParams := executor.GetSqlAndParams()

			if where[i].Opt != Raw {
				whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+"("+subSql+")")
				paramList = append(paramList, subParams...)
			} else {

			}
		} else {
			if where[i].Opt == Eq || where[i].Opt == Ne || where[i].Opt == Gt || where[i].Opt == Ge || where[i].Opt == Lt || where[i].Opt == Le {
				//如果是浮点数查询
				switch where[i].Val.(type) {
				case float32:
					whereList = append(whereList, "CONCAT("+where[i].Field+",'') "+where[i].Opt+" "+"?")
				case float64:
					whereList = append(whereList, "CONCAT("+where[i].Field+",'') "+where[i].Opt+" "+"?")
				default:
					whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+"?")
				}

				paramList = append(paramList, fmt.Sprintf("%v", where[i].Val))
			}

			if where[i].Opt == Between || where[i].Opt == NotBetween {
				values := toAnyArr(where[i].Val)
				whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+"(?) AND (?)")
				paramList = append(paramList, values...)
			}

			if where[i].Opt == Like || where[i].Opt == NotLike {
				values := toAnyArr(where[i].Val)
				var valueStr []string
				for j := 0; j < len(values); j++ {
					str := fmt.Sprintf("%v", values[j])

					if "%" != str {
						//values[j] = "?"
						paramList = append(paramList, str)
						valueStr = append(valueStr, "?")
					} else {
						valueStr = append(valueStr, "'"+str+"'")
					}
				}

				whereList = append(whereList, where[i].Field+" "+where[i].Opt+" concat("+strings.Join(valueStr, ",")+")")
			}

			if where[i].Opt == In || where[i].Opt == NotIn {
				values := toAnyArr(where[i].Val)
				var placeholder []string
				for j := 0; j < len(values); j++ {
					placeholder = append(placeholder, "?")
				}

				whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+"("+strings.Join(placeholder, ",")+")")
				paramList = append(paramList, values...)
			}

			if where[i].Opt == Raw {
				whereList = append(whereList, where[i].Field+fmt.Sprintf("%v", where[i].Val))
			}
		}
	}

	return whereList, paramList
}

//将一个interface抽取成数组
func toAnyArr(val any) []any {
	var values []any
	switch val.(type) {
	case []int:
		for _, value := range val.([]int) {
			values = append(values, value)
		}
	case []int64:
		for _, value := range val.([]int64) {
			values = append(values, value)
		}
	case []float32:
		for _, value := range val.([]float32) {
			values = append(values, value)
		}
	case []float64:
		for _, value := range val.([]float64) {
			values = append(values, value)
		}
	case []string:
		for _, value := range val.([]string) {
			values = append(values, value)
		}
	}

	return values
}

//反射表名,优先从方法获取,没有方法则从名字获取
func reflectTableName(typeOf reflect.Type, valueOf reflect.Value) string {
	method, isSet := typeOf.MethodByName("TableName")
	if isSet {
		var paramList []reflect.Value
		paramList = append(paramList, valueOf)
		res := method.Func.Call(paramList)
		return res[0].String()
	} else {
		arr := strings.Split(typeOf.String(), ".")
		return helper.UnderLine(arr[len(arr)-1])
	}
}

func getFieldNameMap(destValue reflect.Value, destType reflect.Type) map[string]int {
	fieldNameMap := make(map[string]int)
	for i := 0; i < destValue.NumField(); i++ {
		fieldNameMap[destType.Field(i).Name] = i
	}

	return fieldNameMap
}

func getScans(columnNameList []string, fieldNameMap map[string]int, destValue reflect.Value) []interface{} {
	var scans []interface{}
	for _, columnName := range columnNameList {
		fieldName := helper.CamelString(strings.ToLower(columnName))
		index, ok := fieldNameMap[fieldName]
		if ok {
			scans = append(scans, destValue.Field(index).Addr().Interface())
		} else {
			var emptyVal interface{}
			scans = append(scans, &emptyVal)
		}
	}

	return scans
}
