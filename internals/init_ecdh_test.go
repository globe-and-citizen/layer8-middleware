package internals

import (
	"globe-and-citizen/layer8/middleware/js"
	"globe-and-citizen/layer8/middleware/storage"
	"strings"
	"testing"

	utilities "github.com/globe-and-citizen/layer8-utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInitializeECDH(t *testing.T) {
	serverPri, serverPub, err := utilities.GenerateKeyPair(utilities.ECDH)
	assert.Nil(t, err)
	assert.NotNil(t, serverPri)
	assert.NotNil(t, serverPub)

	storage.InitInMemStorage(serverPri, serverPub)
	db := storage.GetInMemStorage()
	assert.NotNil(t, db)

	b64ServerPub, err := serverPub.ExportAsBase64()
	assert.Nil(t, err)
	assert.NotEmpty(t, b64ServerPub)

	clientPri, clientPub, err := utilities.GenerateKeyPair(utilities.ECDH)
	assert.Nil(t, err)
	assert.NotNil(t, clientPri)
	assert.NotNil(t, clientPub)

	b64ClientPub, err := clientPub.ExportAsBase64()
	assert.Nil(t, err)
	assert.NotEmpty(t, b64ClientPub)

	shared, err := serverPri.GetECDHSharedSecret(clientPub)
	assert.Nil(t, err)
	assert.NotNil(t, shared)

	b64Shared, err := shared.ExportAsBase64()
	assert.Nil(t, err)
	assert.NotEmpty(t, b64Shared)

	mpjwt, err := utilities.GenerateStandardToken(uuid.New().String())
	assert.Nil(t, err)
	assert.NotEmpty(t, mpjwt)

	type args struct {
		headers *js.Value
	}
	tests := []struct {
		name          string
		args          args
		wantShared    string
		wantPub       string
		wantMpJWT     string
		wantErr       bool
		wantErrString string
	}{
		{
			name: "init_with_valid_headers",
			args: args{
				headers: js.ValueOf(map[string]interface{}{
					"x-ecdh-init":   b64ClientPub,
					"x-client-uuid": uuid.New().String(),
					"mp-jwt":        mpjwt,
				}),
			},
			wantShared: b64Shared,
			wantPub:    b64ServerPub,
			wantMpJWT:  mpjwt,
			wantErr:    false,
		},
		{
			name: "init_with_invalid_x_ecdh_init",
			args: args{
				headers: js.ValueOf(map[string]interface{}{
					"x-ecdh-init":   "invalid",
					"x-client-uuid": uuid.New().String(),
					"mp-jwt":        mpjwt,
				}),
			},
			wantErr:       true,
			wantErrString: "failure to decode userPubJWK:",
		},
		{
			name: "init_with_no_headers",
			args: args{
				headers: js.ValueOf(map[string]interface{}{}),
			},
			wantErr:       true,
			wantErrString: "missing required headers:",
		},
		{
			name: "init_with_missing_headers",
			args: args{
				headers: js.ValueOf(map[string]interface{}{
					"x-client-uuid": uuid.New().String(),
				}),
			},
			wantErr:       true,
			wantErrString: "missing required headers:",
		},
		{
			name: "init_with_invalid_header_types",
			args: args{
				headers: js.ValueOf(map[string]interface{}{
					"x-ecdh-init":   123,
					"x-client-uuid": 123,
					"mp-jwt":        123,
				}),
			},
			wantErr:       true,
			wantErrString: "invalid headers:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShared, gotPub, gotMpJWT, err := InitializeECDH(tt.args.headers)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Subset(t, strings.Split(err.Error(), " "), strings.Split(tt.wantErrString, " "))
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantShared, gotShared)
				assert.Equal(t, tt.wantPub, gotPub)
				assert.Equal(t, tt.wantMpJWT, gotMpJWT)

				savedShared := db.Keys.Get(tt.args.headers.Get("x-client-uuid").(string))
				assert.NotNil(t, savedShared)

				b64Shared, err := savedShared.ExportAsBase64()
				assert.Nil(t, err)
				assert.NotEmpty(t, b64Shared)
				assert.Equal(t, b64Shared, gotShared)

				savedMpJWT := db.JWTs.Get(tt.args.headers.Get("x-client-uuid").(string))
				assert.NotEmpty(t, savedMpJWT)
				assert.Equal(t, savedMpJWT, gotMpJWT)
			}
		})
	}
}
