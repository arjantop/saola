package saola

import "golang.org/x/net/context"

type Filter interface {
	Do(ctx context.Context, s Service) error
}

type FuncFilter func(ctx context.Context, s Service) error

func (f FuncFilter) Do(ctx context.Context, s Service) error {
	return f(ctx, s)
}

func Chain(f Filter, fs ...Filter) Filter {
	if len(fs) == 0 {
		return f
	} else {
		return FuncFilter(func(ctx context.Context, s Service) error {
			return f.Do(ctx, Apply(s, Chain(fs[0], fs[1:]...)))
		})
	}
}

type Service interface {
	Do(ctx context.Context) error
}

type FuncService func(ctx context.Context) error

func (f FuncService) Do(ctx context.Context) error {
	return f(ctx)
}

func Apply(s Service, fs ...Filter) Service {
	if len(fs) == 0 {
		return s
	} else {
		return FuncService(func(ctx context.Context) error {
			return fs[0].Do(ctx, Apply(s, fs[1:]...))
		})
	}
}
