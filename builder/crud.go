package builder

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/tangpanqing/aorm/base"
	"github.com/tangpanqing/aorm/driver"
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
const RawEq = "RawEq"

// Builder 查询记录所需要的条件
type Builder struct {
	Link base.Link

	table      interface{}
	tableAlias string

	selectList    []SelectItem
	selectExpList []*SelectExpItem
	groupList     []GroupItem
	whereList     []WhereItem
	joinList      []JoinItem
	havingList    []WhereItem
	orderList     []OrderItem
	limitItem     LimitItem

	distinct        bool
	isDebug         bool
	isLockForUpdate bool

	//sql与参数
	query string
	args  []interface{}
}

// Debug 链式操作-是否开启调试,打印sql
func (b *Builder) Debug(isDebug bool) *Builder {
	b.isDebug = isDebug
	return b
}

// Distinct 过滤重复记录
func (b *Builder) Distinct(distinct bool) *Builder {
	b.distinct = distinct
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

// Insert 增加记录
func (b *Builder) Insert(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//主键名字
	var primaryKey = ""

	var keys []string
	var args []any
	var place []string
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		key, tagMap := getFieldNameByStructField(typeOf.Elem().Field(i))

		//如果是Postgres数据库，寻找主键
		if b.Link.DriverName() == driver.Postgres {
			if _, ok := tagMap["primary"]; ok {
				primaryKey = key
			}
		}

		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()

			keys = append(keys, key)
			args = append(args, val)
			place = append(place, "?")
		}
	}

	query := "INSERT INTO " + b.getTableNameCommon(typeOf, valueOf) + " (" + strings.Join(keys, ",") + ") VALUES (" + strings.Join(place, ",") + ")"

	if b.Link.DriverName() == driver.Mssql {
		return b.insertForMssqlOrPostgres(query+"; SELECT SCOPE_IDENTITY()", args...)
	} else if b.Link.DriverName() == driver.Postgres {
		return b.insertForMssqlOrPostgres(convertToPostgresSql(query)+" RETURNING "+primaryKey, args...)
	} else {
		return b.insertForCommon(query, args...)
	}
}

