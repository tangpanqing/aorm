package aorm

import (
	"database/sql"
	"errors"
	"fmt"
	"gopkg.in/guregu/null.v4"
	"reflect"
	"strconv"
	"strings"
	"time"
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
func (db *Executor) Insert(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if db.TableName == "" {
		arr := strings.Split(typeOf.String(), ".")
		db.TableName = UnderLine(arr[len(arr)-1])
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

	sqlStr := "INSERT INTO " + db.TableName + " (" + strings.Join(keys, ",") + ") VALUES (" + strings.Join(place, ",") + ")"

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

// GetMany 查询记录
func (db *Executor) GetMany(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	res := db.GetMapArr()
	for i := 0; i < len(res); i++ {
		dest := reflect.New(destType).Elem()
		for k, v := range res[i] {
			fieldName := CamelString(k)
			if dest.FieldByName(fieldName).CanSet() {
				filedType := dest.FieldByName(fieldName).Type().String()
				x := transToNullType(v, filedType)
				dest.FieldByName(fieldName).Set(x)
			}
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}

	return nil
}

// GetOne 查询某一条记录
func (db *Executor) GetOne(obj interface{}) error {

	dest := reflect.ValueOf(obj).Elem()
	res := db.Limit(0, 1).GetMapArr()
	if len(res) == 0 {
		return errors.New("record not found")
	}

	for k, v := range res[0] {
		fieldName := CamelString(k)
		if dest.FieldByName(fieldName).CanSet() {
			filedType := dest.FieldByName(fieldName).Type().String()
			x := transToNullType(v, filedType)
			dest.FieldByName(fieldName).Set(x)
		}
	}

	return nil
}

func (db *Executor) GetMapArr() []map[string]interface{} {
	var paramList []any
	fieldStr := handleField(db.SelectList)
	whereStr, paramList := handleWhere(db.WhereList, paramList)
	joinStr := handleJoin(db.JoinList)
	groupStr := handleGroup(db.GroupList)
	havingStr, paramList := handleHaving(db.HavingList, paramList)
	orderStr := handleOrder(db.OrderList)
	limitStr, paramList := handleLimit(db.Offset, db.PageSize, paramList)
	lockStr := handleLockForUpdate(db.IsLockForUpdate)

	sqlStr := "SELECT " + fieldStr + " FROM " + db.TableName + joinStr + whereStr + groupStr + havingStr + orderStr + limitStr + lockStr
	res, _ := db.Query(sqlStr, paramList...)

	return res
}

func transToNullType(v interface{}, filedType string) reflect.Value {
	x := reflect.ValueOf("")
	if "null.String" == filedType {
		if nil == v {
			x = reflect.ValueOf(null.String{})
		} else {
			x = reflect.ValueOf(null.StringFrom(fmt.Sprintf("%v", v)))
		}
	} else if "null.Int" == filedType {
		if nil == v {
			x = reflect.ValueOf(null.Int{})
		} else {
			int64Val, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
			x = reflect.ValueOf(null.IntFrom(int64Val))
		}
	} else if "null.Time" == filedType {
		if nil == v {
			x = reflect.ValueOf(null.Time{})
		} else {
			timeStr := fmt.Sprintf("%v", v)
			timeArr := strings.Split(timeStr, " ")
			timeArr1 := strings.Split(timeArr[0], "-")
			timeArr2 := strings.Split(timeArr[1], ":")

			a := time.Date(
				str2Int(timeArr1[0]), time.Month(str2Int(timeArr1[1])), str2Int(timeArr1[2]),
				str2Int(timeArr2[0]),
				str2Int(timeArr2[1]),
				str2Int(timeArr2[2]),
				0,
				time.Local,
			)
			x = reflect.ValueOf(null.TimeFrom(a))
		}
	} else if "null.Bool" == filedType {
		if nil == v {
			x = reflect.ValueOf(null.Bool{})
		} else {
			boolVal, _ := strconv.ParseBool(fmt.Sprintf("%v", v))
			x = reflect.ValueOf(null.BoolFrom(boolVal))
		}
	} else if "null.Float" == filedType {
		if nil == v {
			x = reflect.ValueOf(null.Float{})
		} else {
			float64Val, _ := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
			x = reflect.ValueOf(null.FloatFrom(float64Val))
		}
	} else {
		panic("不受支持的类型转换" + filedType)
	}

	return x
}

// Update 更新记录
func (db *Executor) Update(dest interface{}) (int64, error) {
	var paramList []any
	setStr, paramList := db.handleSet(dest, paramList)
	whereStr, paramList := handleWhere(db.WhereList, paramList)
	sqlStr := "UPDATE " + db.TableName + setStr + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Delete 删除记录
func (db *Executor) Delete() (int64, error) {
	var paramList []any
	whereStr, paramList := handleWhere(db.WhereList, paramList)
	sqlStr := "DELETE FROM " + db.TableName + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Truncate 清空记录
func (db *Executor) Truncate() (int64, error) {
	sqlStr := "TRUNCATE TABLE  " + db.TableName

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

// Value 字段值,注意返回值类型为string
func (db *Executor) Value(fieldName string) (string, error) {
	obj := db.Select(fieldName).Limit(0, 1).GetMapArr()
	if len(obj) == 0 {
		return "", errors.New("找不到值")
	}
	return obj[0][fieldName].(string), nil
}

// ValueInt64 字段值,注意返回值类型为int64
func (db *Executor) ValueInt64(fieldName string) (int64, error) {
	obj := db.Select(fieldName).Limit(0, 1).GetMapArr()
	if len(obj) == 0 {
		return 0, errors.New("找不到值")
	}
	return str2Int64(obj[0][fieldName].(string)), nil
}

// ValueFloat32 字段值,注意返回值类型为float32
func (db *Executor) ValueFloat32(fieldName string) (float32, error) {
	obj := db.Select(fieldName).Limit(0, 1).GetMapArr()
	if len(obj) == 0 {
		return 0, errors.New("找不到值")
	}
	return obj[0][fieldName].(float32), nil
}

// ValueFloat64 字段值,注意返回值类型为float32
func (db *Executor) ValueFloat64(fieldName string) (float64, error) {
	obj := db.Select(fieldName).Limit(0, 1).GetMapArr()
	if len(obj) == 0 {
		return 0, errors.New("找不到值")
	}
	return obj[0][fieldName].(float64), nil
}

// Increment 某字段自增
func (db *Executor) Increment(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := handleWhere(db.WhereList, paramList)
	sqlStr := "UPDATE " + db.TableName + " SET " + fieldName + "=" + fieldName + "+?" + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Decrement 某字段自减
func (db *Executor) Decrement(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := handleWhere(db.WhereList, paramList)
	sqlStr := "UPDATE " + db.TableName + " SET " + fieldName + "=" + fieldName + "-?" + whereStr

	return db.ExecAffected(sqlStr, paramList...)
}

// Query 通用查询
func (db *Executor) Query(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	if db.IsDebug {
		fmt.Println(sqlStr)
		fmt.Println(args...)
	}

	var listData []map[string]interface{}

	smt, err1 := db.LinkCommon.Prepare(sqlStr)
	if err1 != nil {
		return listData, err1
	}

	rows, err2 := smt.Query(args...)
	if err2 != nil {
		return listData, err2
	}

	fieldsTypes, _ := rows.ColumnTypes()
	fields, _ := rows.Columns()

	for rows.Next() {

		data := make(map[string]interface{})

		scans := make([]interface{}, len(fields))
		for i := range scans {
			scans[i] = &scans[i]
		}
		err := rows.Scan(scans...)
		if err != nil {
			return nil, err
		}

		for i, v := range scans {
			if v == nil {
				data[fields[i]] = v
			} else {
				if fieldsTypes[i].DatabaseTypeName() == "VARCHAR" || fieldsTypes[i].DatabaseTypeName() == "TEXT" || fieldsTypes[i].DatabaseTypeName() == "CHAR" || fieldsTypes[i].DatabaseTypeName() == "LONGTEXT" {
					data[fields[i]] = fmt.Sprintf("%s", v)
				} else if fieldsTypes[i].DatabaseTypeName() == "INT" || fieldsTypes[i].DatabaseTypeName() == "BIGINT" || fieldsTypes[i].DatabaseTypeName() == "UNSIGNED INT" || fieldsTypes[i].DatabaseTypeName() == "UNSIGNED BIGINT" {
					data[fields[i]] = fmt.Sprintf("%v", v)
				} else if fieldsTypes[i].DatabaseTypeName() == "DECIMAL" {
					data[fields[i]] = string(v.([]uint8))
				} else {
					data[fields[i]] = v
				}
			}
		}

		listData = append(listData, data)
	}

	db.clear()
	return listData, nil
}

// Exec 通用执行-新增,更新,删除
func (db *Executor) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	if db.IsDebug {
		fmt.Println(sqlStr)
		fmt.Println(args...)
	}

	smt, err1 := db.LinkCommon.Prepare(sqlStr)
	if err1 != nil {
		return nil, err1
	}

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
	db.IsDebug = isDebug
	return db
}

// Select 链式操作-查询哪些字段,默认 *
func (db *Executor) Select(f string) *Executor {
	db.SelectList = append(db.SelectList, f)
	return db
}

// Table 链式操作-从哪个表查询,允许直接写别名,例如 person p
func (db *Executor) Table(tableName string) *Executor {
	db.TableName = tableName
	return db
}

// LeftJoin 链式操作,左联查询,例如 LeftJoin("project p", "p.project_id=o.project_id")
func (db *Executor) LeftJoin(tableName string, condition string) *Executor {
	db.JoinList = append(db.JoinList, "LEFT JOIN "+tableName+" ON "+condition)
	return db
}

// RightJoin 链式操作,右联查询,例如 RightJoin("project p", "p.project_id=o.project_id")
func (db *Executor) RightJoin(tableName string, condition string) *Executor {
	db.JoinList = append(db.JoinList, "RIGHT JOIN "+tableName+" ON "+condition)
	return db
}

// Join 链式操作,内联查询,例如 Join("project p", "p.project_id=o.project_id")
func (db *Executor) Join(tableName string, condition string) *Executor {
	db.JoinList = append(db.JoinList, "INNER JOIN "+tableName+" ON "+condition)
	return db
}

// Where 链式操作,以对象作为查询条件
func (db *Executor) Where(dest interface{}) *Executor {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if db.TableName == "" {
		arr := strings.Split(typeOf.String(), ".")
		db.TableName = UnderLine(arr[len(arr)-1])
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			db.WhereList = append(db.WhereList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return db
}

// WhereArr 链式操作,以数组作为查询条件
func (db *Executor) WhereArr(whereList []WhereItem) *Executor {
	db.WhereList = append(db.WhereList, whereList...)
	return db
}

// GroupBy 链式操作,以某字段进行分组
func (db *Executor) GroupBy(f string) *Executor {
	db.GroupList = append(db.GroupList, f)
	return db
}

// Having 链式操作,以对象作为筛选条件
func (db *Executor) Having(dest interface{}) *Executor {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if db.TableName == "" {
		arr := strings.Split(typeOf.String(), ".")
		db.TableName = UnderLine(arr[len(arr)-1])
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			db.HavingList = append(db.HavingList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return db
}

// HavingArr 链式操作,以数组作为筛选条件
func (db *Executor) HavingArr(havingList []WhereItem) *Executor {
	db.HavingList = havingList
	return db
}

// OrderBy 链式操作,以某字段进行排序
func (db *Executor) OrderBy(field string, orderType string) *Executor {
	db.OrderList = append(db.OrderList, field+" "+orderType)
	return db
}

// Limit 链式操作,分页
func (db *Executor) Limit(offset int, pageSize int) *Executor {
	db.Offset = offset
	db.PageSize = pageSize
	return db
}

// Page 链式操作,分页
func (db *Executor) Page(pageNum int, pageSize int) *Executor {
	db.Offset = (pageNum - 1) * pageSize
	db.PageSize = pageSize
	return db
}

// LockForUpdate 加锁
func (db *Executor) LockForUpdate(isLockForUpdate bool) *Executor {
	db.IsLockForUpdate = isLockForUpdate
	return db
}

//拼接SQL,字段相关
func handleField(selectList []string) string {
	if len(selectList) == 0 {
		return "*"
	}

	return strings.Join(selectList, ",")
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
	if db.TableName == "" {
		arr := strings.Split(typeOf.String(), ".")
		db.TableName = UnderLine(arr[len(arr)-1])
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

func str2Int(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func str2Int64(str string) int64 {
	dataNew, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return dataNew
}
