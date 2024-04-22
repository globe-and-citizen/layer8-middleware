package internals

import (
	"encoding/json"

	"globe-and-citizen/layer8/middleware/js"

	utils "github.com/globe-and-citizen/layer8-utils"
)

func PrepareData(res, data *js.Value, symmKey *utils.JWK, jwt string) *utils.Response {
	var (
		b   []byte
		err error
	)

	if data.Type() == js.TypeObject {
		switch data.Constructor() {
		case "Object":
			b, err = json.Marshal(data.Value().(map[string]interface{}))
			if err != nil {
				println("error serializing json response:", err.Error())
				return &utils.Response{
					Status:     500,
					StatusText: "Could not encode response",
				}
			}
		case "Array":
			b, err = json.Marshal(data.Value().([]interface{}))
			if err != nil {
				println("error serializing json response:", err.Error())
				return &utils.Response{
					Status:     500,
					StatusText: "Could not encode response",
				}
			}
		default:
			b = []byte(data.String())
		}
	} else {
		b = []byte(data.String())
	}

	// Encrypt response
	jres := utils.Response{
		Body:       b,
		Status:     200,
		Headers:    make(map[string]string),
	}
	if res.Get("statusCode") != nil {
		jres.Status = int(res.Get("statusCode").(float64))
	}
	if res.Get("statusText") != nil {
		jres.StatusText = res.Get("statusText").(string)
	}

	if res.Get("headers") == nil {
		res.Set("headers", map[string]interface{}{})
	}
	for k, v := range res.Get("headers").(map[string]interface{}) {
		jres.Headers[k] = v.(string)
	}

	b, err = jres.ToJSON()
	if err != nil {
		println("error serializing json response:", err.Error())
		return &utils.Response{
			Status:     500,
			StatusText: "Could not encode response",
		}
	}

	b, err = symmKey.SymmetricEncrypt(b)
	if err != nil {
		println("error encrypting response:", err.Error())
		return &utils.Response{
			Status:     500,
			StatusText: "Could not encrypt response",
		}
	}

	resHeaders := make(map[string]interface{})
	for k, v := range jres.Headers {
		resHeaders[k] = v
	}

	return &utils.Response{
		Body:       b,
		Status:     jres.Status,
		StatusText: jres.StatusText,
		Headers: map[string]string{
			"content-type": "application/json",
			"mp-JWT":       jwt,
		},
	}
}
