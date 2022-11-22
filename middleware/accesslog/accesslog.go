package accesslog

import (
	"connor/go/web"
	"encoding/json"
	"fmt"
)

type Option func(m *Middleware)

type Middleware struct {
	logFunc func(accessLog string)
}

func LogFunc(f func(log string)) Option {
	return func(m *Middleware) {
		m.logFunc = f
	}
}

func NewMiddleware(options ...Option) web.Middleware {
	m := &Middleware{}
	for _, option := range options {
		option(m)
	}
	if m.logFunc == nil {
		m.logFunc = func(log string) {
			fmt.Println(log)
		}
	}
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				l := &accessLog{
					Host:       ctx.Req.Host,
					Path:       ctx.Req.URL.Path,
					HttpMethod: ctx.Req.Method,
					Route:      ctx.MatchedRoute,
				}
				val, _ := json.Marshal(l)
				m.logFunc(string(val))
			}()
			next(ctx)
		}
	}
}

type accessLog struct {
	Host       string `json:"host"`
	Route      string `json:"route"`
	HttpMethod string `json:"http_method"`
	Path       string `json:"path"`
}
