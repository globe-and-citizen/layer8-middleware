package value

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
	result := &Value{
		Type:        TypeObject,
		Constructor: "Object",
		Value:       value,
	}

	switch val := value.(type) {
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		result.Type = TypeNumber
		result.Constructor = "Number"
		result.Value = val
	case bool:
		result.Type = TypeBoolean
		result.Constructor = "Boolean"
		result.Value = val
	case string:
		result.Type = TypeString
		result.Constructor = "String"
		result.Value = val
	case map[string]interface{}:
		obj := make(map[string]*Value, len(val))
		for k, v := range val {
			obj[k] = ValueOf(v)
		}

		result.Type = TypeObject
		result.Constructor = "Object"
		result.Value = obj
	case []interface{}:
		arr := make([]*Value, len(val))
		for i, v := range val {
			arr[i] = ValueOf(v)
		}

		result.Type = TypeArray
		result.Constructor = "Array"
		result.Value = arr
	}
	return result
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
		val := v.Value.(map[string]*Value)
		result := make(map[string]interface{}, len(val))

		for k, v := range val {
			result[k] = v.GetValue()
		}
		return result
	case TypeArray:
		val := v.Value.([]*Value)
		result := make([]interface{}, len(val))

		for i, v := range val {
			result[i] = v.GetValue()
		}
		return result
	default:
		return nil
	}
}

func (v *Value) Get(key string) interface{} {
	val, ok := v.Value.(map[string]*Value)[key]
	if !ok {
		return nil
	}
	return val.Value
}

func (v *Value) FullGet(key string) *Value {
	val, ok := v.Value.(map[string]*Value)[key]
	if !ok {
		return nil
	}
	return val
}

func (v *Value) Set(key string, value interface{}) {
	switch val := value.(type) {
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		v.Value.(map[string]*Value)[key] = &Value{
			Type:        TypeNumber,
			Constructor: "Number",
			Value:       val,
		}
	case bool:
		v.Value.(map[string]*Value)[key] = &Value{
			Type:        TypeBoolean,
			Constructor: "Boolean",
			Value:       val,
		}
	case string:
		v.Value.(map[string]*Value)[key] = &Value{
			Type:        TypeString,
			Constructor: "String",
			Value:       val,
		}
	case map[string]interface{}:
		for k, v := range val {
			val[k] = ValueOf(v)
		}

		v.Value.(map[string]*Value)[key] = &Value{
			Type:        TypeObject,
			Constructor: "Object",
			Value:       val,
		}
	case []interface{}:
		for i, v := range val {
			val[i] = ValueOf(v)
		}

		v.Value.(map[string]*Value)[key] = &Value{
			Type:        TypeArray,
			Constructor: "Array",
			Value:       val,
		}
	case map[string]*Value:
		v.Value.(map[string]*Value)[key] = &Value{
			Type:        TypeObject,
			Constructor: "Object",
			Value:       val,
		}
	default:
		v.Value.(map[string]*Value)[key] = &Value{
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
