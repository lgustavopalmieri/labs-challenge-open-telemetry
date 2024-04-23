package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ResultViaCep struct {
	Cep        string `json:"cep"`
	Localidade string `json:"localidade"`
	Erro       bool   `json:"erro"`
}

type TemperatureData struct {
	City       string  `json:"city"`
	Celsius    float64 `json:"celsius"`
	Fahrenheit float64 `json:"fahrenheit"`
	Kelvin     float64 `json:"kelvin"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func validateCEP(cep string) bool {
	regex := regexp.MustCompile(`^\d{8}$|^\d{5}-\d{3}$`)
	return regex.MatchString(cep)
}

func GetTemperature(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)
	tracer := otel.Tracer("viacepapi")

	ctx, span := tracer.Start(ctx, "send-viacep")
	defer span.End()

	cep := chi.URLParam(r, "cep")
	if !validateCEP(cep) {
		HandleError(w, http.StatusNotFound, "invalid zipcode", nil)
		return
	}

	url1 := "http://viacep.com.br/ws/" + cep + "/json/"
	req, err := http.NewRequest(http.MethodGet, url1, nil)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	respViacep, err := http.DefaultClient.Do(req)

	if err != nil {
		HandleError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	defer respViacep.Body.Close()

	body, err := io.ReadAll(respViacep.Body)

	if err != nil {
		HandleError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var data ResultViaCep
	err = json.Unmarshal(body, &data)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if data.Erro {
		HandleError(w, http.StatusNotFound, "cannot find zipcode", nil)
		return
	}

	city := data.Localidade

	ctx, renderContent := tracer.Start(ctx, "get-from-microservice2")

	temperatura, err := http.Get("http://service2:8181/" + city)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	defer temperatura.Body.Close()

	var temperatureData TemperatureData
	err = json.NewDecoder(temperatura.Body).Decode(&temperatureData)
	if err != nil {
		HandleError(w, http.StatusInternalServerError, "Failed to decode temperature data", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(temperatureData)
	renderContent.End()
}

func HandleError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errorResponse := map[string]string{"error": message}
	json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		log.Println(err)
	}
}