//对于Mssql,Postgres类型数据库，为了获取最后插入的id，需要改写入为查询
func (b *Builder) insertForMssqlOrPostgres(query string, args ...any) (int64, error) {
	if b.isDebug {
		fmt.Println(query)
		fmt.Println(args...)
	}

	rows, err := b.Link.Query(query, args...)
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
func (b *Builder) insertForCommon(query string, args ...any) (int64, error) {
	res, err := b.RawSql(query, args...).Exec()
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
func (b *Builder) InsertBatch(values interface{}) (int64, error) {
	var keys []string
	var args []any
	var place []string

	valueOf := reflect.ValueOf(values).Elem()

	if valueOf.Len() == 0 {
		return 0, errors.New("the data list for insert batch not found")
	}
	typeOf := reflect.TypeOf(values).Elem().Elem()

	for j := 0; j < valueOf.Len(); j++ {
		var placeItem []string

		for i := 0; i < valueOf.Index(j).Elem().NumField(); i++ {
			isNotNull := valueOf.Index(j).Elem().Field(i).Field(0).Field(1).Bool()
			if isNotNull {
				if j == 0 {
					key, _ := getFieldNameByStructField(typeOf.Elem().Field(i))
					keys = append(keys, key)
				}

				val := valueOf.Index(j).Elem().Field(i).Field(0).Field(0).Interface()
				args = append(args, val)
				placeItem = append(placeItem, "?")
			}
		}

		place = append(place, "("+strings.Join(placeItem, ",")+")")
	}

	query := "INSERT INTO " + b.getTableNameCommon(typeOf, valueOf.Index(0)) + " (" + strings.Join(keys, ",") + ") VALUES " + strings.Join(place, ",")

	if b.Link.DriverName() == driver.Postgres {
		query = convertToPostgresSql(query)
	}

	res, err := b.RawSql(query, args...).Exec()
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
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
	fieldNameMap := getFieldMapByReflect(destValue, destType)

	for rows.Next() {
		scans := getScansAddr(columnNameList, fieldNameMap, destValue)

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

	if rows.Next() {
		destType := reflect.TypeOf(obj).Elem()
		destValue := reflect.ValueOf(obj).Elem()

		//从数据库中读出来的字段名字
		columnNameList, errColumns := rows.Columns()
		if errColumns != nil {
			return errColumns
		}

		//从结构体反射出来的属性名
		fieldNameMap := getFieldMapByReflect(destValue, destType)

		scans := getScansAddr(columnNameList, fieldNameMap, destValue)
		err := rows.Scan(scans...)
		if err != nil {
			return err
		}

		return nil
	} else {
		return errors.New("NOT FOUND")
	}
}

// Update 更新记录
func (b *Builder) Update(dest interface{}) (int64, error) {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	var args []any
	setStr, args := b.handleSet(typeOf, valueOf, args)
	whereStr, args := b.handleWhere(args, false)
	query := "UPDATE " + b.getTableNameCommon(typeOf, valueOf) + setStr + whereStr

	return b.execAffected(query, args...)
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

	var args []any
	whereStr, args := b.handleWhere(args, false)
	query := "DELETE FROM " + tableName + whereStr

	return b.execAffected(query, args...)
}

// GroupBy 链式操作,以某字段进行分组
func (b *Builder) GroupBy(field interface{}, prefix ...string) *Builder {
	b.groupList = append(b.groupList, GroupItem{
		Prefix: prefix,
		Field:  field,
	})
	return b
}

// Limit 链式操作,分页
func (b *Builder) Limit(offset int, pageSize int) *Builder {
	b.limitItem = LimitItem{
		offset:   offset,
		pageSize: pageSize,
	}
	return b
}

// Page 链式操作,分页
func (b *Builder) Page(pageNum int, pageSize int) *Builder {
	b.limitItem = LimitItem{
		offset:   (pageNum - 1) * pageSize,
		pageSize: pageSize,
	}
	return b
}

// LockForUpdate 加锁, sqlite3不支持此操作
func (b *Builder) LockForUpdate(isLockForUpdate bool) *Builder {
	b.isLockForUpdate = isLockForUpdate
	return b
}

// Truncate 清空记录
func (b *Builder) Truncate() (int64, error) {
	query := ""
	if b.Link.DriverName() == driver.Sqlite3 {
		query = "DELETE FROM " + getTableNameByTable(b.table)
	} else {
		query = "TRUNCATE TABLE " + getTableNameByTable(b.table)
	}

	return b.execAffected(query)
}

// RawSql 执行原始的sql语句
func (b *Builder) RawSql(query string, args ...interface{}) *Builder {
	b.query = query
	b.args = args
	return b
}

// GetRows 获取行操作
func (b *Builder) GetRows() (*sql.Rows, error) {
	query, args := b.GetSqlAndParams()

	if b.Link.DriverName() == driver.Postgres {
		query = convertToPostgresSql(query)
	}

	if b.isDebug {
		fmt.Println(query)
		fmt.Println(args...)
	}

	smt, errSmt := b.Link.Prepare(query)
	if errSmt != nil {
		return nil, errSmt
	}
	//defer smt.Close()

	rows, errRows := smt.Query(args...)
	if errRows != nil {
		return nil, errRows
	}

	return rows, nil
}

// Exec 通用执行-新增,更新,删除
func (b *Builder) Exec() (sql.Result, error) {
	if b.Link.DriverName() == driver.Postgres {
		b.query = convertToPostgresSql(b.query)
	}

	if b.isDebug {
		fmt.Println(b.query)
		fmt.Println(b.args...)
	}

	smt, err1 := b.Link.Prepare(b.query)
	if err1 != nil {
		return nil, err1
	}
	defer smt.Close()

	res, err2 := smt.Exec(b.args...)
	if err2 != nil {
		return nil, err2
	}

	//b.clear()
	return res, nil
}

//拼接SQL,查询与筛选通用操作
func (b *Builder) whereAndHaving(where []WhereItem, args []any, isFromHaving bool, needPrefix bool) ([]string, []any) {
	var whereList []string
	for i := 0; i < len(where); i++ {
		valueOfField := reflect.ValueOf(where[i].Field)

		allFieldName := ""
		if needPrefix {
			prefix := getPrefixByField(valueOfField, where[i].Prefix...)
			if prefix != "" {
				allFieldName += prefix + "."
			}
		}

		//如果是mssql或者Postgres,并且来自having的话，需要特殊处理
		if (b.Link.DriverName() == driver.Mssql || b.Link.DriverName() == driver.Postgres) && isFromHaving {
			fieldNameCurrent := getFieldNameByReflectValue(valueOfField)
			for m := 0; m < len(b.selectList); m++ {
				if fieldNameCurrent == getFieldNameByField(b.selectList[m].FieldNew) {
					allFieldName += handleSelectWith(b.selectList[m])
				}
			}
		} else {
			allFieldName += getFieldNameByReflectValue(valueOfField)
		}

		if "**builder.Builder" == reflect.TypeOf(where[i].Val).String() {
			subBuilder := *(**Builder)(unsafe.Pointer(reflect.ValueOf(where[i].Val).Pointer()))
			subSql, subParams := subBuilder.GetSqlAndParams()

			if where[i].Opt != Raw {
				whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+"("+subSql+")")
				args = append(args, subParams...)
			}
		} else {
			if where[i].Opt == Eq || where[i].Opt == Ne || where[i].Opt == Gt || where[i].Opt == Ge || where[i].Opt == Lt || where[i].Opt == Le {
				if b.Link.DriverName() == driver.Sqlite3 {
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

				args = append(args, fmt.Sprintf("%v", where[i].Val))
			}

			if where[i].Opt == Between || where[i].Opt == NotBetween {
				values := toAnyArr(where[i].Val)
				whereList = append(whereList, allFieldName+" "+where[i].Opt+" "+"(?) AND (?)")
				args = append(args, values...)
			}

			if where[i].Opt == Like || where[i].Opt == NotLike {
				values := toAnyArr(where[i].Val)
				var valueStr []string
				for j := 0; j < len(values); j++ {
					str := fmt.Sprintf("%v", values[j])

					if "%" != str {
						args = append(args, str)
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
				args = append(args, values...)
			}

			if where[i].Opt == Raw {
				whereList = append(whereList, allFieldName+fmt.Sprintf("%v", where[i].Val))
			}

			if where[i].Opt == RawEq {
				whereList = append(whereList, allFieldName+Eq+getPrefixByField(reflect.ValueOf(where[i].Val))+"."+getFieldNameByField(where[i].Val))
			}
		}
	}
	return whereList, args
}

func (b *Builder) getConcatForFloat(vars ...string) string {
	if b.Link.DriverName() == driver.Sqlite3 {
		return strings.Join(vars, "||")
	} else if b.Link.DriverName() == driver.Postgres {
		return vars[0]
	} else {
		return "CONCAT(" + strings.Join(vars, ",") + ")"
	}
}

func (b *Builder) getConcatForLike(vars ...string) string {
	if b.Link.DriverName() == driver.Sqlite3 || b.Link.DriverName() == driver.Postgres {
		return strings.Join(vars, "||")
	} else {
		return "CONCAT(" + strings.Join(vars, ",") + ")"
	}
}

func (b *Builder) getTableNameCommon(typeOf reflect.Type, valueOf reflect.Value) string {
	if b.table != nil {
		return getTableNameByTable(b.table)
	}

	return getTableNameByReflect(typeOf, valueOf)
}

func (b *Builder) GetSqlAndParams() (string, []interface{}) {
	if b.query != "" {
		return b.query, b.args
	}

	var args []interface{}
	tableName := getTableNameByTable(b.table)
	fieldStr, args := b.handleSelect(args)
	whereStr, args := b.handleWhere(args, true)
	joinStr, args := b.handleJoin(args)
	groupStr, args := b.handleGroup(args)
	havingStr, args := b.handleHaving(args)
	orderStr, args := b.handleOrder(args)
	limitStr, args := b.handleLimit(args)
	lockStr := b.handleLockForUpdate()

	query := "SELECT " + fieldStr + " FROM " + tableName + " " + b.tableAlias + joinStr + whereStr + groupStr + havingStr + orderStr + limitStr + lockStr

	return query, args
}

// execAffected 通用执行-更新,删除
func (b *Builder) execAffected(query string, args ...interface{}) (int64, error) {
	if b.Link.DriverName() == driver.Postgres {
		query = convertToPostgresSql(query)
	}

	res, err := b.RawSql(query, args...).Exec()
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func getTagMap(fieldTag string) map[string]string {
	var fieldMap = make(map[string]string)
	if "" != fieldTag {
		tagArr := strings.Split(fieldTag, ";")
		for j := 0; j < len(tagArr); j++ {
			tagArrArr := strings.Split(tagArr[j], ":")
			fieldMap[tagArrArr[0]] = ""
			if len(tagArrArr) > 1 {
				fieldMap[tagArrArr[0]] = tagArrArr[1]
			}
		}
	}
	return fieldMap
}

//对于Postgres数据库，不支持?占位符，支持$1,$2类型，需要做转换
func convertToPostgresSql(query string) string {
	t := 1
	for {
		if strings.Index(query, "?") == -1 {
			break
		}
		query = strings.Replace(query, "?", "$"+strconv.Itoa(t), 1)
		t += 1
	}

	return query
}
