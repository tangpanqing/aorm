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

func (ex *Builder) HavingEq(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Eq,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingNe(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Ne,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingGt(field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: "",
		Prefix:   "",
		Field:    field,
		Opt:      Gt,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingGe(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Ge,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingLt(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Lt,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingLe(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Le,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingIn(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      In,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingNotIn(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      NotIn,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingBetween(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Between,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingNotBetween(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      NotBetween,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingLike(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Like,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingNotLike(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      NotLike,
		Val:      val,
	})
	return ex
}

func (ex *Builder) HavingRaw(funcName string, field interface{}, val interface{}, prefix ...string) *Builder {
	ex.havingList = append(ex.havingList, WhereItem{
		FuncName: funcName,
		Prefix:   getPrefixByField(field, prefix...),
		Field:    field,
		Opt:      Raw,
		Val:      val,
	})
	return ex
}
