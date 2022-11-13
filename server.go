package web

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

var _ http.Handler = &HttpServer{}

type HttpServer struct {
	router *router
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		router: newRouter(),
	}
}

func (hs *HttpServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: response,
	}
	mi, ok := hs.router.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || mi == nil || mi.node.handler == nil {
		ctx.Resp.WriteHeader(404)
		ctx.Resp.Write([]byte("Not Found"))
		return
	}
	mi.node.handler(ctx)
}

func (hs *HttpServer) Start(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen Port Error: %v", err)
	}

	// 在这里可以添加前置逻辑

	fmt.Printf("Start Http Server At %s ...", addr)
	if err := http.Serve(l, hs); err != nil {
		log.Fatalf("Start Http Srver Error: %v", err)
	}
}

func (hs *HttpServer) AddRoute(method string, path string, handler handleFunc) {
	hs.router.addRoute(method, path, handler)
}
