package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"

	"github.com/pagefaultgames/pokerogue-server/api"
	"github.com/pagefaultgames/pokerogue-server/db"
)

func main() {
	// flag stuff
	addr := flag.String("addr", "0.0.0.0:80", "network address for api to listen on")
	wwwpath := flag.String("wwwpath", "www", "path to static content to serve")
	tlscert := flag.String("tlscert", "", "path to tls certificate to use for https")
	tlskey := flag.String("tlskey", "", "path to tls private key to use for https")

	dbuser := flag.String("dbuser", "pokerogue", "database username")
	dbpass := flag.String("dbpass", "", "database password")
	dbproto := flag.String("dbproto", "tcp", "protocol for database connection")
	dbaddr := flag.String("dbaddr", "localhost", "database address")
	dbname := flag.String("dbname", "pokeroguedb", "database name")

	flag.Parse()

	// register gob types
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

	// get database connection
	err := db.Init(*dbuser, *dbpass, *dbproto, *dbaddr, *dbname)
	if err != nil {
		log.Fatalf("failed to initialize database: %s", err)
	}

	// start web server
	mux := http.NewServeMux()

	api.Init(mux)

	mux.Handle("/", http.FileServer(http.Dir(*wwwpath)))
	
	if *tlscert != "" && *tlskey != "" {
		err = http.ListenAndServeTLS(*addr, *tlscert, *tlskey, mux)
	} else {
		err = http.ListenAndServe(*addr, mux)
	}
	if err != nil {
		log.Fatalf("failed to create http server or server errored: %s", err)
	}
}
