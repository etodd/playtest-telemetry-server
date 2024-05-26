package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
)

func httpError(w http.ResponseWriter, code int, err error) {
	log.Printf("HTTP %d: %v", code, err)
	writeJSON(w, code, map[string]string{
		"error": err.Error(),
	})
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func listDir(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(files))
	for _, file := range files {
		result = append(result, file.Name())
	}
	// sort by filename descending
	slices.Reverse(result)
	return result, nil
}
