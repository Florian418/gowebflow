package httpd

import (
	"errors"
	"net/http"
)

// HandlerFunc is the handler signature used by goWebFlow.
// Unlike the standard http.HandlerFunc, it returns an error
// that is automatically caught and handled by the framework.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// HTTPError signals a specific HTTP status code to the framework.
// Use ErrHTTP to create one from a route handler.
type HTTPError struct {
	Code int
	Err  error
}

func (e *HTTPError) Error() string { return e.Err.Error() }

// ErrHTTP returns an error that triggers the OnError handler registered for code.
//
//	return httpd.ErrHTTP(http.StatusForbidden, fmt.Errorf("forbidden"))
func ErrHTTP(code int, err error) error {
	return &HTTPError{Code: code, Err: err}
}

// Get registers a handler for GET requests on the given path.
func (a *App) Get(path string, handler HandlerFunc) {
	if path == "/" {
		path = "/{$}"
	}
	a.mux.HandleFunc("GET "+path, a.wrap(handler))
}

// Post registers a handler for POST requests on the given path.
func (a *App) Post(path string, handler HandlerFunc) {
	if path == "/" {
		path = "/{$}"
	}
	a.mux.HandleFunc("POST "+path, a.wrap(handler))
}

// Put registers a handler for PUT requests on the given path.
func (a *App) Put(path string, handler HandlerFunc) {
	if path == "/" {
		path = "/{$}"
	}
	a.mux.HandleFunc("PUT "+path, a.wrap(handler))
}

// Delete registers a handler for DELETE requests on the given path.
func (a *App) Delete(path string, handler HandlerFunc) {
	if path == "/" {
		path = "/{$}"
	}
	a.mux.HandleFunc("DELETE "+path, a.wrap(handler))
}

// wrap converts a goWebFlow HandlerFunc into a standard http.HandlerFunc.
// If the handler returns an HTTPError, the matching OnError handler is called.
// Plain errors fall back to the 500 handler (or http.Error if none is registered).
func (a *App) wrap(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}
		code := http.StatusInternalServerError
		var httpErr *HTTPError
		if errors.As(err, &httpErr) {
			code = httpErr.Code
		}
		if handler, ok := a.errorHandlers[code]; ok {
			if err2 := handler(w, r); err2 != nil {
				http.Error(w, err2.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, err.Error(), code)
	}
}
