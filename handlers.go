package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func admin(w http.ResponseWriter, r *http.Request) {
	versionDirs, err := listDir(telemetryDir)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	type version struct {
		Version   string
		FileCount int
	}
	versions := make([]version, 0, len(versionDirs))
	for _, versionDir := range versionDirs {
		files, err := listDir(filepath.Join(telemetryDir, versionDir))
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		versions = append(versions, version{
			Version:   filepath.Base(versionDir),
			FileCount: len(files),
		})
	}
	data := map[string]interface{}{
		"Versions": versions,
	}
	if err := templates.ExecuteTemplate(w, "admin.html.tpl", data); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
}

func adminTelemetryDownload(w http.ResponseWriter, r *http.Request) {
	version := filepath.Clean(r.URL.Query().Get("version"))
	versionDir := filepath.Join(telemetryDir, version)
	if _, err := os.Stat(versionDir); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	var sessions []session
	files, err := listDir(versionDir)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}

	for _, file := range files {
		fileSessions, err := readSessionsFromFile(filepath.Join(versionDir, file))
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		sessions = append(sessions, fileSessions...)
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.json.gz"`, version, time.Now().UTC().Format("2006-01-02-15-04-05")))
	w.Header().Set("Content-Type", "application/gzip")
	w.WriteHeader(http.StatusOK)
	writeSessions(sessions, w)
}

func adminTelemetryClear(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	version := filepath.Clean(body.Version)
	versionDir := filepath.Join(telemetryDir, version)
	if _, err := os.Stat(versionDir); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	if err := os.RemoveAll(versionDir); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func adminBuildsUpload(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	if filename == "" || filename == "/" || filename == "." || filepath.Clean(filename) != filename {
		httpError(w, http.StatusBadRequest, fmt.Errorf("invalid filename: %s", filename))
		return
	}
	path := filepath.Join(buildsDir, filename)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		httpError(w, http.StatusBadRequest, fmt.Errorf("file already exists: %w", err))
		return
	}
	f, err := os.Create(path)
	if err != nil {
		httpError(w, http.StatusInternalServerError, fmt.Errorf("internal error: %w", err))
		return
	}
	defer f.Close()
	if _, err := io.Copy(f, r.Body); err != nil {
		httpError(w, http.StatusInternalServerError, fmt.Errorf("internal error: %w", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func index(w http.ResponseWriter, r *http.Request) {
	builds, err := listDir(buildsDir)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	data := map[string]interface{}{
		"Builds": builds,
	}
	if err := templates.ExecuteTemplate(w, "index.html.tpl", data); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
}

func buildsDownload(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	if filename == "" || filename == "/" || filename == "." || filepath.Clean(filename) != filename {
		httpError(w, http.StatusNotFound, fmt.Errorf("invalid filename: %s", filename))
		return
	}
	path := filepath.Join(buildsDir, filename)
	stat, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		httpError(w, http.StatusNotFound, fmt.Errorf("build not found: %s", filename))
		return
	} else if err != nil {
		httpError(w, http.StatusInternalServerError, fmt.Errorf("internal error: %w", err))
		return
	}

	f, err := os.Open(path)
	if err != nil {
		httpError(w, http.StatusInternalServerError, fmt.Errorf("internal error: %w", err))
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))
	w.Header().Set("Content-Length", fmt.Sprint(stat.Size()))
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, f)
}

var sessionIDRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func telemetry(w http.ResponseWriter, r *http.Request) {
	sessions, err := readSessions(r.Body)
	if err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	if len(sessions) == 0 {
		httpError(w, http.StatusBadRequest, fmt.Errorf("no sessions found"))
		return
	}
	if sessions[0].Version == "" || sessions[0].Version != filepath.Clean(sessions[0].Version) {
		httpError(w, http.StatusBadRequest, fmt.Errorf("invalid version string %q", sessions[0].Version))
		return
	}
	for _, sess := range sessions[1:] {
		if sess.Version != sessions[0].Version {
			httpError(w, http.StatusBadRequest, fmt.Errorf("inconsistent versions: %q and %q", sessions[0].Version, sess.Version))
			return
		}
		if sess.Device.ID != sessions[0].Device.ID {
			httpError(w, http.StatusBadRequest, fmt.Errorf("inconsistent device IDs: %q and %q", sessions[0].Device.ID, sess.Device.ID))
			return
		}
		if !sessionIDRegex.MatchString(sess.ID) {
			httpError(w, http.StatusBadRequest, fmt.Errorf("invalid session ID: %q", sess.ID))
			return
		}
	}
	versionDir := filepath.Join(telemetryDir, sessions[0].Version)
	if err := os.MkdirAll(versionDir, os.ModePerm); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	filename := filepath.Clean(fmt.Sprintf("%s-%s.json.gz", time.Now().UTC().Format("2006-01-02-15-04-05"), sessions[0].Device.ID))
	path := filepath.Join(versionDir, filename)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		httpError(w, http.StatusBadRequest, fmt.Errorf("file already exists: %w", err))
		return
	}
	if err := writeSessionsToFile(sessions, path); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}
