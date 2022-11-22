package main

import (
	"connor/go/web"
	"connor/go/web/middleware/accesslog"
	"net/http"
)

func main() {
	middlewares := []web.Middleware{
		accesslog.NewMiddleware(),
	}

	server := web.NewHttpServer(web.MiddlewaresOption(middlewares))

	server.AddRoute(http.MethodGet, "/test/:id", func(ctx *web.Context) {
		ctx.RespStatusCode = 200
		ctx.RespData = []byte(ctx.PathParams["id"])
	})

	server.Start(":8080")
}
