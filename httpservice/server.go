package httpservice

import (
	"net/http"

	"code.google.com/p/go.net/context"

	"github.com/julienschmidt/httprouter"
)

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

type key int

const paramsKey key = 0

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
	Do(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

type FuncService func(ctx context.Context, w http.ResponseWriter, r *http.Request)

func (f FuncService) Do(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

type Endpoint struct {
	router *httprouter.Router
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		router: httprouter.New(),
	}
}

func (e *Endpoint) GET(path string, s HttpService) {
	e.router.GET(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := WithParams(context.Background(), Params{p})
		s.Do(ctx, w, r)
	})
}

func (e *Endpoint) POST(path string, s HttpService) {
	e.router.POST(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := WithParams(context.Background(), Params{p})
		s.Do(ctx, w, r)
	})
}

func (e *Endpoint) Do(_ context.Context, w http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(w, r)
}

func Serve(addr string, s HttpService) error {
	return http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Do(context.Background(), w, r)
	}))
}
