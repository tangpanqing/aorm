package aorm

import (
	"database/sql"
	"errors"
	"fmt"
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
	C Int
}

type FloatStruct struct {
	C Float
}

// Insert 增加记录
func (db *Executor) Insert(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if db.tableName == "" {
		db.tableName = reflectTableName(typeOf, valueOf)
	}

	var keys []string
	var paramList []any
	var place []string
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			keys = append(keys, key)
			paramList = append(paramList, val)
			place = append(place, "?")
		}
	}

	sqlStr := "INSERT INTO " + db.tableName + " (" + strings.Join(keys, ",") + ") VALUES (" + strings.Join(place, ",") + ")"

	res, err := db.Exec(sqlStr, paramList...)
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
func (db *Executor) InsertBatch(values interface{}) (int64, error) {

	var keys []string
	var paramList []any
	var place []string

	valueOf := reflect.ValueOf(values).Elem()
	if valueOf.Len() == 0 {
		return 0, errors.New("the data list for insert batch not found")
	}
	typeOf := reflect.TypeOf(values).Elem().Elem()

	//如果没有设置表名
	if db.tableName == "" {
		db.tableName = reflectTableName(typeOf, valueOf.Index(0))
	}

	for j := 0; j < valueOf.Len(); j++ {
		var placeItem []string

		for i := 0; i < valueOf.Index(j).NumField(); i++ {
			isNotNull := valueOf.Index(j).Field(i).Field(0).Field(1).Bool()
			if isNotNull {
				if j == 0 {
					key := UnderLine(typeOf.Field(i).Name)
					keys = append(keys, key)
				}

				val := valueOf.Index(j).Field(i).Field(0).Field(0).Interface()
				paramList = append(paramList, val)
				placeItem = append(placeItem, "?")
			}
		}

		place = append(place, "("+strings.Join(placeItem, ",")+")")
	}

	sqlStr := "INSERT INTO " + db.tableName + " (" + strings.Join(keys, ",") + ") VALUES " + strings.Join(place, ",")

	res, err := db.Exec(sqlStr, paramList...)
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
func (db *Executor) GetRows() (*sql.Rows, error) {
	sqlStr, paramList := db.GetSqlAndParams()

	smt, errSmt := db.linkCommon.Prepare(sqlStr)
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
func (db *Executor) GetMany(values interface{}) error {
	rows, errRows := db.GetRows()
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
func (db *Executor) GetOne(obj interface{}) error {
	db.Limit(0, 1)

	rows, errRows := db.GetRows()
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
func (db *Executor) RawSql(sql string, paramList ...interface{}) *Executor {
	db.sql = sql
	db.paramList = paramList
	return db
}

func (db *Executor) GetSqlAndParams() (string, []interface{}) {
	if db.sql != "" {
		return db.sql, db.paramList
	}

	var paramList []interface{}

	fieldStr, paramList := handleField(db.selectList, db.selectExpList, paramList)
	whereStr, paramList := handleWhere(db.whereList, paramList)
	joinStr := handleJoin(db.joinList)
	groupStr := handleGroup(db.groupList)
	havingStr, paramList := handleHaving(db.havingList, paramList)
	orderStr := handleOrder(db.orderList)
	limitStr, paramList := handleLimit(db.offset, db.pageSize, paramList)
	lockStr := handleLockForUpdate(db.isLockForUpdate)

	sqlStr := "SELECT " + fieldStr + " FROM " + db.tableName + joinStr + whereStr + groupStr + havingStr + orderStr + limitStr + lockStr

	if db.isDebug {
		fmt.Println(sqlStr)
		fmt.Println(paramList...)
	}

	return sqlStr, paramList
}

// Update 更新记录
func (db *Executor) Update(dest interface{}) (int64, error) {
	var paramList []any
	setStr, paramList := db.handleSet(dest, paramList)
	whereStr, paramList := handleWhere(db.whereList, paramList)
	sqlStr := "UPDATE " + db.tableName + setStr + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Delete 删除记录
func (db *Executor) Delete() (int64, error) {
	var paramList []any
	whereStr, paramList := handleWhere(db.whereList, paramList)
	sqlStr := "DELETE FROM " + db.tableName + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Truncate 清空记录
func (db *Executor) Truncate() (int64, error) {
	sqlStr := "TRUNCATE TABLE  " + db.tableName

	return db.ExecAffected(sqlStr)
}

// Count 聚合函数-数量
func (db *Executor) Count(fieldName string) (int64, error) {
	var obj []IntStruct
	err := db.Select("count(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Int64, nil
}

// Sum 聚合函数-合计
func (db *Executor) Sum(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := db.Select("sum(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Avg 聚合函数-平均值
func (db *Executor) Avg(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := db.Select("avg(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Max 聚合函数-最大值
func (db *Executor) Max(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := db.Select("max(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Min 聚合函数-最小值
func (db *Executor) Min(fieldName string) (float64, error) {
	var obj []FloatStruct
	err := db.Select("min(" + fieldName + ") as c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Value 字段值
func (db *Executor) Value(fieldName string, dest interface{}) error {
	db.Select(fieldName).Limit(0, 1)

	rows, errRows := db.GetRows()
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
func (db *Executor) Pluck(fieldName string, values interface{}) error {
	db.Select(fieldName)

	rows, errRows := db.GetRows()
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
func (db *Executor) Increment(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := handleWhere(db.whereList, paramList)
	sqlStr := "UPDATE " + db.tableName + " SET " + fieldName + "=" + fieldName + "+?" + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Decrement 某字段自减
func (db *Executor) Decrement(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := handleWhere(db.whereList, paramList)
	sqlStr := "UPDATE " + db.tableName + " SET " + fieldName + "=" + fieldName + "-?" + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Exec 通用执行-新增,更新,删除
func (db *Executor) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	if db.isDebug {
		fmt.Println(sqlStr)
		fmt.Println(args...)
	}

	smt, err1 := db.linkCommon.Prepare(sqlStr)
	if err1 != nil {
		return nil, err1
	}
	defer smt.Close()

	res, err2 := smt.Exec(args...)
	if err2 != nil {
		return nil, err2
	}

	db.clear()
	return res, nil
}

// ExecAffected 通用执行-更新,删除
func (db *Executor) ExecAffected(sqlStr string, args ...interface{}) (int64, error) {
	res, err := db.Exec(sqlStr, args...)
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
func (db *Executor) Debug(isDebug bool) *Executor {
	db.isDebug = isDebug
	return db
}

// Select 链式操作-查询哪些字段,默认 *
func (db *Executor) Select(fields ...string) *Executor {
	db.selectList = append(db.selectList, fields...)
	return db
}

// SelectCount 链式操作-count(field) as field_new
func (db *Executor) SelectCount(field string, fieldNew string) *Executor {
	db.selectList = append(db.selectList, "count("+field+") AS "+fieldNew)
	return db
}

// SelectSum 链式操作-sum(field) as field_new
func (db *Executor) SelectSum(field string, fieldNew string) *Executor {
	db.selectList = append(db.selectList, "sum("+field+") AS "+fieldNew)
	return db
}

// SelectMin 链式操作-min(field) as field_new
func (db *Executor) SelectMin(field string, fieldNew string) *Executor {
	db.selectList = append(db.selectList, "min("+field+") AS "+fieldNew)
	return db
}

// SelectMax 链式操作-max(field) as field_new
func (db *Executor) SelectMax(field string, fieldNew string) *Executor {
	db.selectList = append(db.selectList, "max("+field+") AS "+fieldNew)
	return db
}

// SelectAvg 链式操作-avg(field) as field_new
func (db *Executor) SelectAvg(field string, fieldNew string) *Executor {
	db.selectList = append(db.selectList, "avg("+field+") AS "+fieldNew)
	return db
}

// SelectExp 链式操作-表达式
func (db *Executor) SelectExp(dbSub **Executor, fieldName string) *Executor {
	db.selectExpList = append(db.selectExpList, &ExpItem{
		Executor:  dbSub,
		FieldName: fieldName,
	})
	return db
}

// Table 链式操作-从哪个表查询,允许直接写别名,例如 person p
func (db *Executor) Table(tableName string) *Executor {
	db.tableName = tableName
	return db
}

// LeftJoin 链式操作,左联查询,例如 LeftJoin("project p", "p.project_id=o.project_id")
func (db *Executor) LeftJoin(tableName string, condition string) *Executor {
	db.joinList = append(db.joinList, "LEFT JOIN "+tableName+" ON "+condition)
	return db
}

// RightJoin 链式操作,右联查询,例如 RightJoin("project p", "p.project_id=o.project_id")
func (db *Executor) RightJoin(tableName string, condition string) *Executor {
	db.joinList = append(db.joinList, "RIGHT JOIN "+tableName+" ON "+condition)
	return db
}

// Join 链式操作,内联查询,例如 Join("project p", "p.project_id=o.project_id")
func (db *Executor) Join(tableName string, condition string) *Executor {
	db.joinList = append(db.joinList, "INNER JOIN "+tableName+" ON "+condition)
	return db
}

// Where 链式操作,以对象作为查询条件
func (db *Executor) Where(dest interface{}) *Executor {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if db.tableName == "" {
		db.tableName = reflectTableName(typeOf, valueOf)
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			db.whereList = append(db.whereList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return db
}

// WhereArr 链式操作,以数组作为查询条件
func (db *Executor) WhereArr(whereList []WhereItem) *Executor {
	db.whereList = append(db.whereList, whereList...)
	return db
}

func (db *Executor) WhereEq(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Eq,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereNe(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Ne,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereGt(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Gt,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereGe(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Ge,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereLt(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Lt,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereLe(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Le,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereIn(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   In,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereNotIn(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   NotIn,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereBetween(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Between,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereNotBetween(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   NotBetween,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereLike(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Like,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereNotLike(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   NotLike,
		Val:   val,
	})
	return db
}

func (db *Executor) WhereRaw(field string, val interface{}) *Executor {
	db.whereList = append(db.whereList, WhereItem{
		Field: field,
		Opt:   Raw,
		Val:   val,
	})
	return db
}

// GroupBy 链式操作,以某字段进行分组
func (db *Executor) GroupBy(fieldName string) *Executor {
	db.groupList = append(db.groupList, fieldName)
	return db
}

// Having 链式操作,以对象作为筛选条件
func (db *Executor) Having(dest interface{}) *Executor {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if db.tableName == "" {
		db.tableName = reflectTableName(typeOf, valueOf)
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			db.havingList = append(db.havingList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return db
}

// HavingArr 链式操作,以数组作为筛选条件
func (db *Executor) HavingArr(havingList []WhereItem) *Executor {
	db.havingList = append(db.havingList, havingList...)
	return db
}

func (db *Executor) HavingEq(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Eq,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingNe(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Ne,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingGt(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Gt,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingGe(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Ge,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingLt(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Lt,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingLe(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Le,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingIn(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   In,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingNotIn(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   NotIn,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingBetween(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Between,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingNotBetween(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   NotBetween,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingLike(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Like,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingNotLike(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   NotLike,
		Val:   val,
	})
	return db
}

func (db *Executor) HavingRaw(field string, val interface{}) *Executor {
	db.havingList = append(db.havingList, WhereItem{
		Field: field,
		Opt:   Raw,
		Val:   val,
	})
	return db
}

// OrderBy 链式操作,以某字段进行排序
func (db *Executor) OrderBy(field string, orderType string) *Executor {
	db.orderList = append(db.orderList, field+" "+orderType)
	return db
}

// Limit 链式操作,分页
func (db *Executor) Limit(offset int, pageSize int) *Executor {
	db.offset = offset
	db.pageSize = pageSize
	return db
}

// Page 链式操作,分页
func (db *Executor) Page(pageNum int, pageSize int) *Executor {
	db.offset = (pageNum - 1) * pageSize
	db.pageSize = pageSize
	return db
}

// LockForUpdate 加锁
func (db *Executor) LockForUpdate(isLockForUpdate bool) *Executor {
	db.isLockForUpdate = isLockForUpdate
	return db
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
func (db *Executor) handleSet(dest interface{}, paramList []any) (string, []any) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if db.tableName == "" {
		db.tableName = reflectTableName(typeOf, valueOf)
	}

	var keys []string
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := UnderLine(typeOf.Elem().Field(i).Name)
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
		if "**aorm.Executor" == reflect.TypeOf(where[i].Val).String() {
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
		return UnderLine(arr[len(arr)-1])
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
		fieldName := CamelString(strings.ToLower(columnName))
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
