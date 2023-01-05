package builder

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/tangpanqing/aorm/helper"
	"github.com/tangpanqing/aorm/model"
	"reflect"
	"strconv"
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

// SelectItem 将某子语句重命名为某字段
type SelectItem struct {
	Executor  **Builder
	FieldName string
}

// Builder 查询记录所需要的条件
type Builder struct {
	//数据库操作连接
	LinkCommon model.LinkCommon

	//查询参数
	tableName       string
	selectList      []string
	selectExpList   []*SelectItem
	groupList       []string
	whereList       []WhereItem
	joinList        []string
	havingList      []WhereItem
	orderList       []string
	offset          int
	pageSize        int
	isDebug         bool
	isLockForUpdate bool

	//sql与参数
	sql       string
	paramList []interface{}

	//驱动名字
	driverName string
}

type WhereItem struct {
	Field string
	Opt   string
	Val   any
}

func (ex *Builder) Driver(driverName string) *Builder {
	ex.driverName = driverName
	return ex
}

// Insert 增加记录
func (ex *Builder) Insert(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = getTableName(typeOf, valueOf)
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

	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

	if ex.driverName == model.Mssql {
		return ex.insertForMssqlOrPostgres(sqlStr+"; select ID = convert(bigint, SCOPE_IDENTITY())", paramList...)
	} else if ex.driverName == model.Postgres {
		return ex.insertForMssqlOrPostgres(sqlStr+" returning id", paramList...)
	} else {
		return ex.insertForCommon(sqlStr, paramList...)
	}
}

