package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	router := gin.Default()
	otelHandler := otelhttp.NewHandler(router, "server")

	// Endpoint to return "Hello, World!" response
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	// Enable OpenTelemetry tracing
	otel.SetTracerProvider(trace.NewTracerProvider())

	// Start server with OpenTelemetry handler
	if err := http.ListenAndServe(":8080", otelHandler); err != nil {
		panic(err)
	}
}
