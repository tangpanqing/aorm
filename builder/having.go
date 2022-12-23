package builder

import (
	"github.com/tangpanqing/aorm/helper"
	"reflect"
)

// Having 链式操作,以对象作为筛选条件
func (ex *Builder) Having(dest interface{}) *Builder {
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
			ex.havingList = append(ex.havingList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return ex
}

// HavingArr 链式操作,以数组作为筛选条件
func (ex *Builder) HavingArr(havingList []WhereItem) *Builder {
	ex.havingList = append(ex.havingList, havingList...)
	return ex
}

func (ex *Builder) HavingEq(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Eq,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingNe(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Ne,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingGt(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Gt,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingGe(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Ge,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingLt(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Lt,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingLe(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Le,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingIn(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   In,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingNotIn(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   NotIn,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingBetween(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Between,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingNotBetween(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   NotBetween,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingLike(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Like,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingNotLike(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   NotLike,
		Val:   val,
	})
	return ex
}

func (ex *Builder) HavingRaw(field string, val interface{}) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		Field: field,
		Opt:   Raw,
		Val:   val,
	})
	return ex
}
