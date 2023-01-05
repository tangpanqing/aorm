package builder

import (
	"github.com/tangpanqing/aorm/helper"
	"reflect"
)

// Where 链式操作,以对象作为查询条件
func (b *Builder) Where(dest interface{}) *Builder {
	typeOf := reflect.TypeOf(dest)
	valueOf := reflect.ValueOf(dest)

	//如果没有设置表名
	if b.tableName == "" {
		b.tableName = getTableName(typeOf, valueOf)
	}

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		isNotNull := valueOf.Elem().Field(i).Field(0).Field(1).Bool()
		if isNotNull {
			key := helper.UnderLine(typeOf.Elem().Field(i).Name)
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
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Eq, val})
	return b
}

func (b *Builder) WhereNe(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Ne, val})
	return b
}

func (b *Builder) WhereGt(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Gt, val})
	return b
}

func (b *Builder) WhereGe(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Ge, val})
	return b
}

func (b *Builder) WhereLt(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Lt, val})
	return b
}

func (b *Builder) WhereLe(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Le, val})
	return b
}

func (b *Builder) WhereIn(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, In, val})
	return b
}

func (b *Builder) WhereNotIn(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, NotIn, val})
	return b
}

func (b *Builder) WhereBetween(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Between, val})
	return b
}

func (b *Builder) WhereNotBetween(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, NotBetween, val})
	return b
}

func (b *Builder) WhereLike(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Like, val})
	return b
}

func (b *Builder) WhereNotLike(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, NotLike, val})
	return b
}

func (b *Builder) WhereRaw(field interface{}, val interface{}, prefix ...string) *Builder {
	b.whereList = append(b.whereList, WhereItem{"", getPrefixByField(field, prefix...), field, Raw, val})
	return b
}
