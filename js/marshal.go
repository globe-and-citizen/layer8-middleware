package js

import (
	"syscall/js"
)

type (
	Type  int
	Value struct {
		_type        Type
		_constructor string

		value interface{}
	}
)

const (
	TypeNumber Type = iota
	TypeBoolean
	TypeString
	TypeObject
	TypeArray
	TypeNull
)

func NewValue(value interface{}) *Value {
	switch value.(type) {
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		return &Value{
			_type:        TypeNumber,
			_constructor: "Number",
			value:        value,
		}
	case bool:
		return &Value{
			_type:        TypeBoolean,
			_constructor: "Boolean",
			value:        value,
		}
	case string:
		return &Value{
			_type:        TypeString,
			_constructor: "String",
			value:        value,
		}
	case map[string]interface{}:
		return &Value{
			_type:        TypeObject,
			_constructor: "Object",
			value:        value,
		}
	case []interface{}:
		return &Value{
			_type:        TypeArray,
			_constructor: "Array",
			value:        value,
		}
	default:
		return &Value{
			_type:        TypeNull,
			_constructor: "Null",
			value:        nil,
		}
	}
}

func (v Value) Value() interface{} {
	return v.value
}

func (v Value) Get(key string) interface{} {
	val, ok := v.value.(map[string]interface{})[key]
	if !ok {
		return nil
	}
	return val.(Value).value
}

func (v Value) Set(key string, value interface{}) {
	switch value.(type) {
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		v.value.(map[string]interface{})[key] = Value{
			_type:        TypeNumber,
			_constructor: "Number",
			value:        value,
		}
	case bool:
		v.value.(map[string]interface{})[key] = Value{
			_type:        TypeBoolean,
			_constructor: "Boolean",
			value:        value,
		}
	case string:
		v.value.(map[string]interface{})[key] = Value{
			_type:        TypeString,
			_constructor: "String",
			value:        value,
		}
	case map[string]interface{}:
		v.value.(map[string]interface{})[key] = Value{
			_type:        TypeObject,
			_constructor: "Object",
			value:        value,
		}
	case []interface{}:
		v.value.(map[string]interface{})[key] = Value{
			_type:        TypeArray,
			_constructor: "Array",
			value:        value,
		}
	default:
		v.value.(map[string]interface{})[key] = Value{
			_type:        TypeNull,
			_constructor: "Null",
			value:        nil,
		}
	}
}

func (v Value) String() string {
	if v._type == TypeString {
		return v.value.(string)
	}
	return ""
}

func (v Value) Type() Type {
	return v._type
}

func (v Value) Constructor() string {
	return v._constructor
}

// Unmarshal unmarshals a `syscall/js.Value` into the internal `Value` type.
//
// Note that only primitive types are supported and will be unmarshaled
//   - Number
//   - Boolean
//   - String
//   - Object (recursively converted to map[string]interface{})
//   - Array (recursively converted to []interface{})
func Unmarshal(v js.Value) *Value {
	keys := js.Global().Get("Object").Call("keys", v)
	m := Value{
		_type:        TypeObject,
		_constructor: v.Get("constructor").Get("name").String(),
		value:        make(map[string]interface{}),
	}

	for i := 0; i < keys.Length(); i++ {
		key := keys.Index(i).String()
		val := v.Get(key)

		switch val.Type() {
		case js.TypeNumber:
			m.value.(map[string]interface{})[key] = Value{
				_type:        TypeNumber,
				_constructor: "Number",
				value:        val.Float(),
			}
		case js.TypeBoolean:
			m.value.(map[string]interface{})[key] = Value{
				_type:        TypeBoolean,
				_constructor: "Boolean",
				value:        val.Bool(),
			}
		case js.TypeString:
			m.value.(map[string]interface{})[key] = Value{
				_type:        TypeString,
				_constructor: "String",
				value:        val.String(),
			}
		case js.TypeObject:
			if val.Get("constructor").Type() == js.TypeUndefined || (val.Get("constructor").Get("name").String() != "Array" &&
			val.Get("constructor").Get("name").String() != "Object") {
				continue
			}
			
			if val.Get("constructor").Get("name").String() == "Array" {
				m.value.(map[string]interface{})[key] = Value{
					_type:        TypeArray,
					_constructor: val.Get("constructor").Get("name").String(),
					value:        parseObjectToSlice(val),
				}
			} else if val.Get("constructor").Get("name").String() == "Object" {
				m.value.(map[string]interface{})[key] = Unmarshal(val)
			}
		default:
			m.value.(map[string]interface{})[key] = Value{
				_type:        TypeNull,
				_constructor: "Null",
				value:        nil,
			}
		}
	}

	return &m
}

func parseObjectToSlice(obj js.Value) []interface{} {
	var s []interface{}

	for i := 0; i < obj.Length(); i++ {
		val := obj.Index(i)

		switch val.Type() {
		case js.TypeNumber:
			s = append(s, val.Float())
		case js.TypeBoolean:
			s = append(s, val.Bool())
		case js.TypeString:
			s = append(s, val.String())
		case js.TypeObject:
			if val.Get("constructor").Get("name").String() == "Array" {
				s = append(s, parseObjectToSlice(val))
				continue
			}
			s = append(s, Unmarshal(val))
		}
	}

	return s
}
