package marshaler

import (
	"encoding/json"

	"gopkg.in/yaml.v2"

	"ksd/secret"
)

type MarshalerInterface interface {
	IsJSON(s []byte) bool
	Unmarshal(in []byte, out interface{}, asJSON bool) error
	Marshal(d interface{}, asJSON bool) ([]byte, error)
	DecodeJSON(data map[string]interface{}) map[string]interface{}
	DecodeYaml(data map[interface{}]interface{}) map[string]interface{}
}

type Marshaler struct {
	Secret secret.SecretInterface
}

func NewMarshaler() *Marshaler {
	return &Marshaler{
		Secret: secret.NewSecret(),
	}
}

func (this *Marshaler) IsJSON(s []byte) bool {
	return json.Unmarshal(s, &json.RawMessage{}) == nil
}

func (this *Marshaler) Unmarshal(in []byte, out interface{}, isJSON bool) error {
	if isJSON {
		return json.Unmarshal(in, out)
	}
	return yaml.Unmarshal(in, out)
}

func (this *Marshaler) Marshal(d interface{}, isJSON bool) ([]byte, error) {
	if isJSON {
		return json.MarshalIndent(d, "", "    ")
	}
	return yaml.Marshal(d)
}

func (this *Marshaler) DecodeJSON(data map[string]interface{}) map[string]interface{} {
	length := len(data)
	secrets := make(chan secret.DecodedSecret, length)
	decoded := make(map[string]interface{}, length)
	for key, encoded := range data {
		go this.Secret.DecodeSecret(key, encoded.(string), secrets)
	}

	for i := 0; i < length; i++ {
		secret := <-secrets
		decoded[secret.Key] = secret.Value
	}

	return decoded
}

func (this *Marshaler) DecodeYaml(data map[interface{}]interface{}) map[string]interface{} {
	length := len(data)
	secrets := make(chan secret.DecodedSecret, length)
	decoded := make(map[string]interface{}, length)
	for key, encoded := range data {
		go this.Secret.DecodeSecret(key.(string), encoded.(string), secrets)
	}

	for i := 0; i < length; i++ {
		secret := <-secrets
		decoded[secret.Key] = secret.Value
	}

	return decoded
}
