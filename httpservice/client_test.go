package httpservice_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arjantop/saola/httpservice"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func NewServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/foo" {
			w.Write([]byte("bar"))
		} else if r.URL.Path == "/timeout" {
			time.Sleep(1 * time.Millisecond)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestClientDoRequestSuccess(t *testing.T) {
	ts := NewServer()
	defer ts.Close()
	c := httpservice.Client{
		Transport: &http.Transport{},
	}
	req, err := http.NewRequest("GET", ts.URL+"/foo", nil)
	assert.NoError(t, err)
	res, err := c.Do(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err)
	assert.Equal(t, "bar", string(content))
}

func TestClientDoRequestFailure(t *testing.T) {
	c := httpservice.Client{
		Transport: &http.Transport{},
	}
	req, err := http.NewRequest("GET", "http://localhost:12345", nil)
	assert.NoError(t, err)
	res, err := c.Do(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestClientDoContextTimeout(t *testing.T) {
	ts := NewServer()
	defer ts.Close()
	c := httpservice.Client{
		Transport: &http.Transport{},
	}
	req, err := http.NewRequest("GET", ts.URL+"/timeout", nil)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
	defer cancel()
	_, err = c.Do(ctx, req)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func BenchmarkClientDo(b *testing.B) {
	ts := NewServer()
	defer ts.Close()
	c := httpservice.Client{
		Transport: &http.Transport{},
	}
	req, _ := http.NewRequest("GET", ts.URL+"/foo", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, _ := c.Do(context.Background(), req)
		res.Body.Close()
	}
}

func BenchmarkClientDoWithTimeout(b *testing.B) {
	ts := NewServer()
	defer ts.Close()
	c := httpservice.Client{
		Transport: &http.Transport{},
	}
	req, _ := http.NewRequest("GET", ts.URL+"/foo", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		res, _ := c.Do(ctx, req)
		res.Body.Close()
		cancel()
	}
}

func BenchmarkClientStandard(b *testing.B) {
	ts := NewServer()
	defer ts.Close()
	c := http.Client{
		Transport: &http.Transport{},
	}
	req, _ := http.NewRequest("GET", ts.URL+"/foo", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, _ := c.Do(req)
		res.Body.Close()
	}
}

func BenchmarkClientStandardWithTimeout(b *testing.B) {
	ts := NewServer()
	defer ts.Close()
	c := http.Client{
		Transport: &http.Transport{},
		Timeout:   100 * time.Millisecond,
	}
	req, _ := http.NewRequest("GET", ts.URL+"/foo", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, _ := c.Do(req)
		res.Body.Close()
	}
}
