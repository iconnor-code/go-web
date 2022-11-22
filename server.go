package web

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

var _ http.Handler = &HttpServer{}

type ServerOption func(server *HttpServer)

type HttpServer struct {
	router      *router
	middlewares []Middleware
}

func NewHttpServer(opts ...ServerOption) *HttpServer {
	server := &HttpServer{
		router: newRouter(),
	}
	if opts != nil {
		for _, opt := range opts {
			opt(server)
		}
	}
	return server
}

func MiddlewaresOption(mids []Middleware) ServerOption {
	return func(server *HttpServer) {
		server.middlewares = mids
	}
}

func (hs *HttpServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: response,
	}

	middlewareChain := hs.serve

	for i := len(hs.middlewares) - 1; i >= 0; i-- {
		middlewareChain = hs.middlewares[i](middlewareChain)
	}

	final := func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			hs.flashResp(ctx)
		}
	}

	middlewareChain = final(middlewareChain)

	middlewareChain(ctx)
}

func (hs *HttpServer) Start(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen Port Error: %v", err)
	}

	// 在这里可以添加前置逻辑

	fmt.Printf("Start Http Server At %s ...\n", addr)
	if err := http.Serve(l, hs); err != nil {
		log.Fatalf("Start Http Srver Error: %v", err)
	}
}

func (hs *HttpServer) AddRoute(method string, path string, handler HandleFunc) {
	hs.router.addRoute(method, path, handler)
}

func (hs *HttpServer) serve(ctx *Context) {
	mi, ok := hs.router.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || mi == nil || mi.node.handler == nil {
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("NOT FOUND")
		return
	}
	ctx.PathParams = mi.pathParams
	ctx.MatchedRoute = mi.node.route
	mi.node.handler(ctx)
}

func (hs *HttpServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		fmt.Printf("写入响应失败 %v", err)
	}
}
