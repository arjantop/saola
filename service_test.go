package saola_test

import (
	"fmt"
	"testing"

	"github.com/arjantop/saola"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func WithString(ctx context.Context) context.Context {
	s := ""
	return context.WithValue(ctx, "value", &s)
}

func WriteString(ctx context.Context, s string) {
	state := ctx.Value("value").(*string)
	*state = *state + s
}

func GetString(ctx context.Context) string {
	state := ctx.Value("value").(*string)
	return *state
}

func NewService() saola.Service {
	return saola.FuncService(func(ctx context.Context) error {
		WriteString(ctx, "service")
		return nil
	})
}

func NewFilter(name string) saola.Filter {
	return saola.FuncFilter(func(ctx context.Context, s saola.Service) error {
		WriteString(ctx, fmt.Sprintf("%s-", name))
		err := s.Do(ctx)
		WriteString(ctx, fmt.Sprintf("-%s", name))
		return err
	})
}

func assertFilter(t *testing.T, f saola.FuncFilter, expected string) {
	service := NewService()

	ctx := WithString(context.Background())
	err := f(ctx, service)
	assert.NoError(t, err)

	assert.Equal(t, expected, GetString(ctx))
}

func TestFilterChainOne(t *testing.T) {
	assertFilter(t, func(ctx context.Context, s saola.Service) error {
		filterA := NewFilter("A")
		return saola.Chain(filterA).Do(ctx, s)
	}, "A-service-A")
}

func TestServerServiceFilterChainMultiple(t *testing.T) {
	assertFilter(t, func(ctx context.Context, s saola.Service) error {
		filterA := NewFilter("A")
		filterB := NewFilter("B")
		filterC := NewFilter("C")
		return saola.Chain(filterA, filterB, filterC).Do(ctx, s)
	}, "A-B-C-service-C-B-A")
}

func TestServerFilterApplyOne(t *testing.T) {
	assertFilter(t, func(ctx context.Context, s saola.Service) error {
		filterA := NewFilter("A")
		return saola.Apply(s, filterA).Do(ctx)
	}, "A-service-A")
}

func TestServerFilterApplyMultiple(t *testing.T) {
	assertFilter(t, func(ctx context.Context, s saola.Service) error {
		filterA := NewFilter("A")
		filterB := NewFilter("B")
		filterC := NewFilter("C")
		return saola.Apply(s, filterA, filterB, filterC).Do(ctx)
	}, "A-B-C-service-C-B-A")
}
