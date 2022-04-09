package rest

import (
	"net/http"
	"net/url"
)

// Interface captures the set of operations for generically interacting with Kubernetes REST apis.
type Interface interface {
	Post() *Request
	Put() *Request
	Get() *Request
	Delete() *Request
}

type RESTClient struct {
	// base is the root URL for all invocations of the client
	base *url.URL

	// Set specific behavior of the client.  If not set http.DefaultClient will be used.
	Client *http.Client
}

// NewRESTClient TODO:creates a new client with restful style
func NewRESTClient(baseURL *url.URL, client *http.Client) (*RESTClient, error) {
	base := *baseURL

	return &RESTClient{
		base:   &base,
		Client: client,
	}, nil
}

// NewVerb begins a request with a verb (GET, POST, PUT, DELETE)
func NewVerb(c *RESTClient, verb string) *Request {
	req := NewRequest(c)
	req.verb = verb
	return req
}

func (c *RESTClient) Post() *Request {
	return NewVerb(c, "POST")
}

func (c *RESTClient) Put() *Request {
	return NewVerb(c, "PUT")
}

func (c *RESTClient) Get() *Request {
	return NewVerb(c, "GET")
}

func (c *RESTClient) Delete() *Request {
	return NewVerb(c, "DELETE")
}
