package internals

import (
	"errors"
	"strings"

	"globe-and-citizen/layer8/middleware/js"
	"globe-and-citizen/layer8/middleware/storage"

	utils "github.com/globe-and-citizen/layer8-utils"
)

// InitializeECDH initializes the ECDH key exchange
//
// Arguments:
//   - request: the request object
//
// Returns:
//   - sharedSecret: the shared secret
//   - pub: the server public key
//   - mpJWT: the JWT
//   - error: an error if the function fails
func InitializeECDH(headers *js.Value) (string, string, string, error) {
	db := storage.GetInMemStorage()

	// validation
	required := map[string]js.Type{
		"x-ecdh-init":   js.TypeString,
		"x-client-uuid": js.TypeString,
		"mp-jwt":        js.TypeString,
	}
	missing := []string{}
	for k := range required {
		if headers.Get(k) == nil {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return "", "", "", errors.New("missing required headers: " + strings.Join(missing, ", "))
	}

	invalid := []string{}
	for k, v := range required {
		if headers.FullGet(k).Type != v {
			invalid = append(invalid, k)
		}
	}
	if len(invalid) > 0 {
		return "", "", "", errors.New("invalid headers: " + strings.Join(invalid, ", "))
	}

	userPubJWK, err := utils.B64ToJWK(headers.Get("x-ecdh-init").(string))
	if err != nil {
		return "", "", "", errors.New("failure to decode userPubJWK: " + err.Error())
	}

	clientUUID := headers.Get("x-client-uuid").(string)

	ss, err := db.ECDH.GetPrivateKey().GetECDHSharedSecret(userPubJWK)
	if err != nil {
		return "", "", "", errors.New("unable to get ECDH shared secret: " + err.Error())
	}
	db.Keys.Add(clientUUID, ss)

	sharedSecret, err := ss.ExportAsBase64()
	if err != nil {
		return "", "", "", errors.New("unable to export shared secret as base64: " + err.Error())
	}

	pub, err := db.ECDH.GetPublicKey().ExportAsBase64()
	if err != nil {
		return "", "", "", errors.New("unable to export public key as base64: " + err.Error())
	}

	mpJWT := headers.Get("mp-jwt").(string)
	db.JWTs.Add(clientUUID, mpJWT)

	return sharedSecret, pub, mpJWT, nil
}
