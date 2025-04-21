package router

import (
	"net/http"
)

// Route represents a single route in the application
type Route struct {
	// Name is the name of the route
	Name string

	// Path is the URL path of the route
	Path string

	// Methods are the HTTP methods the route responds to
	Methods []string

	// Handler is the HTTP handler for the route
	Handler http.Handler

	// Middleware is a list of middleware to apply to the route
	Middleware []func(http.Handler) http.Handler
}

// NewRoute creates a new route
func NewRoute(name, path string, methods []string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return Route{
		Name:       name,
		Path:       path,
		Methods:    methods,
		Handler:    handler,
		Middleware: middleware,
	}
}

// GET creates a new GET route
func GET(name, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return NewRoute(name, path, []string{"GET"}, handler, middleware...)
}

// POST creates a new POST route
func POST(name, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return NewRoute(name, path, []string{"POST"}, handler, middleware...)
}

// PUT creates a new PUT route
func PUT(name, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return NewRoute(name, path, []string{"PUT"}, handler, middleware...)
}

// DELETE creates a new DELETE route
func DELETE(name, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return NewRoute(name, path, []string{"DELETE"}, handler, middleware...)
}

// PATCH creates a new PATCH route
func PATCH(name, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return NewRoute(name, path, []string{"PATCH"}, handler, middleware...)
}

// ANY creates a new route that responds to any HTTP method
func ANY(name, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return NewRoute(name, path, []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}, handler, middleware...)
}

// Methods that handle both GET and POST
func GetPost(name, path string, handler http.Handler, middleware ...func(http.Handler) http.Handler) Route {
	return NewRoute(name, path, []string{"GET", "POST"}, handler, middleware...)
}
