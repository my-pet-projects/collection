package util

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/my-pet-projects/collection/internal/model"
)

type key string

const (
	userKey    = key("user")
	requestKey = key("request")
)

// URL is a view helper that returns the current URL.
// The request path can be accessed with:
//
//	view.URL(ctx).Path // => ex. /login
func URL(ctx context.Context) *url.URL {
	return getContextValue(ctx, requestKey, &http.Request{}).URL
}

func IsSameURL(ctx context.Context, path string) bool {
	return URL(ctx).Path == path
}

func UrlStartsWith(ctx context.Context, path string) bool {
	urlPath := URL(ctx).Path
	return strings.HasPrefix(urlPath, path)
}

// Request is a view helper that returns the current http request.
// The request can be accessed with:
//
//	view.Request(ctx)
func Request(ctx context.Context) *http.Request {
	return getContextValue(ctx, requestKey, &http.Request{})
}

// getContextValue is a helper function to retrieve a value from the context.
// It returns the value if present, otherwise returns the provided default value.
func getContextValue[T any](ctx context.Context, key interface{}, defaultValue T) T {
	value, ok := ctx.Value(key).(T)
	if !ok {
		return defaultValue
	}
	return value
}

func ContextWithRequest(ctx context.Context, value *http.Request) context.Context {
	return context.WithValue(ctx, requestKey, value)
}

func ContextWithUser(ctx context.Context, value model.User) context.Context {
	return context.WithValue(ctx, userKey, value)
}

// UserFromContext returns the user from the context.
func UserFromContext(ctx context.Context) model.User {
	return getContextValue(ctx, userKey, model.User{})
}
