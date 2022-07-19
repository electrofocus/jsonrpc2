package rpc

const (
	Internal      = 1000
	BadRequest    = 1001
	Unprocessable = 1002
	NotFound      = 1003
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}
