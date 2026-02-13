package utils

import "net/http"

type Middleware func(http.Handler) http.Handler

func ApplyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

// CreateStack chains them in order: middleware1 -> middleware2 -> finalHandler
func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}
		return next
	}
}
