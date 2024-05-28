package internals

import (
	"errors"
	"strings"

	"globe-and-citizen/layer8/middleware/storage"
	js "globe-and-citizen/layer8/middleware/utils/value"

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
func InitializeECDH(headers *js.Value, db storage.Storage) (string, string, string, error) {
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

	ss, err := storage.ECDH.GetPrivateKey().GetECDHSharedSecret(userPubJWK)
	if err != nil {
		return "", "", "", errors.New("unable to get ECDH shared secret: " + err.Error())
	}
	db.AddKey(clientUUID, ss)

	sharedSecret, err := ss.ExportAsBase64()
	if err != nil {
		return "", "", "", errors.New("unable to export shared secret as base64: " + err.Error())
	}

	pub, err := storage.ECDH.GetPublicKey().ExportAsBase64()
	if err != nil {
		return "", "", "", errors.New("unable to export public key as base64: " + err.Error())
	}

	mpJWT := headers.Get("mp-jwt").(string)
	db.AddJWT(clientUUID, mpJWT)

	return sharedSecret, pub, mpJWT, nil
}
