package transport

import "net/http"

// HeadersTransport is a http.RoundTripper that supports adding custom HTTP headers to requests.
type HeadersTransport struct {
	// Headers is a set of headers to add to each request.
	Headers map[string]string

	// BaseTransport is the underlying HTTP transport to use when making requests. It will default to http.DefaultTransport if nil.
	BaseTransport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *HeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)

	for k, v := range t.Headers {
		req.Header.Add(k, v)
	}

	return t.transport().RoundTrip(req)
}

// Client returns an *http.Client that makes requests that include additional headers.
func (t *HeadersTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

// transport returns the underlying HTTP transport. If none is set, http.DefaultTransport is used.
func (t *HeadersTransport) transport() http.RoundTripper {
	if t.BaseTransport != nil {
		return t.BaseTransport
	}

	return http.DefaultTransport
}

// cloneRequest returns a clone of the provided *http.Request. The clone is a shallow copy of the struct and its headers map.
func cloneRequest(r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.Header = make(http.Header, len(r.Header))

	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}

	return r2
}
