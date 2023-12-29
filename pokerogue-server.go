package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/Flashfyre/pokerogue-server/api"
	"github.com/Flashfyre/pokerogue-server/db"
)

func main() {
	proto := flag.String("proto", "tcp", "protocol for api to use (tcp, unix)")
	addr := flag.String("addr", "0.0.0.0", "network address for api to listen on")

	dbuser := flag.String("dbuser", "pokerogue", "database username")
	dbpass := flag.String("dbpass", "", "database password")
	dbproto := flag.String("dbproto", "tcp", "protocol for database connection")
	dbaddr := flag.String("dbaddr", "localhost", "database address")
	dbname := flag.String("dbname", "pokeroguedb", "database name")

	flag.Parse()

	err := db.Init(*dbuser, *dbpass, *dbproto, *dbaddr, *dbname)
	if err != nil {
		log.Fatalf("failed to initialize database: %s", err)
	}

	if *proto == "unix" {
		os.Remove(*addr)
	}

	listener, err := net.Listen(*proto, *addr)
	if err != nil {
		log.Fatalf("failed to create net listener: %s", err)
	}

	if *proto == "unix" {
		os.Chmod(*addr, 0777)
	}

	// account
	http.HandleFunc("/account/info", api.HandleAccountInfo)
	http.HandleFunc("/account/register", api.HandleAccountRegister)
	http.HandleFunc("/account/login", api.HandleAccountLogin)
	http.HandleFunc("/account/logout", api.HandleAccountLogout)

	// savedata
	http.HandleFunc("/savedata/get", api.HandleSavedataGet)
	http.HandleFunc("/savedata/update", api.HandleSavedataUpdate)
	http.HandleFunc("/savedata/delete", api.HandleSavedataDelete)
	
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalf("failed to create http server or server errored: %s", err)
	}
}
