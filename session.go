package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
)

type session struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Device  struct {
		CPU       string `json:"cpu"`
		GPU       string `json:"gpu"`
		OSName    string `json:"os_name"`
		OSVersion string `json:"os_version"`
		OSDistro  string `json:"os_distro"`
		ID        string `json:"id"`
		Memory    struct {
			Physical  int64 `json:"physical"`
			Free      int64 `json:"free"`
			Available int64 `json:"available"`
			Stack     int64 `json:"stack"`
		} `json:"memory"`
		Locale     string `json:"locale"`
		WindowSize [2]int `json:"window_size"`
	} `json:"device"`
	Start          float64 `json:"start"`
	End            float64 `json:"end"`
	InGameDuration float64 `json:"in_game_duration"`
	Nodes          map[string]map[string]struct {
		Type int           `json:"type"`
		Data []interface{} `json:"data"`
	} `json:"nodes"`
}

func readSessions(r io.Reader) ([]session, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	var result []session
	if err := json.NewDecoder(gzipReader).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func writeSessions(sessions []session, w io.Writer) error {
	gzipWriter := gzip.NewWriter(w)
	if err := json.NewEncoder(gzipWriter).Encode(sessions); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}
	return nil
}

func readSessionsFromFile(path string) ([]session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readSessions(f)
}

func writeSessionsToFile(sessions []session, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeSessions(sessions, f)
}
