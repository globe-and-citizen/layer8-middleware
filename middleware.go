package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"syscall/js"

	"globe-and-citizen/layer8/middleware/internals"
	gojs "globe-and-citizen/layer8/middleware/js"
	"globe-and-citizen/layer8/middleware/marshaller"
	"globe-and-citizen/layer8/middleware/storage"

	utils "github.com/globe-and-citizen/layer8-utils"
)

const VERSION = "1.0.26"

func init() {
	var err error
	// generate key pair
	pri, pub, err := utils.GenerateKeyPair(utils.ECDH)
	if err != nil {
		panic(err)
	}

	storage.InitInMemStorage(pri, pub)
}

func main() {
	c := make(chan struct{})
	fmt.Printf("L8 WASM Middleware version %s loaded.\n\n", VERSION)
	js.Global().Set("WASMMiddleware", js.FuncOf(WASMMiddleware_v2))
	js.Global().Set("ServeStatic", js.FuncOf(static))
	js.Global().Set("ProcessMultipart", js.FuncOf(multipart))
	js.Global().Set("TestWASM", js.FuncOf(TestWASM))
	<-c
}

// WASM Middleware Version 2 Does not depend on the Express Body Parser//
func WASMMiddleware_v2(this js.Value, args []js.Value) interface{} {
	// Get the request and response objects and the next function
	var (
		req       = args[0]
		res       = args[1]
		next      = args[2]
		headers   = req.Get("headers")
		goHeaders = marshaller.Unmarshal(headers)
		db        = storage.GetInMemStorage()
	)

	// proceed to next middleware/handler request is not a layer8 request
	if headers.String() == "<undefined>" || headers.Get("x-tunnel").String() == "<undefined>" {
		next.Invoke()
		return nil
	}

	initECDH := func() interface{} {
		secret, pub, mpjwt, err := internals.InitializeECDH(goHeaders)
		if err != nil {
			println(err.Error())
			res.Set("statusCode", 500)
			res.Set("statusMessage", "Failure to initialize ECDH")
			res.Call("end", "500 Internal Server Error")
			return nil
		}

		res.Set("statusCode", 200)
		res.Set("statusMessage", "ECDH Successfully Completed!")
		res.Call("setHeader", "x-shared-secret", secret)
		res.Call("setHeader", "mp-JWT", mpjwt)
		res.Call("end", pub)
		return nil
	}

	// Decide if this is a redirect to ECDH init.
	isECDHInit := headers.Get("x-ecdh-init").String()
	clientUUID := headers.Get("x-client-uuid").String()
	if isECDHInit != "<undefined>" || clientUUID == "<undefined>" {
		return initECDH()
	}

	// continue to next middleware/handler if it's a request for static files
	if headers.Get("x-static").String() != "<undefined>" || headers.Get("X-Static").String() != "<undefined>" {
		next.Invoke()
		return nil
	}

	// Get the symmetric key for this client
	var spSymmetricKey *utils.JWK
	for _, v := range db.Keys {
		if v[clientUUID] != nil {
			spSymmetricKey = v[clientUUID]
		}
	}
	if spSymmetricKey == nil {
		return initECDH()
	}

	// Get the JWT for this client
	var MpJWT string
	for _, v := range db.JWTs {
		if v[clientUUID] != "" {
			MpJWT = v[clientUUID]
		}
	}
	if MpJWT == "" {
		return initECDH()
	}

	var body string

	req.Call("on", "data", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		body += args[0].Call("toString").String()
		return nil
	}))

	req.Call("on", "end", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Occur under all circumstances:
		response, request := internals.ProcessData(body, spSymmetricKey)
		if response != nil {
			res.Set("statusCode", response.Status)
			res.Set("statusMessage", response.StatusText)
			return nil
		}

		req.Set("method", request.Method)
		for k, v := range request.Headers {
			headers.Set(k, v)
		}

		// Primary Decisiotn Point
		switch strings.ToLower(request.Headers["Content-Type"]) {
		case "application/layer8.buffer+json": // this is used for multipart/form-data

			var reqBody map[string]interface{}

			json.Unmarshal(request.Body, &reqBody)

			// clear the body as it will be replaced by the formdata
			request.Body = nil

			// pass in reqBody and get out a formData
			formdata, err := convertBodyToFormdata(reqBody)
			if err != nil {
				fmt.Println("error decoding file buffer:", err.Error())
				res.Set("statusCode", 500)
				res.Set("statusMessage", "Could not decode file buffer: "+err.Error())
			}

			boundary, err := getArbitraryBoundary()
			if err != nil {
				fmt.Println("error generating random bytes:", err.Error())
				res.Set("statusCode", 500)
				res.Set("statusMessage", "Could not generate random bytes: "+err.Error())
			}

			request.Headers["Content-Type"] = "multipart/form-data; boundary=" + boundary

			req.Set("body", formdata)

		default:
			if contentType, ok := request.Headers["Content-Type"]; !ok || contentType == "" {
				request.Headers["Content-Type"] = "application/json"
			}
			var body map[string]interface{}
			json.Unmarshal(request.Body, &body)

			req.Set("body", body)
		}

		// continue to next middleware/handler
		next.Invoke()
		return nil
	}))

	// OVERWRITE ALL RESPONSE FUNCTIONS
	respond := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// we'll need to convert some of the args to a map
		// this is for instances where the user wants to send a json response
		// and the args are stringified json
		data := new(gojs.Value)
		if args[0].Type() == js.TypeString {
			var (
				mapData map[string]interface{}
				err     = json.Unmarshal([]byte(args[0].String()), &mapData)
			)
			if err != nil {
				// when JSON data cannot be unmarshalled, we'll just send the string as is
				// this is useful for sending plain text responses
				data = gojs.ValueOf(args[0].String())
			} else {
				data = gojs.ValueOf(mapData)
			}
		} else {
			data = marshaller.Unmarshal(args[0])
		}

		response := internals.PrepareData(marshaller.Unmarshal(res), data, spSymmetricKey, MpJWT)
		res.Set("statusCode", response.Status)
		res.Set("statusMessage", response.StatusText)
		res.Call("set", js.ValueOf(MapOfStringsToMapOfInterfaces(response.Headers)))
		res.Call("end", js.Global().Get("JSON").Call("stringify", js.ValueOf(map[string]interface{}{
			"data": base64.URLEncoding.EncodeToString(response.Body),
		})))
		return nil
	})

	res.Set("send", respond)
	res.Set("json", respond)
	return nil
}

