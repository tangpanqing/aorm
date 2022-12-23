package builder

import (
	"github.com/tangpanqing/aorm/helper"
	"reflect"
)

// Where 链式操作,以对象作为查询条件
func (ex *Builder) Where(dest interface{}) *Builder {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if ex.tableName == "" {
		ex.tableName = getTableName(typeOf, valueOf)
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
func (ex *Builder) WhereArr(whereList []WhereItem) *Builder {
	ex.whereList = append(ex.whereList, whereList...)
	return ex
}

func (ex *Builder) WhereEq(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Eq,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereNe(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Ne,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereGt(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Gt,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereGe(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Ge,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereLt(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Lt,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereLe(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Le,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereIn(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   In,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereNotIn(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   NotIn,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereBetween(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Between,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereNotBetween(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   NotBetween,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereLike(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Like,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereNotLike(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   NotLike,
		Val:   val,
	})
	return ex
}

func (ex *Builder) WhereRaw(field string, val interface{}) *Builder {
	ex.whereList = append(ex.whereList, WhereItem{
		Field: field,
		Opt:   Raw,
		Val:   val,
	})
	return ex
}
