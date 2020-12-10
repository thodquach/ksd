package filehandler

import (
	"bufio"
	"fmt"
	"io"

	"ksd/marshaler"
	"ksd/secret"
)

type FileHandler struct {
	Marshaler marshaler.MarshalerInterface
}

func NewFileHandler() *FileHandler {
	return &FileHandler{
		Marshaler: marshaler.NewMarshaler(),
	}
}

func (this *FileHandler) Parse(input []byte) ([]byte, error) {
	var secretOutput secret.Secret
	isJson := this.Marshaler.IsJSON(input)

	if err := this.Marshaler.Unmarshal(input, &secretOutput, isJson); err != nil {
		return nil, err
	}

	// JSON case
	if isJson {
		dataJson, ok := secretOutput["data"].(map[string]interface{})
		if !ok || len(dataJson) == 0 {
			return input, fmt.Errorf("Secret data field is empty.")
		}
		secretOutput["data"] = this.Marshaler.DecodeJSON(dataJson)
		return this.Marshaler.Marshal(secretOutput["data"], isJson)
	}

	// Yaml case
	dataYaml, ok := secretOutput["data"].(map[interface{}]interface{})
	if !ok || len(dataYaml) == 0 {
		return input, fmt.Errorf("Secret data field is empty.")
	}

	secretOutput["data"] = this.Marshaler.DecodeYaml(dataYaml)
	return this.Marshaler.Marshal(secretOutput["data"], isJson)
}

func (this *FileHandler) Read(rd io.Reader) []byte {
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