//对于Mssql,Postgres类型数据库，为了获取最后插入的id，需要改写入为查询
func (ex *Builder) insertForMssqlOrPostgres(sql string, paramList ...any) (int64, error) {
	rows, err := ex.LinkCommon.Query(sql, paramList...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var lastInsertId1 int64
	for rows.Next() {
		rows.Scan(&lastInsertId1)
	}
	return lastInsertId1, nil
}

//对于非Mssql,Postgres类型数据库，可以直接获取最后插入的id
func (ex *Builder) insertForCommon(sql string, paramList ...any) (int64, error) {
	res, err := ex.Exec(sql, paramList...)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastId, nil
}

//对于Postgres数据库，不支持?占位符，支持$1,$2类型，需要做转换
func convertToPostgresSql(sqlStr string) string {
	t := 1
	for {
		if strings.Index(sqlStr, "?") == -1 {
			break
		}
		sqlStr = strings.Replace(sqlStr, "?", "$"+strconv.Itoa(t), 1)
		t += 1
	}

	return sqlStr
}

// InsertBatch 批量增加记录
func (ex *Builder) InsertBatch(values interface{}) (int64, error) {

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
		ex.tableName = getTableName(typeOf, valueOf.Index(0))
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

	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

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
func (ex *Builder) GetRows() (*sql.Rows, error) {
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
func (ex *Builder) GetMany(values interface{}) error {
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
func (ex *Builder) GetOne(obj interface{}) error {
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
func (ex *Builder) RawSql(sql string, paramList ...interface{}) *Builder {
	ex.sql = sql
	ex.paramList = paramList
	return ex
}

func (ex *Builder) GetSqlAndParams() (string, []interface{}) {
	if ex.sql != "" {
		return ex.sql, ex.paramList
	}

	var paramList []interface{}

	fieldStr, paramList := handleField(ex.selectList, ex.selectExpList, paramList)
	whereStr, paramList := ex.handleWhere(ex.whereList, paramList)
	joinStr := handleJoin(ex.joinList)
	groupStr := handleGroup(ex.groupList)
	havingStr, paramList := ex.handleHaving(ex.havingList, paramList)
	orderStr := handleOrder(ex.orderList)
	limitStr, paramList := ex.handleLimit(ex.offset, ex.pageSize, paramList)
	lockStr := handleLockForUpdate(ex.isLockForUpdate)

	sqlStr := "SELECT " + fieldStr + " FROM " + ex.tableName + joinStr + whereStr + groupStr + havingStr + orderStr + limitStr + lockStr

	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

	if ex.isDebug {
		fmt.Println(sqlStr)
		fmt.Println(paramList...)
	}

	return sqlStr, paramList
}

// Update 更新记录
func (ex *Builder) Update(dest interface{}) (int64, error) {
	var paramList []any
	setStr, paramList := ex.handleSet(dest, paramList)
	whereStr, paramList := ex.handleWhere(ex.whereList, paramList)
	sqlStr := "UPDATE " + ex.tableName + setStr + whereStr

	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

	return ex.ExecAffected(sqlStr, paramList...)
}

// Delete 删除记录
func (ex *Builder) Delete() (int64, error) {
	var paramList []any
	whereStr, paramList := ex.handleWhere(ex.whereList, paramList)
	sqlStr := "DELETE FROM " + ex.tableName + whereStr

	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

	return ex.ExecAffected(sqlStr, paramList...)
}

// Truncate 清空记录, sqlite3不支持此操作
func (ex *Builder) Truncate() (int64, error) {
	sqlStr := "TRUNCATE TABLE " + ex.tableName
	if ex.driverName == model.Sqlite3 {
		sqlStr = "DELETE FROM " + ex.tableName
	}

	return ex.ExecAffected(sqlStr)
}

// Exists 存在某记录
func (ex *Builder) Exists() (bool, error) {
	var obj IntStruct
	err := ex.Select("1 as c").Limit(0, 1).GetOne(&obj)
	if err != nil {
		return false, err
	}

	if obj.C.Int64 == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

// DoesntExist 不存在某记录
func (ex *Builder) DoesntExist() (bool, error) {
	isE, err := ex.Exists()
	return !isE, err
}

// Value 字段值
func (ex *Builder) Value(fieldName string, dest interface{}) error {
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
func (ex *Builder) Pluck(fieldName string, values interface{}) error {
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
func (ex *Builder) Increment(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := ex.handleWhere(ex.whereList, paramList)
	sqlStr := "UPDATE " + ex.tableName + " SET " + fieldName + "=" + fieldName + "+?" + whereStr

	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

	return ex.ExecAffected(sqlStr, paramList...)
}

// Decrement 某字段自减
func (ex *Builder) Decrement(fieldName string, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := ex.handleWhere(ex.whereList, paramList)
	sqlStr := "UPDATE " + ex.tableName + " SET " + fieldName + "=" + fieldName + "-?" + whereStr

	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

	return ex.ExecAffected(sqlStr, paramList...)
}

// Exec 通用执行-新增,更新,删除
func (ex *Builder) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	if ex.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

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
func (ex *Builder) ExecAffected(sqlStr string, args ...interface{}) (int64, error) {
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
func (ex *Builder) Debug(isDebug bool) *Builder {
	ex.isDebug = isDebug
	return ex
}

// Table 链式操作-从哪个表查询,允许直接写别名,例如 person p
func (ex *Builder) Table(tableName string) *Builder {
	ex.tableName = tableName
	return ex
}

// GroupBy 链式操作,以某字段进行分组
func (ex *Builder) GroupBy(fieldName string) *Builder {
	ex.groupList = append(ex.groupList, fieldName)
	return ex
}

// OrderBy 链式操作,以某字段进行排序
func (ex *Builder) OrderBy(field string, orderType string) *Builder {
	ex.orderList = append(ex.orderList, field+" "+orderType)
	return ex
}

// Limit 链式操作,分页
func (ex *Builder) Limit(offset int, pageSize int) *Builder {
	ex.offset = offset
	ex.pageSize = pageSize
	return ex
}

// Page 链式操作,分页
func (ex *Builder) Page(pageNum int, pageSize int) *Builder {
	ex.offset = (pageNum - 1) * pageSize
	ex.pageSize = pageSize
	return ex
}

// LockForUpdate 加锁, sqlte3不支持此操作
func (ex *Builder) LockForUpdate(isLockForUpdate bool) *Builder {
	ex.isLockForUpdate = isLockForUpdate
	return ex
}

//拼接SQL,查询与筛选通用操作
func (ex *Builder) whereAndHaving(where []WhereItem, paramList []any) ([]string, []any) {
	var whereList []string
	for i := 0; i < len(where); i++ {
		if "**builder.Builder" == reflect.TypeOf(where[i].Val).String() {
			executor := *(**Builder)(unsafe.Pointer(reflect.ValueOf(where[i].Val).Pointer()))
			subSql, subParams := executor.GetSqlAndParams()

			if where[i].Opt != Raw {
				whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+"("+subSql+")")
				paramList = append(paramList, subParams...)
			} else {

			}
		} else {
			if where[i].Opt == Eq || where[i].Opt == Ne || where[i].Opt == Gt || where[i].Opt == Ge || where[i].Opt == Lt || where[i].Opt == Le {
				if ex.driverName == model.Sqlite3 {
					whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+"?")
				} else {
					switch where[i].Val.(type) {
					case float32:
						whereList = append(whereList, ex.getConcatForFloat(where[i].Field, "''")+" "+where[i].Opt+" "+"?")
					case float64:
						whereList = append(whereList, ex.getConcatForFloat(where[i].Field, "''")+" "+where[i].Opt+" "+"?")
					default:
						whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+"?")
					}
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
						paramList = append(paramList, str)
						valueStr = append(valueStr, "?")
					} else {
						valueStr = append(valueStr, "'"+str+"'")
					}
				}

				whereList = append(whereList, where[i].Field+" "+where[i].Opt+" "+ex.getConcatForLike(valueStr...))
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
				whereList = append(whereList, fmt.Sprintf("%v", where[i].Val))
				//没有参数
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
func getTableName(typeOf reflect.Type, valueOf reflect.Value) string {
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

func (ex *Builder) getConcatForFloat(vars ...string) string {
	if ex.driverName == model.Sqlite3 {
		return strings.Join(vars, "||")
	} else if ex.driverName == model.Postgres {
		return vars[0]
	} else {
		return "CONCAT(" + strings.Join(vars, ",") + ")"
	}
}

func (ex *Builder) getConcatForLike(vars ...string) string {
	if ex.driverName == model.Sqlite3 || ex.driverName == model.Postgres {
		return strings.Join(vars, "||")
	} else {
		return "CONCAT(" + strings.Join(vars, ",") + ")"
	}
}
