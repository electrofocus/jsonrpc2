package jsonrpc2

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler func(ctx context.Context, id ID, method string, params []byte) (result []byte, err error)

type Router struct {
	handlers map[string]Handler
}

func (r *Router) Handle(method string, handler Handler) {
	if r.handlers == nil {
		r.handlers = make(map[string]Handler)
	}

	r.handlers[method] = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, httpRequest *http.Request) {
	var req request

	decoder := json.NewDecoder(httpRequest.Body)
	decoder.UseNumber()

	if err := decoder.Decode(&req); err != nil {
		var e requestError

		if errors.As(err, &e) {
			_ = json.NewEncoder(w).Encode(errResponse{
				JSONRPC: "2.0",
				Error:   ErrInvalidRequest,
				ID:      req.ID,
			})

			return
		}

		_ = json.NewEncoder(w).Encode(errResponse{
			JSONRPC: "2.0",
			Error:   ErrParse,
			ID:      Null,
		})
	}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      ID              `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func (r *request) UnmarshalJSON(data []byte) error {
	var payload struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      ID              `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.JSONRPC != "2.0" {
		return requestError("'jsonrpc' MUST be exactly '2.0'")
	}

	if strings.HasPrefix(payload.Method, "rpc.") {
		return requestError("'method' MUST NOT begin with 'rpc.'")
	}

	return nil
}

type requestError string

func (e requestError) Error() string {

}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	ID      ID              `json:"id"`
}

type errResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Error   Error  `json:"error"`
	ID      ID     `json:"id"`
}
