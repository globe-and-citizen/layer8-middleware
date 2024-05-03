package marshaller

import (
	"syscall/js"

	gojs "globe-and-citizen/layer8/middleware/js"
)

// Unmarshal unmarshals a `syscall/js.Value` into the internal `Value` type.
//
// Note that only primitive types are supported and will be unmarshaled
//   - Number
//   - Boolean
//   - String
//   - Object (recursively converted to map[string]interface{})
//   - Array (recursively converted to []interface{})
func Unmarshal(v js.Value) *gojs.Value {
	keys := js.Global().Get("Object").Call("keys", v)
	m := &gojs.Value{
		Type:        gojs.TypeObject,
		Constructor: v.Get("constructor").Get("name").String(),
		Value:       make(map[string]interface{}),
	}

	for i := 0; i < keys.Length(); i++ {
		key := keys.Index(i).String()
		val := v.Get(key)

		switch val.Type() {
		case js.TypeNumber:
			m.Value.(map[string]interface{})[key] = &gojs.Value{
				Type:        gojs.TypeNumber,
				Constructor: "Number",
				Value:       val.Float(),
			}
		case js.TypeBoolean:
			m.Value.(map[string]interface{})[key] = &gojs.Value{
				Type:        gojs.TypeBoolean,
				Constructor: "Boolean",
				Value:       val.Bool(),
			}
		case js.TypeString:
			m.Value.(map[string]interface{})[key] = &gojs.Value{
				Type:        gojs.TypeString,
				Constructor: "String",
				Value:       val.String(),
			}
		case js.TypeObject:
			if val.Get("constructor").Type() == js.TypeUndefined || (val.Get("constructor").Get("name").String() != "Array" &&
				val.Get("constructor").Get("name").String() != "Object") {
				continue
			}

			if val.Get("constructor").Get("name").String() == "Array" {
				m.Value.(map[string]interface{})[key] = &gojs.Value{
					Type:        gojs.TypeArray,
					Constructor: val.Get("constructor").Get("name").String(),
					Value:       parseObjectToSlice(val),
				}
			} else if val.Get("constructor").Get("name").String() == "Object" {
				m.Value.(map[string]interface{})[key] = Unmarshal(val)
			}
		default:
			m.Value.(map[string]interface{})[key] = &gojs.Value{
				Type:        gojs.TypeNull,
				Constructor: "Null",
				Value:       nil,
			}
		}
	}

	return m
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
