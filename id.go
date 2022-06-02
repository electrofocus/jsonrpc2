package jsonrpc2

import (
	"bytes"
	"encoding/json"
)

type ID []byte

func (id *ID) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	if string(data) == string(Null) {
		*id = append([]byte(nil), data...)
		return nil
	}

	var number json.Number

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&number); err != nil {
		return requestError("'id' MUST contain a String, Number, or NULL if included")
	}

	*id = append([]byte(nil), data...)
	return nil
}

func (id *ID) MarshalJSON() ([]byte, error) {
	return nil, nil
}

var Null = []byte("null")
