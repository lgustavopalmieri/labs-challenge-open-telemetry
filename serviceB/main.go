package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")


		response := map[string]string{
			"message": "Olá, mundo!",
		}

		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8181", nil)
}
