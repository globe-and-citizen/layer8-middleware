package js

type (
	Type  int
	Value struct {
		Type        Type
		Constructor string

		Value interface{}
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

func ValueOf(value interface{}) *Value {
	switch val := value.(type) {
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		return &Value{
			Type:        TypeNumber,
			Constructor: "Number",
			Value:       val,
		}
	case bool:
		return &Value{
			Type:        TypeBoolean,
			Constructor: "Boolean",
			Value:       val,
		}
	case string:
		return &Value{
			Type:        TypeString,
			Constructor: "String",
			Value:       val,
		}
	case map[string]interface{}:
		for k, v := range val {
			val[k] = ValueOf(v)
		}

		return &Value{
			Type:        TypeObject,
			Constructor: "Object",
			Value:       val,
		}
	case []interface{}:
		for i, v := range val {
			val[i] = ValueOf(v)
		}

		return &Value{
			Type:        TypeArray,
			Constructor: "Array",
			Value:       val,
		}
	default:
		return &Value{
			Type:        TypeObject,
			Constructor: "Object",
			Value:       val,
		}
	}
}

func (v *Value) GetValue() interface{} {
	switch v.Type {
	case TypeNumber:
		return v.Value.(float64)
	case TypeBoolean:
		return v.Value.(bool)
	case TypeString:
		return v.Value.(string)
	case TypeObject:
		val := v.Value.(map[string]interface{})
		if len(val) == 0 {
			return map[string]interface{}{}
		}

		for k, v := range val {
			val[k] = v.(*Value).GetValue()
		}
		return val
	case TypeArray:
		val := v.Value.([]interface{})
		if len(val) == 0 {
			return []interface{}{}
		}

		for i, v := range val {
			val[i] = v.(*Value).GetValue()
		}
		return val
	default:
		return nil
	}
}

func (v *Value) Get(key string) interface{} {
	val, ok := v.Value.(map[string]interface{})[key]
	if !ok {
		return nil
	}
	return val.(*Value).Value
}

func (v *Value) Set(key string, value interface{}) {
	switch val := value.(type) {
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		v.Value.(map[string]interface{})[key] = &Value{
			Type:        TypeNumber,
			Constructor: "Number",
			Value:       val,
		}
	case bool:
		v.Value.(map[string]interface{})[key] = &Value{
			Type:        TypeBoolean,
			Constructor: "Boolean",
			Value:       val,
		}
	case string:
		v.Value.(map[string]interface{})[key] = &Value{
			Type:        TypeString,
			Constructor: "String",
			Value:       val,
		}
	case map[string]interface{}:
		for k, v := range val {
			val[k] = ValueOf(v)
		}

		v.Value.(map[string]interface{})[key] = &Value{
			Type:        TypeObject,
			Constructor: "Object",
			Value:       val,
		}
	case []interface{}:
		for i, v := range val {
			val[i] = ValueOf(v)
		}

		v.Value.(map[string]interface{})[key] = &Value{
			Type:        TypeArray,
			Constructor: "Array",
			Value:       val,
		}
	default:
		v.Value.(map[string]interface{})[key] = &Value{
			Type:        TypeNull,
			Constructor: "Null",
			Value:       nil,
		}
	}
}

func (v *Value) String() string {
	if v.Type == TypeString {
		return v.Value.(string)
	}
	return ""
}

func (v *Value) Bool() bool {
	if v.Type == TypeBoolean {
		return v.Value.(bool)
	}
	return false
}

func (v *Value) Number() float64 {
	switch v.Value.(type) {
	case int:
		return float64(v.Value.(int))
	case int32:
		return float64(v.Value.(int32))
	case int64:
		return float64(v.Value.(int64))
	case uint:
		return float64(v.Value.(uint))
	case uint32:
		return float64(v.Value.(uint32))
	case uint64:
		return float64(v.Value.(uint64))
	case float32:
		return float64(v.Value.(float32))
	case float64:
		return v.Value.(float64)
	}
	return 0
}