func static(this js.Value, args []js.Value) interface{} {
	var (
		req     = args[0]
		res     = args[1]
		dir     = args[2].String()
		fs      = args[3]
		headers = req.Get("headers")
		db      = storage.GetInMemStorage()

		// returns the default EncryptedImageData
		returnEncryptedImage = func() interface{} {
			arrayBuffer := js.Global().Get("Uint8Array").New(len(EncryptedImageData))
			js.CopyBytesToJS(arrayBuffer, EncryptedImageData)

			res.Set("statusCode", 200)
			res.Set("statusMessage", "OK")
			res.Set("content-type", "image/png")
			res.Call("end", arrayBuffer)
			return nil
		}
	)

	// get the file path
	path := req.Get("url").String()
	if path == "/" {
		path = "/index.html"
	}

	path, err := url.QueryUnescape(path)
	if err != nil {
		println("error url decoding path:", err.Error())
		res.Set("statusCode", 500)
		res.Set("statusMessage", "Internal Server Error")
		res.Call("end", "500 Internal Server Error")
		return nil
	}

	path = dir + path
	exists := fs.Call("existsSync", path).Bool()
	if !exists {
		res.Set("statusCode", 404)
		res.Set("statusMessage", "Not Found")
		res.Call("end", "Cannot GET "+req.Get("url").String())
		return nil
	}

	// return the default EncryptedImageData if the request is not a layer8 request
	if headers.String() == "<undefined>" || headers.Get("x-tunnel").String() == "<undefined>" {
		return returnEncryptedImage()
	}

	clientUUID := headers.Get("x-client-uuid").String()
	if clientUUID == "<undefined>" {
		return returnEncryptedImage()
	}

	var mpJWT string
	for _, v := range db.JWTs {
		if v[clientUUID] != "" {
			mpJWT = v[clientUUID]
		}
	}

	var sym *utils.JWK
	for _, v := range db.Keys {
		if v[clientUUID] != nil {
			sym = v[clientUUID]
		}
	}
	if sym == nil {
		return returnEncryptedImage()
	}

	// read the file
	buffer := fs.Call("readFileSync", path)
	b := make([]byte, buffer.Get("length").Int())
	js.CopyBytesToGo(b, buffer)

	// create a response object
	jres := utils.Response{
		Body:       b,
		Status:     http.StatusOK,
		StatusText: http.StatusText(http.StatusOK),
		Headers: map[string]string{
			"content-type": http.DetectContentType(b),
		},
	}

	b, err = jres.ToJSON()
	if err != nil {
		println("error serializing json response:", err.Error())
		res.Set("statusCode", 500)
		res.Set("statusMessage", "Internal Server Error")
		res.Call("end", "500 Internal Server Error")
		return nil
	}

	// encrypt the file
	encrypted, err := sym.SymmetricEncrypt(b)
	if err != nil {
		println("error encrypting file:", err.Error())
		res.Set("statusCode", 500)
		res.Set("statusMessage", "Internal Server Error")
		res.Call("end", "500 Internal Server Error")
		return nil
	}

	// send the response
	res.Set("statusCode", jres.Status)
	res.Set("statusMessage", jres.StatusText)
	res.Call("set", js.ValueOf(map[string]interface{}{
		"content-type": "application/json",
		"mp-JWT":       mpJWT,
	}))
	res.Call("end", js.Global().Get("JSON").Call("stringify", js.ValueOf(map[string]interface{}{
		"data": base64.URLEncoding.EncodeToString(encrypted),
	})))
	return nil
}

