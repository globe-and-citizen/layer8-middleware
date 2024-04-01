package js

import (
	"syscall/js"
)

type Type int

const (
	TypeUndefined Type = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeSymbol
	TypeObject
	TypeFunction
)

func (t Type) String() string {
	switch t {
	case TypeUndefined:
		return "<undefined>"
	case TypeNull:
		return "<null>"
	case TypeBoolean:
		return "<boolean>"
	case TypeNumber:
		return "<number>"
	case TypeString:
		return "<string>"
	case TypeSymbol:
		return "<symbol>"
	case TypeObject:
		return "<object>"
	case TypeFunction:
		return "<function>"
	}
	return "<unknown>"
}

func Global() *Value {
	return &Value{
		value: js.Global(),
	}
}

func ValueOf(value interface{}) *Value {
	switch v := value.(type) {
	case map[string]interface{}:
		return &Value{
			value: js.ValueOf(v),
		}
	case []interface{}:
		return &Value{
			value: js.ValueOf(v),
		}
	case string:
		return &Value{
			value: js.ValueOf(v),
		}
	case bool:
		return &Value{
			value: js.ValueOf(v),
		}
	case int, float64:
		return &Value{
			value: js.ValueOf(v),
		}
	case nil:
		return &Value{
			value: js.Null(),
		}
	default:
		return &Value{
			value: v,
		}
	}
}

func FuncOf(fn func(this *Value, args []*Value) interface{}) *Func {
	return &Func{
		value: js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return fn(&Value{value: this}, convertArgs(args))
		}),
	}
}

func convertArgs(args []js.Value) []*Value {
	converted := make([]*Value, len(args))
	for i, v := range args {
		converted[i] = &Value{value: v}
	}
	return converted
}

func CopyBytesToJS(dest *Value, src []byte) {
	js.CopyBytesToJS(dest.value.(js.Value), src)
}

func CopyBytesToGo(dest []byte, src *Value) {
	js.CopyBytesToGo(dest, src.value.(js.Value))
}
