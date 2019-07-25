package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type secret map[string]interface{}

type decodedSecret struct {
	Key   string
	Value string
}

func main() {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (info.Mode()&os.ModeCharDevice) != 0 || info.Size() < 0 {
		fmt.Fprintln(os.Stderr, "the command is intended to work with pipes.")
		fmt.Fprintln(os.Stderr, "usage: kubectl get secret <secret-name> -o <yaml|json> |", os.Args[0])
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "< secret.<yaml|json>")
		os.Exit(1)
	}

	stdin := read(os.Stdin)
	output, err := parse(stdin)

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not decode secret: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, string(output))
}

func parse(in []byte) ([]byte, error) {
	isJSON := isJSONString(in)

	var s secret
	if err := unmarshal(in, &s, isJSON); err != nil {
		return nil, err
	}
	if isJSON {
		datajson, ok := s["data"].(map[string]interface{})

		if !ok || len(datajson) == 0 {
			return in, nil
		}

		s["data"] = decodeJSON(datajson)
		return marshal(s["data"], isJSON)
	}

	// input is in yaml format
	datayaml, ok := s["data"].(map[interface{}]interface{})

	if !ok || len(datayaml) == 0 {
		return in, nil
	}

	s["data"] = decodeYaml(datayaml)
	return marshal(s["data"], isJSON)
}

func read(rd io.Reader) []byte {
	var output []byte
	reader := bufio.NewReader(rd)
	for {
		input, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}
	return output
}

func unmarshal(in []byte, out interface{}, asJSON bool) error {
	if asJSON {
		return json.Unmarshal(in, out)
	}
	return yaml.Unmarshal(in, out)
}

func marshal(d interface{}, asJSON bool) ([]byte, error) {
	if asJSON {
		return json.MarshalIndent(d, "", "    ")
	}
	return yaml.Marshal(d)
}

func decodeSecret(key, secret string, secrets chan decodedSecret) {
	var value string
	// avoid wrong encoded secrets
	if decoded, err := base64.StdEncoding.DecodeString(secret); err == nil {
		value = string(decoded)
	} else {
		value = secret
	}
	secrets <- decodedSecret{Key: key, Value: value}
}

func decodeYaml(data map[interface{}]interface{}) map[string]interface{} {
	length := len(data)
	secrets := make(chan decodedSecret, length)
	decoded := make(map[string]interface{}, length)
	for key, encoded := range data {
		go decodeSecret(key.(string), encoded.(string), secrets)
	}

	for i := 0; i < length; i++ {
		secret := <-secrets
		decoded[secret.Key] = secret.Value
	}

	return decoded
}

func decodeJSON(data map[string]interface{}) map[string]interface{} {
	length := len(data)
	secrets := make(chan decodedSecret, length)
	decoded := make(map[string]interface{}, length)
	for key, encoded := range data {
		go decodeSecret(key, encoded.(string), secrets)
	}

	for i := 0; i < length; i++ {
		secret := <-secrets
		decoded[secret.Key] = secret.Value
	}

	return decoded
}

func isJSONString(s []byte) bool {
	return json.Unmarshal(s, &json.RawMessage{}) == nil
}
