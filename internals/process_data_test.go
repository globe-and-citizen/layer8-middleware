package internals

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	utils "github.com/globe-and-citizen/layer8-utils"
	"github.com/stretchr/testify/assert"
)

func TestProcessData(t *testing.T) {
	pri, pub, err := utils.GenerateKeyPair(utils.ECDH)
	assert.Nil(t, err)
	assert.NotNil(t, pri)
	assert.NotNil(t, pub)

	shared, err := pri.GetECDHSharedSecret(pub)
	assert.Nil(t, err)
	assert.NotNil(t, shared)

	pri2, pub2, err := utils.GenerateKeyPair(utils.ECDH)
	assert.Nil(t, err)
	assert.NotNil(t, pri2)
	assert.NotNil(t, pub2)

	shared2, err := pri2.GetECDHSharedSecret(pub2)
	assert.Nil(t, err)
	assert.NotNil(t, shared2)

	encrypt := func(data *utils.Request, key *utils.JWK) string {
		b, err := data.ToJSON()
		assert.Nil(t, err)

		enc, err := key.SymmetricEncrypt(b)
		assert.Nil(t, err)

		d := map[string]string{
			"data": base64.URLEncoding.EncodeToString(enc),
		}
		enc, err = json.Marshal(d)
		assert.Nil(t, err)

		return string(enc)
	}

	bodybytes := func(data map[string]interface{}) []byte {
		b, err := json.Marshal(data)
		assert.Nil(t, err)

		return b
	}

	tests := []struct {
		name         string
		key          *utils.JWK
		rawData      string
		expectedBody []byte
		expectError  bool // if true, a response object is returned
	}{
		{
			name: "process_data_with_valid_key",
			key:  shared,
			rawData: encrypt(&utils.Request{
				Method: "GET",
				Headers: map[string]string{
					"x-test": "test",
				},
				Body: bodybytes(map[string]interface{}{
					"test": "test",
				}),
			}, shared),
			expectedBody: bodybytes(map[string]interface{}{
				"test": "test",
			}),
			expectError: false,
		},
		{
			name: "process_data_with_invalid_key",
			key:  shared2,
			rawData: encrypt(&utils.Request{
				Method: "GET",
				Headers: map[string]string{
					"x-test": "test",
				},
				Body: bodybytes(map[string]interface{}{
					"test": "test",
				}),
			}, shared),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, request := ProcessData(tt.rawData, tt.key)
			if tt.expectError {
				assert.NotNil(t, response)
				assert.Nil(t, request)
			} else {
				assert.Nil(t, response)
				assert.NotNil(t, request)
				assert.Equal(t, tt.expectedBody, request.Body)
			}
		})
	}
}
