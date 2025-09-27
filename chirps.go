package main

import (
	"encoding/json"
	"net/http"
)

type Chirp struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ChirpValidationResponse struct {
	Valid bool `json:"valid"`
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	chirp := Chirp{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&chirp)
	if err != nil {
		writeServerError(w)
		return
	}

	if len(chirp.Body) > 140 {
		writeAsJson(
			w,
			ErrorResponse{
				Error: "Chirp is too long",
			},
			http.StatusBadRequest)
	} else {
		writeAsJson(
			w,
			ChirpValidationResponse{
				Valid: true,
			},
			http.StatusOK)
	}
}

func writeAsJson(w http.ResponseWriter, value any, statucode int) {
	data, err := json.Marshal(value)
	if err != nil {
		writeServerError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statucode)
	w.Write(data)
}

func writeServerError(w http.ResponseWriter) {
	errorResp := ErrorResponse{
		Error: "Something went wrong",
	}

	w.WriteHeader(http.StatusInternalServerError)

	data, err := json.Marshal(errorResp)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
