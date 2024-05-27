package internals

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	utils "github.com/globe-and-citizen/layer8-utils"
)

// ProcessData decodes and decrypts the request body. It returns a Response object
// and a Request object. When the Response object is not nil, it means that an error
// occurred and the request should be stopped.
func ProcessData(rawdata string, key *utils.JWK) (*utils.Response, *utils.Request) {
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
	if jreq.Headers == nil {
		jreq.Headers = make(map[string]string)
	}

	return nil, jreq
}
