package builder

import (
	"github.com/tangpanqing/aorm/utils"
	"reflect"
)

// Where 链式操作,以对象作为查询条件
func (b *Builder) Where(dest interface{}) *Builder {
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
			b.whereList = append(b.whereList, WhereItem{Field: key, Opt: Eq, Val: val})
		}
	}

	return b
}

// WhereArr 链式操作,以数组作为查询条件
func (b *Builder) WhereArr(whereList []WhereItem) *Builder {
	b.whereList = append(b.whereList, whereList...)
	return b
}

func (b *Builder) WhereEq(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Eq, val, prefix...)
}

func (b *Builder) WhereNe(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Ne, val, prefix...)
}

func (b *Builder) WhereGt(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Gt, val, prefix...)
}

func (b *Builder) WhereGe(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Ge, val, prefix...)
}

func (b *Builder) WhereLt(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Lt, val, prefix...)
}

func (b *Builder) WhereLe(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Le, val, prefix...)
}

func (b *Builder) WhereIn(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, In, val, prefix...)
}

func (b *Builder) WhereNotIn(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, NotIn, val, prefix...)
}

func (b *Builder) WhereBetween(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Between, val, prefix...)
}

func (b *Builder) WhereNotBetween(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, NotBetween, val, prefix...)
}

func (b *Builder) WhereLike(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Like, val, prefix...)
}

func (b *Builder) WhereNotLike(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, NotLike, val, prefix...)
}

func (b *Builder) WhereRaw(val interface{}) *Builder {
	return b.whereItemAppend("", Raw, val)
}

func (b *Builder) WhereFindInSet(val interface{}, field interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, FindInSet, val, prefix...)
}

func (b *Builder) WhereIsNull(field interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Raw, "IS NULL", prefix...)
}

func (b *Builder) WhereIsNOTNull(field interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, Raw, "IS NOT NULL", prefix...)
}

func (b *Builder) WhereRawEq(field interface{}, val interface{}, prefix ...string) *Builder {
	return b.whereItemAppend(field, RawEq, val, prefix...)
}

func (b *Builder) whereItemAppend(field interface{}, opt string, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{prefix, field, opt, val})
	return b
}
