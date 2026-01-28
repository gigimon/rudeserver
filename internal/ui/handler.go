package ui

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
)

//go:embed assets/*
var assets embed.FS

func Handler() (http.Handler, error) {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			serveFile(sub, "index.html", w, r)
			return
		case path.Dir(r.URL.Path) == "/ui" || path.Dir(r.URL.Path) == "/ui/assets":
			http.StripPrefix("/ui", fileServer).ServeHTTP(w, r)
			return
		default:
			http.NotFound(w, r)
			return
		}
	}), nil
}

func serveFile(fsys fs.FS, name string, w http.ResponseWriter, r *http.Request) {
	file, err := fsys.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, _ := file.Stat()
	content, err := io.ReadAll(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	reader := bytes.NewReader(content)
	http.ServeContent(w, r, name, info.ModTime(), reader)
}
