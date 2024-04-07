package api

import (
	"log"
	"net/http"
)

func httpError(w http.ResponseWriter, r *http.Request, error string, code int) {
	log.Printf("%s: %s\n", r.URL.Path, error)
	http.Error(w, error, code)
}
