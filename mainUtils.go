package main

import (
	"fmt"
	"syscall/js"
)

// TODO Probably a more general Error?
// send500Error(...)
// send404Error(...)
// send(...)
// sendGeneralError(...)

func send500Error(nodeResponseObject js.Value, err error) {
	fmt.Println("[Middleware]", err.Error())
	nodeResponseObject.Set("statusCode", 500)
	nodeResponseObject.Set("statusMessage", "Internal Server Error")
	nodeResponseObject.Call("end", "Default response body. Internal Server Error")
}
