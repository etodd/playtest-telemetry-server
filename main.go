package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
)

const maxPayloadBytes = 10_000_000 // 10 MB

//go:embed index.html.tpl
var indexHTML string
var indexTemplate = template.Must(template.New("index").Parse(indexHTML))

var dataDir = getEnv("DATA_DIR", "data")

func main() {
	if err := setup(); err != nil {
		log.Fatal(err)
	}
	basicAuth := basicAuthMiddleware(getEnv("USERNAME", "admin"), getEnv("PASSWORD", "password"))
	bearerAuth := bearerAuthMiddleware(getEnv("API_KEY", "testkey"))

	http.Handle("/", middlewareStack(
		logRequests,
		requireMethod(http.MethodGet),
		basicAuth,
	).WrapFunc(index))

	http.Handle("/download", middlewareStack(
		logRequests,
		requireMethod(http.MethodGet),
		basicAuth,
	).WrapFunc(download))

	http.Handle("/clear", middlewareStack(
		logRequests,
		requireMethod(http.MethodPost),
		basicAuth,
	).WrapFunc(clear))

	http.Handle("/upload", middlewareStack(
		logRequests,
		requireMethod(http.MethodPost),
		bearerAuth,
		limitPayloadMiddleware(maxPayloadBytes),
	).WrapFunc(upload))

	http.ListenAndServe(getEnv("LISTEN_ADDR", ":8000"), nil)
}

func setup() error {
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return err
	}
	return nil
}
