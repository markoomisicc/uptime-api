package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type fieldError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logError("write json", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}

func writeFieldErrors(w http.ResponseWriter, status int, code, message string, fields map[string]fieldError) {
	writeJSON(w, status, map[string]any{"error": code, "message": message, "fields": fields})
}

func writeInternalError(w http.ResponseWriter, context string, err error) {
	logError(context, err)
	writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
}

func logError(context string, err error) {
	log.Printf("%s: %v", context, err)
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if decoder.More() {
		return errors.New("body must contain a single JSON object")
	}

	return nil
}
