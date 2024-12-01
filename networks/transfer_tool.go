package networks

import (
	"bytes"
	"encoding/json"
)

// Encode encodes any data type into a byte slice for network transmission
func Encode(data interface{}) ([]byte, error) {
	marshal, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

// Decode decodes a byte slice back into the original data type
func Decode(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(v)
	if err != nil {
		return err
	}
	return nil
}
