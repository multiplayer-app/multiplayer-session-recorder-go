package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/multiplayer-app/multiplayer-otlp-go/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()



	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	
	// Configure middleware options
	options := middleware.NewMiddlewareOptions(
		middleware.WithCaptureHeaders(true),
		middleware.WithCaptureBody(true),
		middleware.WithMaskHeadersEnabled(true),
		middleware.WithMaskBodyEnabled(true),
		middleware.WithMaxPayloadSizeBytes(5000),
		middleware.WithMaskHeadersList([]string{"Authorization", "Cookie", "Set-Cookie", "Accept-Encoding"}),
		middleware.WithHeadersToExclude([]string{"User-Agent"}),
	)
	
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Apply our middleware for request/response data capture
		handler := middleware.WithRequestData(http.HandlerFunc(handlerFunc), options)
		handler = middleware.WithResponseData(handler, options)
		
		// Apply OpenTelemetry HTTP instrumentation
		handler = otelhttp.WithRouteTag(pattern, handler)
		mux.Handle(pattern, handler)
	}

	// Register handlers.
	handleFunc("/rolldice/", rolldice)
	handleFunc("/rolldice/{player}", rolldice)
	handleFunc("/health", healthCheck)
	
	// Wrap the entire mux with OpenTelemetry instrumentation
	handler := otelhttp.NewHandler(mux, "/")

	return handler
}



func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
