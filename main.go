package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const maxPayloadBytes = 10_000_000 // 10 MB

//go:embed templates
var templateFS embed.FS
var templates = template.Must(template.ParseFS(templateFS, "templates/*"))

var dataDir = getEnv("PLAYTEST_DATA_DIR", "data")
var buildsDir = filepath.Join(dataDir, "builds")
var telemetryDir = filepath.Join(dataDir, "telemetry")

func main() {
	if err := setup(); err != nil {
		log.Fatal(err)
	}
	basicAuth := basicAuthMiddleware(getEnv("PLAYTEST_USERNAME", "admin"), getEnv("PLAYTEST_PASSWORD", "password"))
	bearerAuth := bearerAuthMiddleware(getEnv("PLAYTEST_API_KEY", "testkey"))

	http.Handle("/admin", middlewareStack(
		logRequests,
		requireMethod(http.MethodGet),
		basicAuth,
	).WrapFunc(admin))

	http.Handle("/admin/builds/", middlewareStack(
		logRequests,
		requireMethod(http.MethodPost),
		basicAuth,
	).WrapFunc(adminBuildsUpload))

	http.Handle("/admin/telemetry/download", middlewareStack(
		logRequests,
		requireMethod(http.MethodGet),
		basicAuth,
	).WrapFunc(adminTelemetryDownload))

	http.Handle("/admin/telemetry/clear", middlewareStack(
		logRequests,
		requireMethod(http.MethodPost),
		basicAuth,
	).WrapFunc(adminTelemetryClear))

	http.Handle("/telemetry", middlewareStack(
		logRequests,
		requireMethod(http.MethodPost),
		bearerAuth,
		limitPayloadMiddleware(maxPayloadBytes),
	).WrapFunc(telemetry))

	http.Handle("/", middlewareStack(
		logRequests,
		requireMethod(http.MethodGet),
	).WrapFunc(index))

	http.Handle("/builds/", middlewareStack(
		logRequests,
		requireMethod(http.MethodGet),
	).WrapFunc(buildsDownload))

	http.ListenAndServe(getEnv("LISTEN_ADDR", ":8000"), nil)
}

func setup() error {
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(telemetryDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(buildsDir, os.ModePerm); err != nil {
		return err
	}
	return nil
}
