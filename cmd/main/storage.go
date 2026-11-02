//go:build !js

package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// settingsPath returns where settings are kept, in the user's configuration directory.
func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ultimate-tic-tac-toe", "settings.json"), nil
}

func readSettings() (settings, bool) {
	var s settings
	path, err := settingsPath()
	if err != nil {
		return s, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s, false
	}
	return s, json.Unmarshal(data, &s) == nil
}

func writeSettings(s settings) {
	path, err := settingsPath()
	if err == nil {
		err = os.MkdirAll(filepath.Dir(path), 0o755)
	}
	if err == nil {
		data, _ := json.Marshal(s)
		err = os.WriteFile(path, data, 0o644)
	}
	if err != nil {
		log.Printf("saving settings: %v", err)
	}
}

// prefersDark reports whether the system asks for a dark theme; unknown natively.
func prefersDark() bool { return false }
