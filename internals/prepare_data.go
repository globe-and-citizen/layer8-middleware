package internals

import (
	"encoding/json"
	"fmt"

	"globe-and-citizen/layer8/middleware/js"

	utils "github.com/globe-and-citizen/layer8-utils"
)

func PrepareData(res, data *js.Value, symmKey *utils.JWK, jwt string) *utils.Response {
	var (
		b   []byte
		err error
	)

	switch data.Type {
	case js.TypeObject:
		b, err = json.Marshal(data.GetValue().(map[string]interface{}))
		if err != nil {
			println("error serializing json response:", err.Error())
			return &utils.Response{
				Status:     500,
				StatusText: "Could not encode response",
			}
		}
	case js.TypeArray:
		b, err = json.Marshal(data.GetValue().([]interface{}))
		if err != nil {
			println("error serializing json response:", err.Error())
			return &utils.Response{
				Status:     500,
				StatusText: "Could not encode response",
			}
		}
	case js.TypeString:
		b = []byte(data.String())
	case js.TypeNumber:
		b = []byte(fmt.Sprintf("%f", data.Number()))
	case js.TypeBoolean:
		b = []byte(fmt.Sprintf("%t", data.Bool()))
	default:
		b = []byte(fmt.Sprintf("%v", data.GetValue()))
	}

	// Encrypt response
	jres := utils.Response{
		Body:    b,
		Status:  200,
		Headers: make(map[string]string),
	}
	if res.Get("statusCode") != nil {
		jres.Status = int(res.Get("statusCode").(float64))
	}
	if res.Get("statusText") != nil {
		jres.StatusText = res.Get("statusText").(string)
	}

	if res.Get("headers") == nil {
		res.Set("headers", map[string]*js.Value{})
	}
	for k, v := range res.Get("headers").(map[string]*js.Value) {
		if v.Type != js.TypeString {
			continue
		}
		jres.Headers[k] = v.Value.(string)
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
