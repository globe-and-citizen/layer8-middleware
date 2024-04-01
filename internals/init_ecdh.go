package internals

import (
	"fmt"

	"globe-and-citizen/layer8/middleware/js"
	"globe-and-citizen/layer8/middleware/storage"

	utils "github.com/globe-and-citizen/layer8-utils"
)

func InitializeECDH(request, response *js.Value) {
	db := storage.GetInMemStorage()

	headers := request.Get("headers")
	userPubJWK, err := utils.B64ToJWK(headers.Get("x-ecdh-init").String())
	if err != nil {
		fmt.Println("Failure to decode userPubJWK", err.Error())
		return
	}

	clientUUID := headers.Get("x-client-uuid").String()

	ss, err := db.ECDH.GetPrivateKey().GetECDHSharedSecret(userPubJWK)
	if err != nil {
		fmt.Println("Unable to get ECDH shared secret", err.Error())
		return
	}
	db.Keys.Add(clientUUID, ss)

	sharedSecret, err := ss.ExportAsBase64()
	if err != nil {
		fmt.Println("Unable to export shared secret as base64", err.Error())
		return
	}

	mpJWT := headers.Get("mp-jwt").String()
	db.JWTs.Add(clientUUID, mpJWT)

	response.Set("send", js.FuncOf(func(this *js.Value, args []*js.Value) interface{} {
		// encrypt response
		jres := utils.Response{}
		jres.Body = []byte(sharedSecret)
		jres.Status = 200
		jres.StatusText = "ECDH Successfully Completed!"
		// jres.Headers = make(map[string]string)
		// jres.Headers["x-shared-secret"] = sharedSecret

		if err != nil {
			println("error serializing json response:", err.Error())
			response.Set("statusCode", 500)
			response.Set("statusMessage", "Failure to encode ECDH init response")
			return nil
		}

		// send response
		response.Set("statusCode", jres.Status)
		response.Set("statusMessage", jres.StatusText)

		serverPub, _ := db.ECDH.GetPublicKey().ExportAsBase64()

		response.Call("end", serverPub)
		return nil
	}))

	response.Call("setHeader", "x-shared-secret", sharedSecret)
	response.Call("setHeader", "mp-JWT", mpJWT)
	response.Call("send")
	return
}
