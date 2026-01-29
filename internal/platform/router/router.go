package router

import "net/http"

type Router interface {
	Handle(method, path string, handler http.HandlerFunc)
}
