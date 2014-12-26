package httpservice

import (
	"net/http"

	"github.com/arjantop/saola"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

type key int

const httpRequest key = 0

type HttpRequest struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func WithHttpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequest, &HttpRequest{w, r})
}

func GetHttpRequest(ctx context.Context) *HttpRequest {
	r, ok := ctx.Value(httpRequest).(*HttpRequest)
	if !ok {
		panic("missing http request in context")
	}
	return r
}

type Params struct {
	params httprouter.Params
}

func EmptyParams() Params {
	return Params{make([]httprouter.Param, 0)}
}

func (p Params) Get(key string) string {
	return p.params.ByName(key)
}

func (p *Params) Set(key, value string) {
	for i, _ := range p.params {
		p := &p.params[i]
		if p.Key == key {
			p.Value = value
			return
		}
	}
	p.params = append(p.params, httprouter.Param{
		Key:   key,
		Value: value,
	})
}

const paramsKey key = 1

func WithParams(ctx context.Context, p Params) context.Context {
	return context.WithValue(ctx, paramsKey, p)
}

func GetParams(ctx context.Context) Params {
	if params, ok := ctx.Value(paramsKey).(Params); ok {
		return params
	}
	return EmptyParams()
}

type HttpService interface {
	DoHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	saola.Service
}

type FuncService func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

func (f FuncService) DoHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return f(ctx, w, r)
}

func (f FuncService) Do(ctx context.Context) error {
	r := GetHttpRequest(ctx)
	return f.DoHTTP(ctx, r.Writer, r.Request)
}

type Endpoint struct {
	router *httprouter.Router
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		router: httprouter.New(),
	}
}

func (e *Endpoint) GET(path string, s saola.Service) {
	e.router.GET(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx := WithParams(WithHttpRequest(cctx, w, r), Params{p})
		s.Do(ctx)
	})
}

func (e *Endpoint) POST(path string, s saola.Service) {
	e.router.POST(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx := WithParams(WithHttpRequest(cctx, w, r), Params{p})
		s.Do(ctx)
	})
}

func (e *Endpoint) PUT(path string, s saola.Service) {
	e.router.PUT(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx := WithParams(WithHttpRequest(cctx, w, r), Params{p})
		s.Do(ctx)
	})
}

func (e *Endpoint) DoHTTP(_ context.Context, w http.ResponseWriter, r *http.Request) error {
	e.router.ServeHTTP(w, r)
	return nil
}

func (e *Endpoint) Do(ctx context.Context) error {
	return Do(e, ctx)
}

func Serve(addr string, s saola.Service) error {
	return http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx := WithHttpRequest(cctx, NewResponseWriter(w), r)
		s.Do(ctx)
	}))
}

func Do(s HttpService, ctx context.Context) error {
	r := GetHttpRequest(ctx)
	return s.DoHTTP(ctx, r.Writer, r.Request)
}
