package chain

import (
	"bytes"
	"encoding/gob"
)

// 编码函数：将 [][]byte 编码为 []byte
func EncodeGob(data [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(data) // 将 [][]byte 编码为 []byte
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 解码函数：将 []byte 解码回 [][]byte
func DecodeGob(data []byte) ([][]byte, error) {
	var result [][]byte
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&result) // 解码回 [][]byte
	if err != nil {
		return nil, err
	}
	return result, nil
}
