package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Request allows for building up a request to a server in a chained fashion.
// Any errors are stored until the end of your call, so you only have to
// check once.
type Request struct {
	c *RESTClient

	timeout time.Duration

	// generic components accessible via method setters
	verb       string
	pathPrefix string
	subpath    string
	params     url.Values
	headers    http.Header

	// structural elements of the request that are part of the Kubernetes API conventions
	namespace    string
	namespaceSet bool

	// output
	err  error
	body io.Reader
}

type Result struct {
	body       []byte
	err        error
	statusCode int
}

// NewRequest TODO:NewRequest creates a new request
func NewRequest(c *RESTClient) *Request {
	return nil
}

// Do formats and executes the request.
func (r *Request) Do(ctx context.Context) Result {
	var result Result
	err := r.request(ctx, func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(resp, req)
	})
	if err != nil {
		return Result{err: err}
	}

	return result
}

// transformResponse convert body type and handle status code
func (r *Request) transformResponse(resp *http.Response, req *http.Request) Result {
	var raw []byte
	if resp.Body != nil {
		data, _ := ioutil.ReadAll(resp.Body)
		raw = data
	}
	return Result{
		body:       raw,
		statusCode: resp.StatusCode,
	}
}

func (r *Request) request(ctx context.Context, fn func(*http.Request, *http.Response)) error {
	req, err := r.newHTTPRequest(ctx)
	if err != nil {
		return err
	}

	client := r.c.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)

	f := func(req *http.Request, resp *http.Response) {
		if resp == nil {
			return
		}
		fn(req, resp)
	}
	f(req, resp)
	return nil
}

func (r *Request) newHTTPRequest(ctx context.Context) (*http.Request, error) {
	url := r.URL().String()
	req, err := http.NewRequest(r.verb, url, r.body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header = r.headers
	return req, nil
}

// URL TODO: make url according to request
func (r *Request) URL() *url.URL {

	return nil
}

// Namespace set namespace of request
func (r *Request) Namespace(namespace string) *Request {
	r.namespaceSet = true
	r.namespace = namespace
	return r
}

// Body pass go struct to request body
func (r *Request) Body(obj interface{}) *Request {
	if r.err != nil {
		return r
	}
	result, err := json.Marshal(obj)
	json := string(result)
	data, err := ioutil.ReadFile(json)
	if err != nil {
		r.err = err
		return r
	}
	r.body = bytes.NewReader(data)

	return r
}

/***********************Result************************/

func (r *Result) GetBody() []byte {
	return r.body
}
