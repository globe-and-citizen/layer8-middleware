package internals

import (
	"globe-and-citizen/layer8/middleware/js"
	"testing"

	utils "github.com/globe-and-citizen/layer8-utils"
	"github.com/stretchr/testify/assert"
)

func TestPrepareData(t *testing.T) {
	pri, pub, err := utils.GenerateKeyPair(utils.ECDH)
	assert.Nil(t, err)
	assert.NotNil(t, pri)
	assert.NotNil(t, pub)

	key, err := pri.GetECDHSharedSecret(pub)
	assert.Nil(t, err)
	assert.NotNil(t, key)

	tests := []struct {
		name string
		data *js.Value
		res  *js.Value
		want *utils.Response
	}{
		{
			name: "prepare_data_with_object_body",
			data: js.ValueOf(map[string]interface{}{
				"hello": "world",
			}),
			res: js.ValueOf(map[string]interface{}{
				"statusCode": float64(200),
				"statusText": "OK",
				"headers": map[string]interface{}{
					"x-key": "value",
				},
			}),
			want: &utils.Response{
				Body:       []byte("{\"hello\":\"world\"}"),
				Status:     200,
				StatusText: "OK",
				Headers: map[string]string{
					"x-key": "value",
				},
			},
		},
		{
			name: "prepare_data_with_array_body",
			data: js.ValueOf([]interface{}{
				"hello", "world",
			}),
			res: js.ValueOf(map[string]interface{}{
				"statusCode": float64(200),
				"statusText": "OK",
				"headers": map[string]interface{}{
					"x-key": "value",
				},
			}),
			want: &utils.Response{
				Body:       []byte("[\"hello\",\"world\"]"),
				Status:     200,
				StatusText: "OK",
				Headers: map[string]string{
					"x-key": "value",
				},
			},
		},
		{
			name: "prepare_data_with_string_body",
			data: js.ValueOf("hello world"),
			res: js.ValueOf(map[string]interface{}{
				"statusCode": float64(200),
				"statusText": "OK",
				"headers": map[string]interface{}{
					"x-key": "value",
				},
			}),
			want: &utils.Response{
				Body:       []byte("hello world"),
				Status:     200,
				StatusText: "OK",
				Headers: map[string]string{
					"x-key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := PrepareData(tt.res, tt.data, key, "test_mp_jwt")
			assert.NotNil(t, response)
			assert.Equal(t, int(tt.res.Get("statusCode").(float64)), response.Status)
			assert.Equal(t, tt.res.Get("statusText").(string), response.StatusText)

			b, err := key.SymmetricDecrypt(response.Body)
			assert.Nil(t, err)

			got, err := utils.FromJSONResponse(b)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
