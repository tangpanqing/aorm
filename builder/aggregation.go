package builder

import "github.com/tangpanqing/aorm/null"

type IntStruct struct {
	C null.Int
}

type FloatStruct struct {
	C null.Float
}

// Count 聚合函数-数量
func (b *Builder) Count(fieldName interface{}) (int64, error) {
	var obj []IntStruct
	err := b.SelectCount(fieldName, "c", "").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Int64, nil
}

// Sum 聚合函数-合计
func (b *Builder) Sum(fieldName interface{}) (float64, error) {
	var obj []FloatStruct
	err := b.SelectSum(fieldName, "c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Avg 聚合函数-平均值
func (b *Builder) Avg(fieldName interface{}) (float64, error) {
	var obj []FloatStruct
	err := b.SelectAvg(fieldName, "c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Max 聚合函数-最大值
func (b *Builder) Max(fieldName interface{}) (float64, error) {
	var obj []FloatStruct
	err := b.SelectMax(fieldName, "c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}

// Min 聚合函数-最小值
func (b *Builder) Min(fieldName interface{}) (float64, error) {
	var obj []FloatStruct
	err := b.SelectMin(fieldName, "c").GetMany(&obj)
	if err != nil {
		return 0, err
	}

	return obj[0].C.Float64, nil
}
