package http

import (
	"errors"
	"fmt"

	"github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/remote"
	"go.bytecodealliance.org/cm"
)

// CallBatch forwards a batch of RemoteRequest to the host via
// tangent:remote/remote.call-batch and returns the mapped responses.
//
// This is the only place in the SDK that touches the generated remote bindings
// and cm.List types.
func CallBatch(reqs []Request) ([]Response, error) {
	// Map SDK-level RemoteRequest -> generated remote.Request
	internal := make([]remote.Request, len(reqs))

	for i, r := range reqs {
		// Method mapping
		var m remote.Method
		switch r.Method {
		case MethodGet:
			m = remote.MethodGet
		case MethodPost:
			m = remote.MethodPost
		case MethodPut:
			m = remote.MethodPut
		case MethodDelete:
			m = remote.MethodDelete
		case MethodPatch:
			m = remote.MethodPatch
		default:
			return nil, fmt.Errorf("invalid RemoteMethod: %d", r.Method)
		}

		hdrs := make([][2]string, len(r.Headers))
		for j, h := range r.Headers {
			hdrs[j] = [2]string{h.Name, h.Value}
		}

		var timeoutOpt cm.Option[uint32]
		if r.TimeoutMs != nil {
			timeoutOpt = cm.Some(*r.TimeoutMs)
		} else {
			timeoutOpt = cm.None[uint32]()
		}

		var cacheTtlOpt cm.Option[uint32]
		if r.CacheTtlMs != nil {
			cacheTtlOpt = cm.Some(*r.CacheTtlMs)
		} else {
			cacheTtlOpt = cm.None[uint32]()
		}

		internal[i] = remote.Request{
			ID:         r.ID,
			Method:     m,
			URL:        r.URL,
			Headers:    cm.ToList(hdrs),
			Body:       cm.ToList(r.Body),
			TimeoutMs:  timeoutOpt,
			CacheTTLMs: cacheTtlOpt,
		}
	}

	// Call host via generated binding. The exact signature depends on the
	// generator, but it will be something like:
	//
	//   func CallBatch(reqs cm.List[Request]) (cm.Result[cm.List[Response], string])
	//
	result := remote.CallBatch(cm.ToList(internal))

	if result.IsErr() {
		// Top-level error (e.g. host rejected the batch entirely)
		return nil, errors.New(*result.Err())
	}

	respList := result.OK().Slice()
	out := make([]Response, len(respList))

	for i, r := range respList {
		// Map headers back
		hdrs := r.Headers.Slice()
		outHeaders := make([]Header, len(hdrs))
		for j, h := range hdrs {
			outHeaders[j] = Header{
				Name:  h[0],
				Value: h[1],
			}
		}

		// Body is a cm.List[uint8]; convert to []byte
		bodyBytes := append([]byte(nil), r.Body.Slice()...)

		var errPtr *string
		if r.Error.Some() != nil {
			e := r.Error.Some()
			errPtr = e
		}

		out[i] = Response{
			ID:      r.ID,
			Status:  r.Status,
			Headers: outHeaders,
			Body:    bodyBytes,
			Error:   errPtr,
		}
	}

	return out, nil
}

// RemoteCallBatch forwards a batch of RemoteRequest to the host via
// tangent:remote/remote.call-batch and returns the mapped responses.
//
// This is the only place in the SDK that touches the generated remote bindings
// and cm.List types.
func Call(req Request) (Response, error) {
	// Method mapping
	var m remote.Method
	switch req.Method {
	case MethodGet:
		m = remote.MethodGet
	case MethodPost:
		m = remote.MethodPost
	case MethodPut:
		m = remote.MethodPut
	case MethodDelete:
		m = remote.MethodDelete
	case MethodPatch:
		m = remote.MethodPatch
	default:
		return Response{}, fmt.Errorf("invalid RemoteMethod: %d", req.Method)
	}

	hdrs := make([][2]string, len(req.Headers))
	for j, h := range req.Headers {
		hdrs[j] = [2]string{h.Name, h.Value}
	}

	var timeoutOpt cm.Option[uint32]
	if req.TimeoutMs != nil {
		timeoutOpt = cm.Some(*req.TimeoutMs)
	} else {
		timeoutOpt = cm.None[uint32]()
	}

	var cacheTtlOpt cm.Option[uint32]
	if req.CacheTtlMs != nil {
		cacheTtlOpt = cm.Some(*req.CacheTtlMs)
	} else {
		cacheTtlOpt = cm.None[uint32]()
	}

	result := remote.CallBatch(cm.ToList([]remote.Request{
		{
			ID:         req.ID,
			Method:     m,
			URL:        req.URL,
			Headers:    cm.ToList(hdrs),
			Body:       cm.ToList(req.Body),
			TimeoutMs:  timeoutOpt,
			CacheTTLMs: cacheTtlOpt,
		},
	}))

	if result.IsErr() {
		// Top-level error (e.g. host rejected the batch entirely)
		return Response{}, errors.New(*result.Err())
	}

	resp := result.OK().Slice()[0]

	outHeaders := make([]Header, len(hdrs))
	for j, h := range hdrs {
		outHeaders[j] = Header{
			Name:  h[0],
			Value: h[1],
		}
	}

	// Body is a cm.List[uint8]; convert to []byte
	bodyBytes := append([]byte(nil), resp.Body.Slice()...)

	var errPtr *string
	if resp.Error.Some() != nil {
		e := resp.Error.Some()
		errPtr = e
	}

	return Response{
		ID:      resp.ID,
		Status:  resp.Status,
		Headers: outHeaders,
		Body:    bodyBytes,
		Error:   errPtr,
	}, nil
}
