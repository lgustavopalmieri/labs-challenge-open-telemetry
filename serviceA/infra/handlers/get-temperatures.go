package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func GetTemperature(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)
	tracer := otel.Tracer("microservice-tracer")

	ctx, span := tracer.Start(ctx, "get-temperature")
	defer span.End()

	resp, err := http.Get("http://service2:8181")
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao fazer a chamada HTTP: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	ctx, renderContent := tracer.Start(ctx, "render-content")
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao escrever a resposta HTTP: %v", err), http.StatusInternalServerError)
		return
	}
	time.Sleep(time.Millisecond * 100)
	w.WriteHeader(http.StatusOK)
	renderContent.End()
}
