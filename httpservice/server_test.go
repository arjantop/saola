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
