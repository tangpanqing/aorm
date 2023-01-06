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

const Count = "COUNT"

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
const RawEq = "RawEq"

// SelectExpItem 将某子语句重命名为某字段
type SelectExpItem struct {
	Executor  **Builder
	FieldName interface{}
}

// Builder 查询记录所需要的条件
type Builder struct {
	//数据库操作连接
	LinkCommon model.LinkCommon

	table      interface{}
	tableAlias string

	//查询参数
	tableName       string
	selectList      []SelectItem
	selectExpList   []*SelectExpItem
	groupList       []GroupItem
	whereList       []WhereItem
	joinList        []JoinItem
	havingList      []WhereItem
	orderList       []OrderItem
	offset          int
	pageSize        int
	distinct        bool
	isDebug         bool
	isLockForUpdate bool

	//sql与参数
	sql       string
	paramList []interface{}

	//驱动名字
	driverName string
}

func (b *Builder) Distinct(distinct bool) *Builder {
	b.distinct = distinct
	return b
}

func (b *Builder) Driver(driverName string) *Builder {
	b.driverName = driverName
	return b
}

func (b *Builder) getTableNameCommon(typeOf reflect.Type, valueOf reflect.Value) string {
	if b.table != nil {
		return getTableNameByTable(b.table)
	}

	return getTableName(typeOf, valueOf)
}

// Insert 增加记录
func (b *Builder) Insert(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//主键名字
	var primaryKey = ""

	var keys []string
	var paramList []any
	var place []string
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		key := helper.UnderLine(typeOf.Elem().Field(i).Name)

		//如果是Postgres数据库，寻找主键
		if b.driverName == model.Postgres {
			tag := typeOf.Elem().Field(i).Tag.Get("aorm")
			if -1 != strings.Index(tag, "primary") {
				primaryKey = key
			}
		}

		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()

			keys = append(keys, key)
			paramList = append(paramList, val)
			place = append(place, "?")
		}
	}

	sqlStr := "INSERT INTO " + b.getTableNameCommon(typeOf, valueOf) + " (" + strings.Join(keys, ",") + ") VALUES (" + strings.Join(place, ",") + ")"

	if b.driverName == model.Mssql {
		return b.insertForMssqlOrPostgres(sqlStr+"; SELECT SCOPE_IDENTITY()", paramList...)
	} else if b.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
		return b.insertForMssqlOrPostgres(sqlStr+" RETURNING "+primaryKey, paramList...)
	} else {
		return b.insertForCommon(sqlStr, paramList...)
	}
}

