package internals

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"globe-and-citizen/layer8/middleware/js"

	utils "github.com/globe-and-citizen/layer8-utils"
)

func ProcessData(req, res, headers, next *js.Value, body string, symmKey *utils.JWK) interface{} {
	// parse body and decrypt the "data" field
	var enc map[string]interface{}
	json.Unmarshal([]byte(body), &enc)

	data, err := base64.URLEncoding.DecodeString(enc["data"].(string))
	if err != nil {
		fmt.Println("error decoding request:", err.Error())
		res.Set("statusText", "Could not decode request: "+err.Error())
		res.Set("statusCode", 500)
		return nil
	}

	b, err := symmKey.SymmetricDecrypt(data)
	if err != nil {
		fmt.Println("error decrypting request:", err.Error())
		res.Set("statusText", "Could not decrypt request: "+err.Error())
		res.Set("statusCode", 500)
		return nil
	}

	// parse the decrypted data into a request object
	jreq, err := utils.FromJSONRequest(b)
	if err != nil {
		fmt.Println("error serializing json request:", err.Error())
		res.Set("statusText", "Could not decode request: "+err.Error())
		res.Set("statusCode", 500)
		return nil
	}

	switch strings.ToLower(jreq.Headers["Content-Type"]) {
	case "application/layer8.buffer+json": // this is used for multipart/form-data
		var (
			reqBody  map[string]interface{}
			formData = js.Global().Get("FormData").New()
		)

		json.Unmarshal(jreq.Body, &reqBody)

		randomBytes := make([]byte, 16)
		_, err = rand.Read(randomBytes)
		if err != nil {
			fmt.Println("error generating random bytes:", err.Error())
			res.Set("statusCode", 500)
			res.Set("statusMessage", "Could not generate random bytes: "+err.Error())
			return nil
		}
		boundary := fmt.Sprintf("----Layer8FormBoundary%s", base64.StdEncoding.EncodeToString(randomBytes))

		for k, v := range reqBody {
			// formdata can have multiple entries with the same key
			// that is why each key from the interceptor is a slice
			// of maps containing all the values for that key
			// hence the O(n^2) complexity (i.e. 2 for loops)
			for _, val := range v.([]interface{}) {
				val := val.(map[string]interface{})

				switch val["_type"].(string) {
				case "File":
					buff, err := base64.StdEncoding.DecodeString(val["buff"].(string))
					if err != nil {
						fmt.Println("error decoding file buffer:", err.Error())
						res.Set("statusCode", 500)
						res.Set("statusMessage", "Could not decode file buffer: "+err.Error())
						return nil
					}

					// converting the byte array to a uint8array so that it can be sent to the next
					// handler as a file object
					uInt8Array := js.Global().Get("Uint8Array").New(val["size"].(float64))
					js.CopyBytesToJS(uInt8Array, buff)

					file := js.Global().Get("File").New(
						[]interface{}{uInt8Array},
						val["name"].(string),
						map[string]interface{}{"type": val["type"].(string)},
					)
					formData.Call("append", k, file)
				case "String":
					formData.Call("append", k, val["value"].(string))
				case "Number":
					formData.Call("append", k, val["value"].(float64))
				case "Boolean":
					formData.Call("append", k, val["value"].(bool))
				}
			}
		}

		headers.Set("Content-Type", "multipart/form-data; boundary="+boundary)
		req.Set("body", formData)
	default:
		var reqBody map[string]interface{}
		json.Unmarshal(jreq.Body, &reqBody)

		req.Set("body", reqBody)
		headers.Set("Content-Type", "application/json")
	}

	// set the method and headers
	req.Set("method", jreq.Method)
	for k, v := range jreq.Headers {
		if strings.ToLower(k) == "content-type" {
			continue
		}
		headers.Set(k, v)
	}

	// continue to next middleware/handler
	next.Invoke()
	return nil
}
