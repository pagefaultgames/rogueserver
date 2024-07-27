/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/pagefaultgames/rogueserver/api"
	"github.com/pagefaultgames/rogueserver/db"
)

func main() {
	// flag stuff
	debug := flag.Bool("debug", false, "use debug mode")

	proto := flag.String("proto", "tcp", "protocol for api to use (tcp, unix)")
	addr := flag.String("addr", "0.0.0.0:8001", "network address for api to listen on")
	tlscert := flag.String("tlscert", "", "tls certificate path")
	tlskey := flag.String("tlskey", "", "tls key path")

	dbuser := flag.String("dbuser", "pokerogue", "database username")
	dbpass := flag.String("dbpass", "pokerogue", "database password")
	dbproto := flag.String("dbproto", "tcp", "protocol for database connection")
	dbaddr := flag.String("dbaddr", "localhost", "database address")
	dbname := flag.String("dbname", "pokeroguedb", "database name")

	discordclientid := flag.String("discordclientid", "dcid", "Discord Oauth2 Client ID")
	discordsecretid := flag.String("discordsecretid", "dsid", "Discord Oauth2 Secret ID")

	googleclientid := flag.String("googleclientid", "gcid", "Google Oauth2 Client ID")
	googlesecretid := flag.String("googlesecretid", "gsid", "Google Oauth2 Secret ID")
	callbackurl := flag.String("callbackurl", "http://localhost:8001/", "Callback URL for Oauth2 Client")

	gameurl := flag.String("gameurl", "https://pokerogue.net", "URL for game server")

	flag.Parse()

	// set discord client id as env variable
	os.Setenv("DISCORD_CLIENT_ID", *discordclientid)
	os.Setenv("DISCORD_CLIENT_SECRET", *discordsecretid)
	os.Setenv("DISCORD_CALLBACK_URL", *callbackurl+"/auth/discord/callback")

	os.Setenv("GOOGLE_CLIENT_ID", *googleclientid)
	os.Setenv("GOOGLE_CLIENT_SECRET", *googlesecretid)
	os.Setenv("GOOGLE_CALLBACK_URL", *callbackurl+"/auth/google/callback")
	os.Setenv("GAME_URL", *gameurl)

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

	mux := http.NewServeMux()

	// init api
	if err := api.Init(mux); err != nil {
		log.Fatal(err)
	}

	// start web server
	handler := prodHandler(mux, gameurl)
	if *debug {
		handler = debugHandler(mux)
	}

	if *tlscert == "" {
		err = http.Serve(listener, handler)
	} else {
		err = http.ServeTLS(listener, handler, *tlscert, *tlskey)
	}
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
		if err := os.Chmod(addr, 0777); err != nil {
			listener.Close()
			return nil, err
		}
	}

	return listener, nil
}

func prodHandler(router *http.ServeMux, clienturl *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST")
		w.Header().Set("Access-Control-Allow-Origin", *clienturl)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		router.ServeHTTP(w, r)
	})
}

func debugHandler(router *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		router.ServeHTTP(w, r)
	})
}
