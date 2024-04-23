package handlers

type Location struct {
	Localtime string `json:"localtime"`
}

type CurrentWeather struct {
	TemperatureC float64 `json:"temp_c"`
}

type WeatherResponse struct {
	Location Location       `json:"location"`
	Current  CurrentWeather `json:"current"`
}

type Temperature struct {
	Celsius    float64 `json:"celsius"`
	Fahrenheit float64 `json:"fahrenheit"`
	Kelvin     float64 `json:"kelvin"`
}
