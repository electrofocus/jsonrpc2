package jsonrpc2

import (
	"encoding/json"
	"errors"
)

var Null = []byte("null")

type ID []byte

// UnmarshalJSON is not called, if member 'id' is not provided in corresponding JSON value.
func (id *ID) UnmarshalJSON(data []byte) error {
	var value any

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch value.(type) {
	case float64:
		break
	case string:
		break
	case nil:
		break
	default:
		return errors.New("invalid ID")
	}

	*id = append([]byte(nil), data...)
	return nil
}

func (id ID) MarshalJSON() ([]byte, error) {
	if id == nil {
		return Null, nil
	}

	return id, nil
}

func (id ID) IsSet() bool {
	return id != nil
}

func (id ID) IsNull() bool {
	return len(id) == 4 && id[0] == 'n' && id[1] == 'u' && id[2] == 'l' && id[3] == 'l'
}

func (id ID) Int64() (int64, bool) {
	num, ok := id.number()
	if !ok {
		return 0, false
	}

	i, err := num.Int64()
	if err != nil {
		return 0, false
	}

	return i, true
}

func (id ID) Float64() (float64, bool) {
	num, ok := id.number()
	if !ok {
		return 0, false
	}

	f, err := num.Float64()
	if err != nil {
		return 0, false
	}

	return f, true
}

func (id ID) String() (string, bool) {
	if id.IsNull() {
		return "", false
	}

	if id == nil {
		return "", false
	}

	var str string

	if err := json.Unmarshal(id, &str); err == nil {
		return str, true
	}

	num, ok := id.number()
	if !ok {
		return "", false
	}

	return num.String(), true
}

func (id ID) number() (json.Number, bool) {
	var num json.Number

	if err := json.Unmarshal(id, &num); err != nil {
		return num, false
	}

	return num, true
}
