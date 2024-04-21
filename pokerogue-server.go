package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pagefaultgames/pokerogue-server/api"
	"github.com/pagefaultgames/pokerogue-server/db"
)

func main() {
	// flag stuff
	debug := flag.Bool("debug", false, "debug mode")

	proto := flag.String("proto", "tcp", "protocol for api to use (tcp, unix)")
	addr := flag.String("addr", "0.0.0.0", "network address for api to listen on")

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

	// create listener
	listener, err := createListener(*proto, *addr)
	if err != nil {
		log.Fatalf("failed to create net listener: %s", err)
	}

	// create exit handler
	var exit sync.RWMutex
	createExitHandler(&exit)

	// init api
	api.Init()

	// start web server
	err = http.Serve(listener, &api.Server{Debug: *debug, Exit: &exit})
	if err != nil {
		log.Fatalf("failed to create http server or server errored: %s", err)
	}
}

func createListener(proto, addr string) (net.Listener, error) {
	if proto == "unix" {
		os.Remove(addr)
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	if proto == "unix" {
		os.Chmod(addr, 0777)
	}

	return listener, nil
}

func createExitHandler(mtx *sync.RWMutex) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// wait for exit signal of some kind
		<-s

		// block new requests and wait for existing ones to finish
		mtx.Lock()

		// bail
		os.Exit(0)
	}()
}
