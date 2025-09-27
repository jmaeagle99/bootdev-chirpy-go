package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

type Chirp struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ChirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

var bannedWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func Map[T any, U any](input []T, fn func(T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = fn(v)
	}
	return result
}

var bannedWordRegexps = Map(bannedWords, func(word string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
})

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
		return
	}

	content := chirp.Body
	for _, regexp := range bannedWordRegexps {
		content = regexp.ReplaceAllString(content, "****")
	}

	writeAsJson(
		w,
		ChirpResponse{
			CleanedBody: content,
		},
		http.StatusOK)
}

func writeAsJson(w http.ResponseWriter, value interface{}, statucode int) {
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
