package builder

import (
	"github.com/tangpanqing/aorm/utils"
	"reflect"
)

// Having 链式操作,以对象作为筛选条件
func (b *Builder) Having(dest interface{}) *Builder {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if b.table == nil {
		b.table = getTableNameByReflect(typeOf, valueOf)
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := utils.UnderLine(typeOf.Elem().Field(i).Name)
			val := valueOf.Elem().Field(i).Field(0).Field(0).Interface()
			b.havingList = append(b.havingList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return b
}

// HavingArr 链式操作,以数组作为筛选条件
func (b *Builder) HavingArr(havingList []WhereItem) *Builder {
	b.havingList = append(b.havingList, havingList...)
	return b
}

func (b *Builder) HavingEq(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Eq, val)
}

func (b *Builder) HavingNe(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Ne, val)
}

func (b *Builder) HavingGt(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Gt, val)
}

func (b *Builder) HavingGe(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Ge, val)
}

func (b *Builder) HavingLt(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Lt, val)
}

func (b *Builder) HavingLe(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Le, val)
}

func (b *Builder) HavingIn(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, In, val)
}

func (b *Builder) HavingNotIn(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, NotIn, val)
}

func (b *Builder) HavingBetween(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Between, val)
}

func (b *Builder) HavingNotBetween(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, NotBetween, val)
}

func (b *Builder) HavingLike(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, Like, val)
}

func (b *Builder) HavingNotLike(field interface{}, val interface{}) *Builder {
	return b.havingItemAppend(field, NotLike, val)
}

func (b *Builder) HavingRaw(val interface{}) *Builder {
	return b.havingItemAppend("", Raw, val)
}

func (b *Builder) havingItemAppend(field interface{}, opt string, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{[]string{""}, field, opt, val})
	return b
}
