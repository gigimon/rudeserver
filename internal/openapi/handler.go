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

const envSpecPath = "OPENAPI_SPEC_PATH"

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
	if env := os.Getenv(envSpecPath); env != "" {
		if fileExists(env) {
			return env, nil
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, "openapi", "openapi.yaml")
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "openapi", "openapi.yaml")
		if fileExists(candidate) {
			return candidate, nil
		}
	}

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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
