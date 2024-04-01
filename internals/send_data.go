package internals

import (
	"encoding/base64"
	"encoding/json"

	"globe-and-citizen/layer8/middleware/js"

	utils "github.com/globe-and-citizen/layer8-utils"
)

func SendData(res, data *js.Value, symmKey *utils.JWK, jwt string) interface{} {
	var (
		b   []byte
		err error
	)

	if data.Type() == js.TypeObject {
		switch data.Get("constructor").Get("name").String() {
		case "Object":
			b, err = json.Marshal(parseJSObjectToMap(data))
			if err != nil {
				println("error serializing json response:", err.Error())
				res.Set("statusCode", 500)
				res.Set("statusMessage", "Could not encode response")
				return nil
			}
		case "Array":
			b, err = json.Marshal(parseJSObjectToSlice(data))
			if err != nil {
				println("error serializing json response:", err.Error())
				res.Set("statusCode", 500)
				res.Set("statusMessage", "Could not encode response")
				return nil
			}
		default:
			b = []byte(data.String())
		}
	} else {
		b = []byte(data.String())
	}

	// Encrypt response
	jres := utils.Response{}
	jres.Body = b
	jres.Status = res.Get("statusCode").Int()
	jres.StatusText = res.Get("statusMessage").String()
	jres.Headers = make(map[string]string)
	if res.Get("headers").String() == "<undefined>" {
		res.Set("headers", js.ValueOf(map[string]interface{}{}))
	}
	js.Global().Get("Object").Call("keys", res.Get("headers")).Call("forEach", js.FuncOf(func(this *js.Value, args []*js.Value) interface{} {
		jres.Headers[args[0].String()] = args[1].String()
		return nil
	}))
	b, err = jres.ToJSON()
	if err != nil {
		println("error serializing json response:", err.Error())
		res.Set("statusCode", 500)
		res.Set("statusMessage", "Could not encode response")
		return nil
	}

	b, err = symmKey.SymmetricEncrypt(b)
	if err != nil {
		println("error encrypting response:", err.Error())
		res.Set("statusCode", 500)
		res.Set("statusMessage", "Could not encrypt response")
		return nil
	}

	// RAVI THIS WAS ADDED
	resHeaders := make(map[string]interface{})
	for k, v := range jres.Headers {
		resHeaders[k] = v
	}
	//resHeaders["mp-jwt"] = MpJWT

	// Send response
	res.Set("statusCode", jres.Status)
	res.Set("statusMessage", jres.StatusText)
	res.Call("set", js.ValueOf(map[string]interface{}{
		"content-type": "application/json",
		"mp-JWT":       jwt, //RAVI notice this addition too.... is this what daniel is referring too...
	}))
	res.Call("end", js.Global().Get("JSON").Call("stringify", js.ValueOf(map[string]interface{}{
		"data": base64.URLEncoding.EncodeToString(b),
	})))

	return nil
}

func parseJSObjectToMap(obj *js.Value) map[string]interface{} {
	m := map[string]interface{}{}

	keys := js.Global().Get("Object").Call("keys", obj)
	for i := 0; i < keys.Length(); i++ {
		key := keys.Index(i).String()
		val := obj.Get(key)

		switch val.Type() {
		case js.TypeNumber:
			m[key] = val.Float()
		case js.TypeBoolean:
			m[key] = val.Bool()
		case js.TypeString:
			m[key] = val.String()
		case js.TypeObject:
			if val.Get("constructor").Get("name").String() == "Array" {
				m[key] = parseJSObjectToSlice(val)
				continue
			}
			m[key] = parseJSObjectToMap(val)
		}
	}

	return m
}

func parseJSObjectToSlice(obj *js.Value) []interface{} {
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
				s = append(s, parseJSObjectToSlice(val))
				continue
			}
			s = append(s, parseJSObjectToMap(val))
		}
	}

	return s
}