//对于Mssql,Postgres类型数据库，为了获取最后插入的id，需要改写入为查询
func (b *Builder) insertForMssqlOrPostgres(sql string, paramList ...any) (int64, error) {
	if b.isDebug {
		fmt.Println(sql)
		fmt.Println(paramList...)
	}

	rows, err := b.LinkCommon.Query(sql, paramList...)
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
func (b *Builder) insertForCommon(sql string, paramList ...any) (int64, error) {
	res, err := b.Exec(sql, paramList...)
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
func (b *Builder) InsertBatch(values interface{}) (int64, error) {

	var keys []string
	var paramList []any
	var place []string

	valueOf := reflect.ValueOf(values).Elem()
	if valueOf.Len() == 0 {
		return 0, errors.New("the data list for insert batch not found")
	}
	typeOf := reflect.TypeOf(values).Elem().Elem()

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

	sqlStr := "INSERT INTO " + b.getTableNameCommon(typeOf, valueOf.Index(0)) + " (" + strings.Join(keys, ",") + ") VALUES " + strings.Join(place, ",")

	if b.driverName == model.Postgres {
		sqlStr = convertToPostgresSql(sqlStr)
	}

	res, err := b.Exec(sqlStr, paramList...)
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
func (b *Builder) GetRows() (*sql.Rows, error) {
	sqlStr, paramList := b.GetSqlAndParams()

	smt, errSmt := b.LinkCommon.Prepare(sqlStr)
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
func (b *Builder) GetMany(values interface{}) error {
	rows, errRows := b.GetRows()
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
func (b *Builder) GetOne(obj interface{}) error {
	b.Limit(0, 1)

	rows, errRows := b.GetRows()
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
func (b *Builder) RawSql(sql string, paramList ...interface{}) *Builder {
	b.sql = sql
	b.paramList = paramList
	return b
}

func (b *Builder) GetSqlAndParams() (string, []interface{}) {
	if b.sql != "" {
		return b.sql, b.paramList
	}

	var paramList []interface{}
	tableName := getTableNameByTable(b.table)
	fieldStr, paramList := b.handleField(paramList)
	whereStr, paramList := b.handleWhere(paramList)
	joinStr, paramList := b.handleJoin(paramList)
	groupStr, paramList := b.handleGroup(paramList)
	havingStr, paramList := b.handleHaving(paramList)
	orderStr, paramList := b.handleOrder(paramList)
	limitStr, paramList := b.handleLimit(paramList)
	lockStr := b.handleLockForUpdate()

	sql := "SELECT " + fieldStr + " FROM " + tableName + " " + b.tableAlias + joinStr + whereStr + groupStr + havingStr + orderStr + limitStr + lockStr

	if b.driverName == model.Postgres {
		sql = convertToPostgresSql(sql)
	}

	if b.isDebug {
		fmt.Println(sql)
		fmt.Println(paramList...)
	}

	return sql, paramList
}

// Update 更新记录
func (b *Builder) Update(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	var paramList []any
	setStr, paramList := b.handleSet(typeOf, valueOf, paramList)
	whereStr, paramList := b.handleWhere(paramList)
	sqlStr := "UPDATE " + b.getTableNameCommon(typeOf, valueOf) + setStr + whereStr

	return b.ExecAffected(sqlStr, paramList...)
}

// Delete 删除记录
func (b *Builder) Delete(destList ...interface{}) (int64, error) {
	tableName := ""

	if len(destList) > 0 {
		b.Where(destList[0])

		typeOf := reflect.TypeOf(destList[0])
		valueOf := reflect.ValueOf(destList[0])
		tableName = b.getTableNameCommon(typeOf, valueOf)
	}

	if tableName == "" {
		tableName = getTableNameByTable(b.table)
	}

	var paramList []any
	whereStr, paramList := b.handleWhere(paramList)
	sqlStr := "DELETE FROM " + tableName + whereStr

	return b.ExecAffected(sqlStr, paramList...)
}

// Truncate 清空记录
func (b *Builder) Truncate() (int64, error) {
	sqlStr := ""
	if b.driverName == model.Sqlite3 {
		sqlStr = "DELETE FROM " + getTableNameByTable(b.table)
	} else {
		sqlStr = "TRUNCATE TABLE " + getTableNameByTable(b.table)
	}

	return b.ExecAffected(sqlStr)
}

// Exists 存在某记录
func (b *Builder) Exists() (bool, error) {
	var obj IntStruct

	err := b.selectCommon("", "1 AS c", nil, "").Limit(0, 1).GetOne(&obj)
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
func (b *Builder) DoesntExist() (bool, error) {
	isE, err := b.Exists()
	return !isE, err
}

// Value 字段值
func (b *Builder) Value(field interface{}, dest interface{}) error {
	b.Select(field).Limit(0, 1)

	fieldName := getFieldName(field)

	rows, errRows := b.GetRows()
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
func (b *Builder) Pluck(field interface{}, values interface{}) error {
	b.Select(field)
	fieldName := getFieldName(field)

	rows, errRows := b.GetRows()
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
func (b *Builder) Increment(field interface{}, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := b.handleWhere(paramList)
	sqlStr := "UPDATE " + getTableNameByTable(b.table) + " SET " + getFieldName(field) + "=" + getFieldName(field) + "+?" + whereStr

	return b.ExecAffected(sqlStr, paramList...)
}

// Decrement 某字段自减
func (b *Builder) Decrement(field interface{}, step int) (int64, error) {
	var paramList []any
	paramList = append(paramList, step)
	whereStr, paramList := b.handleWhere(paramList)
	sqlStr := "UPDATE " + getTableNameByTable(b.table) + " SET " + getFieldName(field) + "=" + getFieldName(field) + "-?" + whereStr

	return b.ExecAffected(sqlStr, paramList...)
}

// Exec 通用执行-新增,更新,删除
func (b *Builder) Exec(sql string, paramList ...interface{}) (sql.Result, error) {
	if b.driverName == model.Postgres {
		sql = convertToPostgresSql(sql)
	}

	if b.isDebug {
		fmt.Println(sql)
		fmt.Println(paramList...)
	}

	smt, err1 := b.LinkCommon.Prepare(sql)
	if err1 != nil {
		return nil, err1
	}
	defer smt.Close()

	res, err2 := smt.Exec(paramList...)
	if err2 != nil {
		return nil, err2
	}

	//b.clear()
	return res, nil
}

// ExecAffected 通用执行-更新,删除
func (b *Builder) ExecAffected(sql string, paramList ...interface{}) (int64, error) {
	if b.driverName == model.Postgres {
		sql = convertToPostgresSql(sql)
	}

	res, err := b.Exec(sql, paramList...)
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
func (b *Builder) Debug(isDebug bool) *Builder {
	b.isDebug = isDebug
	return b
}

// Table 链式操作-从哪个表查询,允许直接写别名,例如 person p
func (b *Builder) Table(table interface{}, alias ...string) *Builder {
	b.table = table
	if len(alias) > 0 {
		b.tableAlias = alias[0]
	}
	return b
}

// GroupBy 链式操作,以某字段进行分组
func (b *Builder) GroupBy(field interface{}, prefix ...string) *Builder {
	b.groupList = append(b.groupList, GroupItem{
		Prefix: getPrefixByField(field, prefix...),
		Field:  field,
	})
	return b
}

// OrderBy 链式操作,以某字段进行排序
func (b *Builder) OrderBy(field interface{}, orderType string, prefix ...string) *Builder {
	b.orderList = append(b.orderList, OrderItem{
		Prefix:    getPrefixByField(field, prefix...),
		Field:     field,
		OrderType: orderType,
	})

	return b
}

// Limit 链式操作,分页
func (b *Builder) Limit(offset int, pageSize int) *Builder {
	b.offset = offset
	b.pageSize = pageSize
	return b
}

// Page 链式操作,分页
func (b *Builder) Page(pageNum int, pageSize int) *Builder {
	b.offset = (pageNum - 1) * pageSize
	b.pageSize = pageSize
	return b
}

// LockForUpdate 加锁, sqlte3不支持此操作
func (b *Builder) LockForUpdate(isLockForUpdate bool) *Builder {
	b.isLockForUpdate = isLockForUpdate
	return b
}

//拼接SQL,查询与筛选通用操作
func (b *Builder) whereAndHaving(where []WhereItem, paramList []any, isFromHaving bool) ([]string, []any) {
	var whereList []string
	for i := 0; i < len(where); i++ {
		allFieldName := ""
		if where[i].Prefix != "" {
			allFieldName += where[i].Prefix + "."
		}

		//如果是mssql或者Postgres,并且来自having的话，需要特殊处理
		if (b.driverName == model.Mssql || b.driverName == model.Postgres) && isFromHaving {
			fieldNameCurrent := getFieldName(where[i].Field)
			for m := 0; m < len(b.selectList); m++ {
				if fieldNameCurrent == getFieldName(b.selectList[m].FieldNew) {
					allFieldName += handleFieldWith(b.selectList[m])
				}
			}
		} else {
			allFieldName += getFieldName(where[i].Field)
		}

		if "**builder.Builder" == reflect.TypeOf(where[i].Val).String() {
			executor := *(**Builder)(unsafe.Pointer(reflect.ValueOf(where[i].Val).Pointer()))
			subSql, subParams := executor.GetSqlAndParams()

			if where[i].Opt != Raw {
				whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+"("+subSql+")")
				paramList = append(paramList, subParams...)
			}
		} else {
			if where[i].Opt == Eq || where[i].Opt == Ne || where[i].Opt == Gt || where[i].Opt == Ge || where[i].Opt == Lt || where[i].Opt == Le {
				if b.driverName == model.Sqlite3 {
					whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+"?")
				} else {
					switch where[i].Val.(type) {
					case float32:
						whereList = append(whereList, b.getConcatForFloat(allFieldName, "''")+" "+where[i].Opt+" "+"?")
					case float64:
						whereList = append(whereList, b.getConcatForFloat(allFieldName, "''")+" "+where[i].Opt+" "+"?")
					default:
						whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+"?")
					}
				}

				paramList = append(paramList, fmt.Sprintf("%v", where[i].Val))
			}

			if where[i].Opt == Between || where[i].Opt == NotBetween {
				values := toAnyArr(where[i].Val)
				whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+"(?) AND (?)")
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

				whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+b.getConcatForLike(valueStr...))
			}

			if where[i].Opt == In || where[i].Opt == NotIn {
				values := toAnyArr(where[i].Val)
				var placeholder []string
				for j := 0; j < len(values); j++ {
					placeholder = append(placeholder, "?")
				}

				whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+"("+strings.Join(placeholder, ",")+")")
				paramList = append(paramList, values...)
			}

			if where[i].Opt == Raw {
				whereList = append(whereList, allFieldName+fmt.Sprintf("%v", where[i].Val))
			}

			if where[i].Opt == RawEq {
				whereList = append(whereList, allFieldName+Eq+getPrefixByField(where[i].Val)+"."+getFieldName(where[i].Val))
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

func (b *Builder) getConcatForFloat(vars ...string) string {
	if b.driverName == model.Sqlite3 {
		return strings.Join(vars, "||")
	} else if b.driverName == model.Postgres {
		return vars[0]
	} else {
		return "CONCAT(" + strings.Join(vars, ",") + ")"
	}
}

func (b *Builder) getConcatForLike(vars ...string) string {
	if b.driverName == model.Sqlite3 || b.driverName == model.Postgres {
		return strings.Join(vars, "||")
	} else {
		return "CONCAT(" + strings.Join(vars, ",") + ")"
	}
}
