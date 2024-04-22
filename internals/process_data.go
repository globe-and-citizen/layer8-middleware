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

// ProcessData decrypts the request body and processes it. It then sets the method and
// headers for the request and appends the data to the FormData object
//
// It returns a Response object and a Request object. When the Response object is not nil,
// it means that an error occurred and the request should be stopped.
func ProcessData(rawdata string, headers *js.Value, key *utils.JWK, fd *js.Formdata) (
	*utils.Response, *utils.Request,
) {
	response := new(utils.Response)
	// parse body and decrypt the "data" field
	var enc map[string]interface{}
	json.Unmarshal([]byte(rawdata), &enc)

	data, err := base64.URLEncoding.DecodeString(enc["data"].(string))
	if err != nil {
		fmt.Println("error decoding request:", err.Error())
		response.Status = 500
		response.StatusText = "Could not decode request: " + err.Error()
		return response, nil
	}

	b, err := key.SymmetricDecrypt(data)
	if err != nil {
		fmt.Println("error decrypting request:", err.Error())
		response.Status = 500
		response.StatusText = "Could not decrypt request: " + err.Error()
		return response, nil
	}

	// parse the decrypted data into a request object
	jreq, err := utils.FromJSONRequest(b)
	if err != nil {
		fmt.Println("error serializing json request:", err.Error())
		response.Status = 500
		response.StatusText = "Could not decode request: " + err.Error()
		return response, nil
	}

	switch strings.ToLower(jreq.Headers["Content-Type"]) {
	case "application/layer8.buffer+json": // this is used for multipart/form-data
		var reqBody map[string]interface{}

		json.Unmarshal(jreq.Body, &reqBody)

		randomBytes := make([]byte, 16)
		_, err = rand.Read(randomBytes)
		if err != nil {
			fmt.Println("error generating random bytes:", err.Error())
			response.Status = 500
			response.StatusText = "Could not generate random bytes: " + err.Error()
			return response, nil
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
						response.Status = 500
						response.StatusText = "Could not decode file buffer: " + err.Error()
						return response, nil
					}

					file := js.File{
						Size: val["size"].(float64),
						Name: val["name"].(string),
						Type: val["type"].(string),
						Buff: buff,
					}
					fd.AppendFile(k, file)
				case "String":
					fd.Append(k, val["value"], js.TypeString)
				case "Number":
					fd.Append(k, val["value"], js.TypeNumber)
				case "Boolean":
					fd.Append(k, val["value"], js.TypeBoolean)
				}
			}
		}

		jreq.Headers = map[string]string{
			"Content-Type": "multipart/form-data; boundary=" + boundary,
		}
	default:
		jreq.Headers = map[string]string{
			"Content-Type": "application/json",
		}
	}

	return nil, jreq
}
