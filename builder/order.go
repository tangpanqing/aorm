package builder

func (b *Builder) OrderDescBy(field interface{}, prefix ...string) *Builder {
	return b.OrderBy(field, Desc, prefix...)
}

func (b *Builder) OrderAscBy(field interface{}, prefix ...string) *Builder {
	return b.OrderBy(field, Asc, prefix...)
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