func multipart(this js.Value, args []js.Value) interface{} {
	var (
		options = args[0]
		fs      = args[1]

		dest = options.Get("dest").String()
	)

	single := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var (
			req   = args[0]
			next  = args[2]
			field = args[3].String()
		)

		if dest == "" {
			dest = "tmp"
		}
		dest = strings.Trim(dest, "/")

		// if the destination directory does not exist, create it
		if !fs.Call("existsSync", dest).Bool() {
			fs.Call("mkdirSync", dest, map[string]interface{}{"recursive": true})
		}

		body := req.Get("body")
		if body.String() == "<undefined>" {
			next.Invoke()
			return nil
		}

		file := body.Call("get", field)
		if file.String() == "<undefined>" {
			next.Invoke()
			return nil
		}

		// check that file has a File constructor
		if file.Get("constructor").Get("name").String() != "File" {
			next.Invoke()
			return nil
		}

		file.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			uint8Array := js.Global().Get("Uint8Array").New(args[0])

			// write the file to the destination directory
			filePath := fmt.Sprintf("%s/%s", dest, file.Get("name").String())
			fs.Call("writeFileSync", filePath, uint8Array)

			// set the file to the request body
			req.Set("file", file)

			// continue to next middleware/handler
			next.Invoke()
			return nil
		}))

		return nil
	})

	array := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var (
			req   = args[0]
			next  = args[2]
			field = args[3].String()
		)

		if dest == "" {
			dest = "tmp"
		}
		dest = strings.Trim(dest, "/")

		// if the destination directory does not exist, create it
		if !fs.Call("existsSync", dest).Bool() {
			fs.Call("mkdirSync", dest, map[string]interface{}{"recursive": true})
		}

		body := req.Get("body")
		if body.String() == "<undefined>" {
			next.Invoke()
			return nil
		}

		files := body.Call("getAll", field)
		if files.String() == "<undefined>" {
			next.Invoke()
			return nil
		}

		// write the files to the destination directory
		fileObjs := []interface{}{}
		files.Call("forEach", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			file := args[0]
			index := args[1].Int()

			if file.Get("constructor").Get("name").String() != "File" {
				return nil
			}

			file.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				uint8Array := js.Global().Get("Uint8Array").New(args[0])

				// write the file to the destination directory
				filePath := fmt.Sprintf("%s/%s", dest, file.Get("name").String())
				fs.Call("writeFileSync", filePath, uint8Array)

				// append the file to the fileObjs slice
				fileObjs = append(fileObjs, file)

				// if all the files have been written to the destination directory
				// set the files to the request body and continue to next middleware/handler
				if index == files.Get("length").Int()-1 {
					req.Set("files", js.ValueOf(fileObjs))
					next.Invoke()
				}
				return nil
			}))

			return nil
		}))

		return nil
	})

	return map[string]interface{}{
		"single": single,
		"array":  array,
	}
}

// UTILS
func async_test_WASM(this js.Value, args []js.Value) interface{} {
	fmt.Println("Fisrt argument: ", args[0])
	fmt.Println("Second argument: ", args[1])
	var resolve_reject_internals = func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		//reject := args[1]
		go func() {
			// Main function body
			//fmt.Println(string(args[2]))
			resolve.Invoke(js.ValueOf(fmt.Sprintf("WASM Middleware version %s successfully loaded.", VERSION)))
			//reject.Invoke()
		}()
		return nil
	}
	promiseConstructor := js.Global().Get("Promise")
	promise := promiseConstructor.New(js.FuncOf(resolve_reject_internals))
	return promise
}

func TestWASM(this js.Value, args []js.Value) interface{} {
	fmt.Println("TestWasm Ran")
	return js.ValueOf("42")
}

func MapOfStringsToMapOfInterfaces(m map[string]string) map[string]interface{} {
	mi := make(map[string]interface{})
	for k, v := range m {
		mi[k] = v
	}
	return mi
}

// ADDED FUNCTIONS

func convertBodyToFormdata(reqBody map[string]interface{}) (js.Value, error) {
	formdata := js.Global().Get("FormData").New()

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
					return js.ValueOf(nil), fmt.Errorf("Could not decode file buffer: " + err.Error())
				}

				uint8Array := js.Global().Get("Uint8Array").New(val["size"].(float64))
				js.CopyBytesToJS(uint8Array, buff)

				file := js.Global().Get("File").New(
					[]interface{}{uint8Array},
					val["name"].(string),
					map[string]interface{}{
						"type": val["type"].(string),
					},
				)
				formdata.Call("append", k, file)
			case "String":
				formdata.Call("append", k, val["value"].(string))
			case "Number":
				formdata.Call("append", k, val["value"].(float64))
			case "Boolean":
				formdata.Call("append", k, val["value"].(bool))
			}
		}
	}
	return formdata, nil
}

func getArbitraryBoundary() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("unable to generate random bytes")
	}
	return fmt.Sprintf("----Layer8FormBoundary%s", base64.StdEncoding.EncodeToString(randomBytes)), nil
}
