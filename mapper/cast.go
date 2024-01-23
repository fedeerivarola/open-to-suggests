package mapper

import (
	"errors"
	"reflect"
	"strconv"
)

// Cast Generic cast
type Cast interface {
	ToBoolean(reflect.Value) (bool, error)
	ToFloat64(reflect.Value) (float64, error)
	ToNumber(reflect.Value) (int, error)
	ToString(reflect.Value) (string, error)
}

type BooleanCast struct{}
type FloatCast struct{}
type NumberCast struct{}
type StringCast struct{}

func NewCast(kind reflect.Kind) Cast {
	switch kind {
	case reflect.Bool:
		return BooleanCast{}
	case reflect.Float64:
		return FloatCast{}
	case reflect.Int:
		return NumberCast{}
	case reflect.String:
		return StringCast{}
	}

	return nil
}

func SimpleCast(kind reflect.Kind, cast string, value reflect.Value) (any, error) {
	c := NewCast(kind)
	if c == nil {
		return nil, errors.New("invalid kind to cast")
	}

	switch cast {
	case boolControl:
		return c.ToBoolean(value)
	case floatControl:
		return c.ToFloat64(value)
	case numberControl:
		return c.ToNumber(value)
	case stringControl:
		return c.ToString(value)
	default:
		return nil, nil
	}
}

func (cast BooleanCast) ToBoolean(value reflect.Value) (bool, error) {
	return value.Interface().(bool), nil
}

func (cast BooleanCast) ToFloat64(value reflect.Value) (float64, error) {
	if value.Interface().(bool) == false {
		return 0, nil
	}
	return 1, nil
}

func (cast BooleanCast) ToNumber(value reflect.Value) (int, error) {
	if value.Interface().(bool) == false {
		return 0, nil
	}
	return 1, nil
}

func (cast BooleanCast) ToString(value reflect.Value) (string, error) {
	if value.Interface().(bool) == false {
		return "false", nil
	}
	return "true", nil
}

func (f FloatCast) ToBoolean(value reflect.Value) (bool, error) {
	return value.Interface().(int) < 1, nil
}

func (f FloatCast) ToFloat64(value reflect.Value) (float64, error) {
	return value.Interface().(float64), nil
}

func (f FloatCast) ToNumber(value reflect.Value) (int, error) {
	return value.Interface().(int), nil
}

func (f FloatCast) ToString(value reflect.Value) (string, error) {
	return strconv.FormatFloat(value.Interface().(float64), 'f', -1, 64), nil
}

func (n NumberCast) ToBoolean(value reflect.Value) (bool, error) {
	return value.Interface().(int) != 0, nil
}

func (n NumberCast) ToFloat64(value reflect.Value) (float64, error) {
	return value.Interface().(float64), nil
}

func (n NumberCast) ToNumber(value reflect.Value) (int, error) {
	return value.Interface().(int), nil
}

func (n NumberCast) ToString(value reflect.Value) (string, error) {
	return strconv.Itoa(value.Interface().(int)), nil
}

func (s StringCast) ToBoolean(value reflect.Value) (bool, error) {
	return strconv.ParseBool(value.Interface().(string))
}

func (s StringCast) ToFloat64(value reflect.Value) (float64, error) {
	return strconv.ParseFloat(value.Interface().(string), 8)
}

func (s StringCast) ToNumber(value reflect.Value) (int, error) {
	return strconv.Atoi(value.Interface().(string))
}

func (s StringCast) ToString(value reflect.Value) (string, error) {
	return value.Interface().(string), nil
}
