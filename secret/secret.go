package secret

import (
	"encoding/base64"
)

type DecodedSecret struct {
	Key   string
	Value string
}

type SecretInterface interface {
	DecodeSecret(key, secretData string, secrets chan DecodedSecret)
}

type Secret map[string]interface{}

func NewSecret() *Secret {
	return &Secret{}
}

func (this *Secret) DecodeSecret(key, secretData string, secrets chan DecodedSecret) {
	var value string

	// avoid wrong encoded secrets
	if decoded, err := base64.StdEncoding.DecodeString(secretData); err == nil {
		value = string(decoded)
	} else {
		value = secretData
	}
	secrets <- DecodedSecret{Key: key, Value: value}
}
