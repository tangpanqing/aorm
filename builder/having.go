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
	b.havingList = append(b.havingList, WhereItem{"", field, Eq, val})
	return b
}

func (b *Builder) HavingNe(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, Ne, val})
	return b
}

func (b *Builder) HavingGt(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, Gt, val})
	return b
}

func (b *Builder) HavingGe(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, Ge, val})
	return b
}

func (b *Builder) HavingLt(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, Lt, val})
	return b
}

func (b *Builder) HavingLe(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, Le, val})
	return b
}

func (b *Builder) HavingIn(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, In, val})
	return b
}

func (b *Builder) HavingNotIn(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, NotIn, val})
	return b
}

func (b *Builder) HavingBetween(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, Between, val})
	return b
}

func (b *Builder) HavingNotBetween(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, NotBetween, val})
	return b
}

func (b *Builder) HavingLike(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, Like, val})
	return b
}

func (b *Builder) HavingNotLike(field interface{}, val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", field, NotLike, val})
	return b
}

func (b *Builder) HavingRaw(val interface{}) *Builder {
	b.havingList = append(b.havingList, WhereItem{"", "", Raw, val})
	return b
}
