package js

type (
	// FormData is a struct that represents a FormData object in JavaScript
	// This is created to make it easier to work with FormData objects in Go
	// and to make mocking FormData objects easier in tests
	//
	// Example usage:
	//
	// 	formData := FormData{
	// 		FormData: js.Global().Get("FormData").New(),
	// 		Append: func(key string, value interface{}, t Type) {
	// 			switch t {
	// 			case TypeString:
	// 				formData.FormData.Call("append", key, value.(string))
	// 			case TypeNumber:
	// 				formData.FormData.Call("append", key, value.(float64))
	// 			case TypeBoolean:
	// 				formData.FormData.Call("append", key, value.(bool))
	// 			default:
	// 				panic("invalid type")
	// 			}
	// 		},
	// 		AppendFile: func(key string, file File) {
	// 			uint8Array := js.Global().Get("Uint8Array").New(file.Size)
	// 			js.CopyBytesToJS(uint8Array, file.Buff)
	//
	// 			fileObj := js.Global().Get("File").New(
	// 				[]interface{}{uint8Array},
	// 				file.Name,
	// 				map[string]interface{}{"type": file.Type},
	// 			)
	//
	// 			formData.FormData.Call("append", key, fileObj)
	// 		},
	// 	}
	//
	// 	formData.Append("key", "value")
	// 	formData.AppendFile("file", File{
	// 		Size: 10.0,
	// 		Name: "file.txt",
	// 		Type: "text/plain",
	// 		Buff: []byte("file content"),
	// 	})
	//
	// In tests, you can mock FormData objects like this:
	//
	// 	formData := FormData{
	// 		FormData: []interface{}{},
	// 		Append: func(key, value string) {
	// 			formData.FormData = append(formData.FormData, map[string]interface{}{
	// 			 	"isFile": false,
	// 				"key": key,
	// 				"value": value,
	// 			})
	// 		},
	// 		AppendFile: func(key string, file File) {
	// 			formData.FormData = append(formData.FormData, map[string]interface{}{
	// 				"isFile": true,
	// 				"key": key,
	// 				"value": file,
	// 			})
	// 		},
	// 	}
	//
	// 	formData.Append("key", "value")
	// 	formData.AppendFile("file", File{
	// 		Size: 10.0,
	// 		Name: "file.txt",
	// 		Type: "text/plain",
	// 		Buff: []byte("file content"),
	// 	})
	//
	// 	// Now you can use formData.FormData to check the values that were appended
	//
	// This is useful for testing code that uses FormData objects without having to
	// create a real FormData object in the tests
	Formdata struct {
		FormData   interface{}
		Append     func(key string, value interface{}, t Type)
		AppendFile func(key string, file File)
	}

	// File is a struct that represents a File object in JavaScript
	File struct {
		Size float64 // Size is the size of the file in bytes
		Name string
		Type string
		Buff []byte
	}
)
