package httpservice

import (
	"net/http"

	"golang.org/x/net/context"
)

type CancellableRoundTripper interface {
	http.RoundTripper
	CancelRequest(*http.Request)
}

type Client struct {
	Transport CancellableRoundTripper
}

type result struct {
	Response *http.Response
	Error    error
}

func (c *Client) Do(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	client := http.Client{Transport: c.Transport}
	r := make(chan result, 1)
	go func() {
		resp, err := client.Do(req)
		r <- result{resp, err}
	}()
	select {
	case <-ctx.Done():
		c.Transport.CancelRequest(req)
		<-r
		return nil, ctx.Err()
	case result := <-r:
		return result.Response, result.Error
	}
}
