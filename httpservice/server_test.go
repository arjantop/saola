package httpservice_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arjantop/saola/httpservice"
	"github.com/stretchr/testify/assert"

	"code.google.com/p/go.net/context"
)

func TestServerParamsInContext(t *testing.T) {
	params := httpservice.EmptyParams()
	params.Set("k", "v")

	retrieved := httpservice.GetParams(httpservice.WithParams(context.Background(), params))
	assert.Equal(t, "v", retrieved.Get("k"))
}

func TestServerParamsNotInContext(t *testing.T) {
	retrieved := httpservice.GetParams(context.Background())
	assert.Equal(t, "", retrieved.Get("k"))
}

func TestServerParamsMultipleSet(t *testing.T) {
	params := httpservice.EmptyParams()
	params.Set("k", "v")
	params.Set("k", "v2")

	assert.Equal(t, "v2", params.Get("k"))
}

func TestServerEndpointGET(t *testing.T) {
	endpoint := httpservice.NewEndpoint()
	endpoint.POST("/hello/:name", nil)
	endpoint.GET("/hello/:name", httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		params := httpservice.GetParams(ctx)
		fmt.Fprintf(w, params.Get("name"))
	}))

	req, err := http.NewRequest("GET", "http://example.com/hello/bob", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	endpoint.Do(context.Background(), w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "bob", w.Body.String())
}

func TestServerEndpointPOST(t *testing.T) {
	endpoint := httpservice.NewEndpoint()
	endpoint.GET("/hello/:name", nil)
	endpoint.POST("/hello/:name", httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		params := httpservice.GetParams(ctx)
		fmt.Fprintf(w, params.Get("name"))
	}))

	req, err := http.NewRequest("POST", "http://example.com/hello/lucian", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	endpoint.Do(context.Background(), w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "lucian", w.Body.String())
}

func TestServerEndpointPUT(t *testing.T) {
	endpoint := httpservice.NewEndpoint()
	endpoint.POST("/hello/:name", nil)
	endpoint.PUT("/hello/:name", httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		params := httpservice.GetParams(ctx)
		fmt.Fprintf(w, params.Get("name"))
	}))

	req, err := http.NewRequest("PUT", "http://example.com/hello/john", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	endpoint.Do(context.Background(), w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "john", w.Body.String())
}

func NewService() httpservice.HttpService {
	return httpservice.FuncService(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "service")
	})
}

func NewFilter(name string) httpservice.ServiceFilter {
	return httpservice.FuncFilter(func(ctx context.Context, w http.ResponseWriter, r *http.Request, s httpservice.HttpService) {
		fmt.Fprintf(w, "%s-", name)
		s.Do(ctx, w, r)
		fmt.Fprintf(w, "-%s", name)
	})
}

func assertFilter(t *testing.T, f func(ctx context.Context, w http.ResponseWriter, r *http.Request, s httpservice.HttpService), expected string) {
	service := NewService()

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), "foo", "bar")
	f(ctx, w, req, service)

	assert.Equal(t, expected, w.Body.String())
	value, ok := ctx.Value("foo").(string)
	assert.True(t, ok)
	assert.Equal(t, "bar", value)
}

func TestServerServiceFilterChainOne(t *testing.T) {
	assertFilter(t, func(ctx context.Context, w http.ResponseWriter, r *http.Request, s httpservice.HttpService) {
		filterA := NewFilter("A")
		httpservice.Chain(filterA).Do(ctx, w, r, s)
	}, "A-service-A")
}

func TestServerServiceFilterChainMultiple(t *testing.T) {
	assertFilter(t, func(ctx context.Context, w http.ResponseWriter, r *http.Request, s httpservice.HttpService) {
		filterA := NewFilter("A")
		filterB := NewFilter("B")
		filterC := NewFilter("C")
		httpservice.Chain(filterA, filterB, filterC).Do(ctx, w, r, s)
	}, "A-B-C-service-C-B-A")
}

func TestServerFilterApplyOne(t *testing.T) {
	assertFilter(t, func(ctx context.Context, w http.ResponseWriter, r *http.Request, s httpservice.HttpService) {
		filterA := NewFilter("A")
		httpservice.Apply(s, filterA).Do(ctx, w, r)
	}, "A-service-A")
}

func TestServerFilterApplyMultiple(t *testing.T) {
	assertFilter(t, func(ctx context.Context, w http.ResponseWriter, r *http.Request, s httpservice.HttpService) {
		filterA := NewFilter("A")
		filterB := NewFilter("B")
		filterC := NewFilter("C")
		httpservice.Apply(s, filterA, filterB, filterC).Do(ctx, w, r)
	}, "A-B-C-service-C-B-A")
}
