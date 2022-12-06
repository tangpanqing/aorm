package aorm

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

// Int 整数
type Int struct {
	sql.NullInt64
}

// IntFrom 创建整数
func IntFrom(i int64) Int {
	return Int{
		NullInt64: sql.NullInt64{
			Int64: i,
			Valid: true,
		},
	}
}

// String 字符串
type String struct {
	sql.NullString
}

// StringFrom 创建字符串
func StringFrom(s string) String {
	return String{
		NullString: sql.NullString{
			String: s,
			Valid:  true,
		},
	}
}

// Float 浮点数
type Float struct {
	sql.NullFloat64
}

// FloatFrom 创建浮点数
func FloatFrom(f float64) Float {
	return Float{
		NullFloat64: sql.NullFloat64{
			Float64: f,
			Valid:   true,
		},
	}
}

// Bool 布尔值
type Bool struct {
	sql.NullBool
}

// BoolFrom 创建布尔值
func BoolFrom(b bool) Bool {
	return Bool{
		NullBool: sql.NullBool{
			Bool:  b,
			Valid: true,
		},
	}
}

// Time 时间
type Time struct {
	sql.NullTime
}

// TimeFrom 创建时间
func TimeFrom(t time.Time) Time {
	return Time{
		NullTime: sql.NullTime{
			Time:  t,
			Valid: true,
		},
	}
}

var nullBytes = []byte("null")

// UnmarshalJSON 反序列化浮点数
func (f *Float) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		f.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &f.Float64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need float or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to float: %w", err)
			}
			f.Float64 = n
			f.Valid = true
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	f.Valid = true
	return nil
}

// MarshalJSON 序列化浮点数
func (f Float) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	if math.IsInf(f.Float64, 0) || math.IsNaN(f.Float64) {
		return nil, &json.UnsupportedValueError{
			Value: reflect.ValueOf(f.Float64),
			Str:   strconv.FormatFloat(f.Float64, 'g', -1, 64),
		}
	}
	return []byte(strconv.FormatFloat(f.Float64, 'f', -1, 64)), nil
}

// UnmarshalJSON 反序列化布尔值
func (b *Bool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		b.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &b.Bool); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	b.Valid = true
	return nil
}

// MarshalJSON 序列化布尔值
func (b Bool) MarshalJSON() ([]byte, error) {
	if !b.Valid {
		return []byte("null"), nil
	}
	if !b.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// UnmarshalJSON 反序列化时间
func (t *Time) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		t.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &t.Time); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	t.Valid = true
	return nil
}

// MarshalJSON 序列化时间
func (t Time) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return t.Time.MarshalJSON()
}

// UnmarshalJSON 反序列化整数
func (i *Int) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		i.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &i.Int64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need int or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to int: %w", err)
			}
			i.Int64 = n
			i.Valid = true
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	i.Valid = true
	return nil
}

// MarshalJSON 序列化整数
func (i Int) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(i.Int64, 10)), nil
}

// UnmarshalJSON 反序列化字符串
func (s *String) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		s.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &s.String); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	s.Valid = true
	return nil
}

// MarshalJSON 序列化字符串
func (s String) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(s.String)
}
