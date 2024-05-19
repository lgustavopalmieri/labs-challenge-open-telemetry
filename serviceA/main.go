package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/lgustavopalmieri/labs-challenge-open-telemetry/serviceA/infra/handlers"
	"github.com/lgustavopalmieri/labs-challenge-open-telemetry/serviceA/infra/opentel"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	otelProvider := &opentel.OpenTelemetryProvider{
		ServiceName:  "microservice-tracer",
		CollectorURL: "otel-collector:4317",
	}

	otelShutdown, err := otelProvider.InitProvider()
	if err != nil {
		fmt.Println("Erro ao inicializar o provedor OpenTelemetry:", err)
		return
	}

	defer otelShutdown(ctx)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/temperature/{cep}", handlers.GetTemperature)
	r.Get("/metrics", promhttpHandler())

	go func() {
		log.Println("Starting server on port", ":8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatal(err)
		}
	}()

	select {
	case <-sigCh:
		log.Println("Shutting down gracefully, CTRL+C pressed...")
	case <-ctx.Done():
		log.Println("Shutting down due to other reason...")
	}

	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
}

func promhttpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	}
}