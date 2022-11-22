package opentelemetry

import (
	"connor/go/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

type Option func(m *Middleware)

type Middleware struct {
	tracerProvider *sdktrace.TracerProvider
	spanExporter   sdktrace.SpanExporter
}

func WithSpanExporter(exporter sdktrace.SpanExporter) Option {
	return func(m *Middleware) {
		m.spanExporter = exporter
	}
}

func NewMiddleware(options ...Option) web.Middleware {
	m := &Middleware{}
	for _, option := range options {
		option(m)
	}
	if m.tracerProvider == nil {
		m.setTracerProvider()
	}
	Tracer = m.tracerProvider.Tracer("github.com/iconnor-code/go-web/middleware/opentelemetry")
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			reqCtx := ctx.Req.Context()
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))
			reqCtx, span := Tracer.Start(reqCtx, "go-web")

			span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
			span.SetAttributes(attribute.String("peer.hostname", ctx.Req.Host))
			span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(attribute.String("http.scheme", ctx.Req.URL.Scheme))
			span.SetAttributes(attribute.String("span.kind", "server"))
			span.SetAttributes(attribute.String("component", "web"))
			span.SetAttributes(attribute.String("peer.address", ctx.Req.RemoteAddr))
			span.SetAttributes(attribute.String("http.proto", ctx.Req.Proto))

			defer span.End()

			ctx.Req = ctx.Req.WithContext(reqCtx)
			next(ctx)

			// 使用命中的路由来作为 span 的名字
			if ctx.MatchedRoute != "" {
				span.SetName(ctx.MatchedRoute)
			}

			// 怎么拿到响应的状态呢？比如说用户有没有返回错误，响应码是多少，怎么办？
			span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))
		}
	}
}

func (m *Middleware) setTracerProvider() {
	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(m.spanExporter),
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("opentelemetry-demo"),
			attribute.String("environment", "dev"),
			attribute.Int64("ID", 1),
		)),
	)

	otel.SetTracerProvider(tp)
	m.tracerProvider = tp
}
