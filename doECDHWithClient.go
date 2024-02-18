package main

import (
	"fmt"
	"syscall/js" // Internals defined by NOT needing this import

	utils "github.com/globe-and-citizen/layer8-utils"
)

func doECDHWithClient(nodeRequestObject, nodeResponseObject js.Value) {
	fmt.Println("[Middleware] ECDH Initialized")

	// Harvest Necessary Headers
	headers := nodeRequestObject.Get("headers")
	if headers.String() == "<undefined>" || headers.String() == "" {
		fmt.Println("[Middleware] headers on ECDH Init Request are '<undefined>' or and empty string.")
		err := fmt.Errorf("Necessary Headers not received to complete ECDH Init.")
		send500Error(nodeResponseObject, err)
	}

	clientUUID := headers.Get("x-client-uuid").String()
	if clientUUID == "<undefined>" || clientUUID == "" {
		err := fmt.Errorf("clientUUID not received to complete ECDH Init.")
		send500Error(nodeResponseObject, err)
	}
	fmt.Println("[Middleware] clientUUID: ", clientUUID)

	MpJWT := headers.Get("mp-jwt").String()
	if MpJWT == "<undefined>" || MpJWT == "" {
		err := fmt.Errorf("mp-jwt not received to complete init ECDH.")
		send500Error(nodeResponseObject, err)
	}
	UUIDMapOfJWTs = append(UUIDMapOfJWTs, map[string]string{clientUUID: MpJWT})
	fmt.Println("[Middleware] mp-jwt: ", MpJWT)

	userPubJWK := headers.Get("x-ecdh-init").String()
	if userPubJWK == "<undefined>" || userPubJWK == "" {
		err := fmt.Errorf("x-ecdh-init not received to complete ECDH Init.")
		send500Error(nodeResponseObject, err)
		return
	}
	fmt.Println("[Middleware] userPubJWK: ", userPubJWK)

	// Derive Shared Secret
	userPubJWKConverted, err := utils.B64ToJWK(userPubJWK)
	if err != nil {
		send500Error(nodeResponseObject, err)
		return
	}
	sharedSecret, err := privKey_ECDH.GetECDHSharedSecret(userPubJWKConverted)
	if err != nil {
		send500Error(nodeResponseObject, err)
		return
	}
	UUIDMapOfKeys = append(UUIDMapOfKeys, map[string]*utils.JWK{clientUUID: sharedSecret})

	// Set Headers, Get Server's Public Key, Send Response
	nodeResponseObject.Set("statusCode", 200)
	nodeResponseObject.Set("statusMessage", "ECDH Successfully Completed!")
	nodeResponseObject.Call("setHeader", "mp-JWT", MpJWT)
	server_pubKeyECDH, _ := pubKey_ECDH.ExportAsBase64()
	nodeResponseObject.Call("end", server_pubKeyECDH)

	fmt.Println("[Middleware] ECDH successfully initialized")
	return
}

// This legacy code demonstrates how the send function can be overwritten and then latter called.
// Maybe a pattern to recreated to form internals?

// func doECDHWithClient(nodeRequestObject, nodeResponseObject js.Value) {
// 	fmt.Println("[Middleware] ECDH Initialized")
// 	headers := nodeRequestObject.Get("headers")
// 	fmt.Println("[Middleware] headers on ECDH Init Request: ", headers)

// 	if headers.Get("x-ecdh-inat") == "<undefined>" {
// 		err := fmt.Errorf("Header 'x-ecdh-init' not received")
// 		send500Error("Failure to decode userPubJWK", err)
// 		return
// 	}

// 	userPubJWK := headers.Get("x-ecdh-init").String()
// 	userPubJWKConverted, err := utils.B64ToJWK(userPubJWK)
// 	if err != nil {
// 		send500Error(nodeResponseObject, err)
// 		return
// 	}

// 	clientUUID := headers.Get("x-client-uuid").String()
// 	fmt.Println("clientUUID: ", clientUUID)

// 	ss, err := privKey_ECDH.GetECDHSharedSecret(userPubJWKConverted)
// 	if err != nil {
// 		send500Error(nodeResponseObject, err)
// 		return
// 	}

// 	UUIDMapOfKeys = append(UUIDMapOfKeys, map[string]*utils.JWK{clientUUID: ss})

// 	MpJWT := headers.Get("mp-jwt").String()
// 	fmt.Println("MpJWT at SP BE (Middleware): ", MpJWT)

// 	UUIDMapOfJWTs = append(UUIDMapOfJWTs, map[string]string{clientUUID: MpJWT})

// 	nodeResponseObject.Set("send", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
// 		// encrypt response
// 		jres := utils.Response{}
// 		jres.Status = 200
// 		jres.StatusText = "ECDH Successfully Completed!"

// 		// send response
// 		nodeResponseObject.Set("statusCode", jres.Status)
// 		nodeResponseObject.Set("statusMessage", jres.StatusText)
// 		server_pubKeyECDH, _ := pubKey_ECDH.ExportAsBase64()

// 		nodeResponseObject.Call("end", server_pubKeyECDH) //It's almost as if, this overwrites the body of line 98
// 		return nil
// 	}))

// 	nodeResponseObject.Call("setHeader", "mp-JWT", MpJWT)
// 	nodeResponseObject.Call("send")
// 	fmt.Println("[Middleware] ECDH successfully initialized")
// 	return
// }
