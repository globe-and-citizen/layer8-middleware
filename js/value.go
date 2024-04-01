package js

import (
	"syscall/js"
)

type Value struct {
	value interface{}
}

func (rs *Value) New(args ...interface{}) *Value {
	arguments := []interface{}{}
	for _, arg := range args {
		switch v := arg.(type) {
		case *Value:
			arguments = append(arguments, v.value.(js.Value))
		case *Func:
			arguments = append(arguments, v.value.(js.Func))
		default:
			arguments = append(arguments, v)
		}
	}

	return &Value{
		value: rs.value.(js.Value).New(arguments...),
	}
}

func (rs *Value) Get(value string) *Value {
	return &Value{
		value: rs.value.(js.Value).Get(value),
	}
}

func (rs *Value) Set(name string, dst interface{}) {
	switch v := dst.(type) {
	case *Value:
		rs.value.(js.Value).Set(name, v.value.(js.Value))
	case *Func:
		rs.value.(js.Value).Set(name, v.value.(js.Func))
	default:
		rs.value.(js.Value).Set(name, v)
	}
}

func (rs *Value) Call(key string, args ...interface{}) *Value {
	arguments := []interface{}{}
	for _, arg := range args {
		switch v := arg.(type) {
		case *Value:
			arguments = append(arguments, v.value.(js.Value))
		case *Func:
			arguments = append(arguments, v.value.(js.Func))
		default:
			arguments = append(arguments, v)
		}
	}

	return &Value{
		value: rs.value.(js.Value).Call(key, arguments...),
	}
}

func (rs *Value) Invoke(args ...interface{}) *Value {
	return &Value{
		value: rs.value.(js.Value).Invoke(args...),
	}
}

func (rs *Value) String() string {
	return rs.value.(js.Value).String()
}

func (rs *Value) Int() int {
	return rs.value.(js.Value).Int()
}

func (rs *Value) Float() float64 {
	return rs.value.(js.Value).Float()
}

func (rs *Value) Bool() bool {
	return rs.value.(js.Value).Bool()
}

func (rs *Value) Type() Type {
	return Type(rs.value.(js.Value).Type())
}

func (rs *Value) Length() int {
	return rs.value.(js.Value).Length()
}

func (rs *Value) Index(i int) *Value {
	return &Value{
		value: rs.value.(js.Value).Index(i),
	}
}
