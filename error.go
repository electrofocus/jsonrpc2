package jsonrpc2

import "encoding/json"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

var errParse = []byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}`)

func errInternal(id ID) []byte {
	result, _ := json.Marshal(errResponse{
		JSONRPC: "2.0",
		Error:   Error{1000, "Internal error"},
		ID:      id,
	})

	return result
}

func errMethodNotFound(id ID) []byte {
	result, _ := json.Marshal(errResponse{
		JSONRPC: "2.0",
		Error:   Error{-32601, "Method not found"},
		ID:      id,
	})

	return result
}

func errInvalidRequest(id ID) []byte {
	result, _ := json.Marshal(errResponse{
		JSONRPC: "2.0",
		Error:   Error{-32600, "Invalid Request"},
		ID:      id,
	})

	return result
}
