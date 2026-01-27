package protocol

import (
	"net/http"

	"rudeserver/internal/scenario"
)

func WriteREST(w http.ResponseWriter, sc scenario.Scenario) {
	WriteHTTP(w, sc)
}
