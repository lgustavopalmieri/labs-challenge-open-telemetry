package handlers

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func CelsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

func CelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

func GetTemperature(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)
	tracer := otel.Tracer("wheatherapi")

	ctx, span := tracer.Start(ctx, "send-weatherapi")
	defer span.End()

	city := chi.URLParam(r, "city")
	sc := url.QueryEscape(city)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	url := "https://api.weatherapi.com/v1/current.json?q=" + sc + "&key=360ddfd38d0d4cd3b72102808240403"
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Weather service returned non-OK status", resp.StatusCode)
		return
	}

	ctx, calculateTemps := tracer.Start(ctx, "calculate-temperatures")

	var data WeatherResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Failed to decode weather data", http.StatusInternalServerError)
		return
	}

	result := &Temperature{
		Celsius:    data.Current.TemperatureC,
		Fahrenheit: CelsiusToFahrenheit(data.Current.TemperatureC),
		Kelvin:     CelsiusToKelvin(data.Current.TemperatureC),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := `{"city":"` + city + `","celsius":` + strconv.FormatFloat(result.Celsius, 'f', 1, 64) +
		`,"fahrenheit":` + strconv.FormatFloat(result.Fahrenheit, 'f', 1, 64) +
		`,"kelvin":` + strconv.FormatFloat(result.Kelvin, 'f', 1, 64) + `}`

	w.Write([]byte(response))

	calculateTemps.End()
}
