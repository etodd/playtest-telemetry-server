package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func index(w http.ResponseWriter, r *http.Request) {
	versionDirs, err := listDir(dataDir)
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
		files, err := listDir(filepath.Join(dataDir, versionDir))
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
	if err := indexTemplate.Execute(w, data); err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	version := filepath.Clean(r.URL.Query().Get("version"))
	versionDir := filepath.Join(dataDir, version)
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

func clear(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	version := filepath.Clean(body.Version)
	versionDir := filepath.Join(dataDir, version)
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

var sessionIDRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func upload(w http.ResponseWriter, r *http.Request) {
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
	versionDir := filepath.Join(dataDir, sessions[0].Version)
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
