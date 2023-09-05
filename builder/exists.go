package builder

// Exists 存在某记录
func (b *Builder) Exists() (bool, error) {
	stmt, rows, err := b.selectCommon("", "1", nil, "").Limit(0, 1).GetRows()
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	defer rows.Close()

	if rows.Next() {
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
