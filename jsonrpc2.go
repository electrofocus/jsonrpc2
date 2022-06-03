package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
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

func (r *Router) handle(ctx context.Context, payload []byte) []byte {
	dec := json.NewDecoder(bytes.NewReader(payload))

	tok, err := dec.Token()
	if err != nil {
		return errParse
	}

	if delim, ok := tok.(json.Delim); ok && delim == json.Delim('[') {
		var b batch

		if err := dec.Decode(&b); err != nil {
			return errParse
		}

		return r.batch(ctx, b)
	}

	return r.single(ctx, payload)
}

func (r *Router) single(ctx context.Context, payload []byte) []byte {
	var req request

	if err := json.Unmarshal(payload, &req); err != nil {
		var e requestError

		if errors.As(err, &e) {
			return errInvalidRequest(req.ID)
		}

		return errParse
	}

	handler, ok := r.handlers[req.Method]
	if !ok {
		return errMethodNotFound(req.ID)
	}

	result, err := handler(ctx, req.ID, req.Method, req.Params)
	if err != nil {
		var e Error

		if errors.As(err, &e) {
			return encodeErr(req.ID, e)
		}

		return errInternal(req.ID)
	}

	return encodeResult(req.ID, result)
}

func (r *Router) batch(ctx context.Context, b batch) []byte {
	resp := make([][]byte, 0, len(b))

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	wg.Add(len(b))

	for _, payload := range b {
		go func(p []byte, mu *sync.Mutex, wg *sync.WaitGroup) {
			mu.Lock()
			resp = append(resp, r.single(ctx, p))
			mu.Unlock()
			wg.Done()
		}(payload, &mu, &wg)
	}

	wg.Wait()

	return encodeBatch(resp)
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

func encodeErr(id ID, err Error) []byte {
	result, _ := json.Marshal(errResponse{
		JSONRPC: "2.0",
		Error:   err,
		ID:      id,
	})

	return result
}

func encodeResult(id ID, result []byte) []byte {
	raw, _ := json.Marshal(response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	})

	return raw
}

func encodeBatch(b [][]byte) []byte {
	result, _ := json.Marshal(b)
	return result
}

type batch []json.RawMessage

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      ID              `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func (i *request) UnmarshalJSON(data []byte) error {
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

	*i = request{
		JSONRPC: payload.JSONRPC,
		ID:      payload.ID,
		Method:  payload.Method,
		Params:  payload.Params,
	}

	return nil
}

type requestError string

func (e requestError) Error() string {
	return string(e)
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
