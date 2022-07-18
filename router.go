package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
)

type Handler func(ctx context.Context, id ID, method string, params []byte) (result []byte, err error)

type Router struct {
	handlers map[string]Handler
}

func NewRouter() Router {
	return Router{
		handlers: make(map[string]Handler),
	}
}

func (r *Router) Handle(method string, handler Handler) {
	if r.handlers == nil {
		r.handlers = make(map[string]Handler)
	}

	r.handlers[method] = handler
}

func (r *Router) Serve(ctx context.Context, msg []byte) []byte {
	dec := json.NewDecoder(bytes.NewReader(msg))

	tok, err := dec.Token()
	if err != nil {
		return errParse
	}

	if delim, ok := tok.(json.Delim); ok && delim == json.Delim('{') {
		return r.serve(ctx, msg)
	} else if ok && delim == json.Delim('[') {
		var b batch

		if err := dec.Decode(&b); err != nil {
			return errParse
		}

		return r.serveBatch(ctx, b)
	}

	return errParse
}

func (r *Router) serve(ctx context.Context, msg []byte) (res []byte) {
	var req request

	if err := json.Unmarshal(msg, &req); err != nil {
		var e requestError

		if errors.As(err, &e) {
			return errInvalidRequest(req.ID)
		}

		return errParse
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %s\n%s", err, string(debug.Stack()))

			res = encodeErr(req.ID, Error{
				Code:    0, // TODO: fill code
				Message: fmt.Sprint(err),
			})
		}
	}()

	handler, ok := r.handlers[req.Method]
	if !ok {
		return errMethodNotFound(req.ID)
	}

	res, err := handler(ctx, req.ID, req.Method, req.Params)
	if err != nil {
		var e Error

		if errors.As(err, &e) {
			return encodeErr(req.ID, e)
		}

		return errInternal(req.ID)
	}

	return encodeResult(req.ID, res)
}

func (r *Router) serveBatch(ctx context.Context, b batch) []byte {
	res := make(batch, 0, len(b))

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	wg.Add(len(b))

	for _, payload := range b {
		go func(p []byte, mu *sync.Mutex, wg *sync.WaitGroup) {
			mu.Lock()
			res = append(res, r.serve(ctx, p))
			mu.Unlock()
			wg.Done()
		}(payload, &mu, &wg)
	}

	wg.Wait()

	return encodeBatch(res)
}

func encodeResult(id ID, res []byte) []byte {
	raw, _ := json.Marshal(response{
		JSONRPC: "2.0",
		Result:  res,
		ID:      id,
	})

	return raw
}

func encodeBatch(b batch) []byte {
	result, _ := json.Marshal(b)
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

		*i = request{
			ID: payload.ID,
		}

		return requestError("'jsonrpc' MUST be exactly '2.0'")
	}

	if strings.HasPrefix(payload.Method, "rpc.") {

		*i = request{
			ID: payload.ID,
		}

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
	Result  json.RawMessage `json:"result,omitempty"`
	ID      ID              `json:"id"`
}

type errResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Error   Error  `json:"error"`
	ID      ID     `json:"id"`
}
