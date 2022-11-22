//go:build manual

package opentelemetry

import (
	"connor/go/web"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestTracing(t *testing.T) {
	//spanExporter, file := newFile()
	//defer file.Close()

	exporter := newJeager()
	middlewares := []web.Middleware{
		NewMiddleware(WithSpanExporter(exporter)),
	}

	s := web.NewHttpServer(web.MiddlewaresOption(middlewares))

	s.AddRoute(http.MethodGet, "/", func(ctx *web.Context) {
		ctx.Resp.Write([]byte("hello, world"))
	})
	s.AddRoute(http.MethodGet, "/user", func(ctx *web.Context) {
		c, span := Tracer.Start(ctx.Req.Context(), "first_layer")
		defer span.End()

		c, second := Tracer.Start(c, "second_layer")
		time.Sleep(time.Second)
		c, third1 := Tracer.Start(c, "third_layer_1")
		time.Sleep(100 * time.Millisecond)
		third1.End()
		c, third2 := Tracer.Start(c, "third_layer_1")
		time.Sleep(300 * time.Millisecond)
		third2.End()
		second.End()
		ctx.RespStatusCode = 200
		ctx.RespData = []byte("hello, world")
	})

	s.Start(":8081")
}

func newFile() (sdktrace.SpanExporter, *os.File) {
	l := log.New(os.Stdout, "", 0)
	f, err := os.Create("traces.txt")
	if err != nil {
		l.Fatal(err)
	}

	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(f),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
	if err != nil {
		if err != nil {
			log.Fatal(err)
		}
	}
	return exp, f
}

func newZipkin() sdktrace.SpanExporter {
	exp, err := zipkin.New(
		"http://127.0.0.1:9411/api/v2/spans",
		zipkin.WithLogger(log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)),
	)
	if err != nil {
		log.Fatal(err)
	}
	//return sdktrace.NewBatchSpanProcessor(exp)
	return exp
}

func newJeager() sdktrace.SpanExporter {
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint("http://127.0.0.1:14268/api/traces"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	return exp
}
