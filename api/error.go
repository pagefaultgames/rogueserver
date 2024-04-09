package api

import (
	"log"
	"net/http"
)

func httpError(w http.ResponseWriter, r *http.Request, err error, code int) {
	log.Printf("%s: %s\n", r.URL.Path, err)
	http.Error(w, err.Error(), code)
}
