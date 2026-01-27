package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Register loads the spec-first OpenAPI artifact and serves it as YAML and JSON.
func Register(mux *http.ServeMux) error {
	specPath, err := resolveSpecPath()
	if err != nil {
		return fmt.Errorf("resolve openapi spec path: %w", err)
	}

	yamlBytes, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read openapi spec: %w", err)
	}

	var doc any
	if err := yaml.Unmarshal(yamlBytes, &doc); err != nil {
		return fmt.Errorf("parse openapi yaml: %w", err)
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal openapi json: %w", err)
	}

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(yamlBytes)
	})

	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonBytes)
	})

	return nil
}

func resolveSpecPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}

	dir := filepath.Dir(filename)
	for {
		if dir == string(filepath.Separator) || dir == "." {
			return "", fmt.Errorf("go.mod not found while resolving spec path")
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "openapi", "openapi.yaml"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found while resolving spec path")
		}
		dir = parent
	}
}
