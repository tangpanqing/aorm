package builder

// Exists 存在某记录
func (b *Builder) Exists() (bool, error) {
	var obj IntStruct

	err := b.selectCommon("", "1 AS c", nil, "").Limit(0, 1).GetOne(&obj)
	if err != nil {
		return false, err
	}

	if obj.C.Int64 == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

// DoesntExist 不存在某记录
func (b *Builder) DoesntExist() (bool, error) {
	isE, err := b.Exists()
	return !isE, err
}
