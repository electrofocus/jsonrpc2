package jsonrpc2

var (
	ErrParse          = Error{-32700, "Parse error"}
	ErrInvalidRequest = Error{-32600, "Invalid Request"}
	ErrMethodNotFound = Error{-32601, "Method not found"}
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}
