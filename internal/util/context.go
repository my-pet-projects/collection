package util

import (
	"context"
	"net/http"
	"net/url"
)

type (
	RequestKey struct{}
)

// URL is a view helper that returns the current URL.
// The request path can be accessed with:
//
//	view.URL(ctx).Path // => ex. /login
func URL(ctx context.Context) *url.URL {
	return getContextValue(ctx, RequestKey{}, &http.Request{}).URL
}

func IsSameURL(ctx context.Context, path string) bool {
	return URL(ctx).Path == path
}

// Request is a view helper that returns the current http request.
// The request can be accessed with:
//
//	view.Request(ctx)
func Request(ctx context.Context) *http.Request {
	return getContextValue(ctx, RequestKey{}, &http.Request{})
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
