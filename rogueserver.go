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
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/pagefaultgames/rogueserver/api"
	"github.com/pagefaultgames/rogueserver/api/account"
	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/db"
)

func main() {
	// env stuff
	debug, _ := strconv.ParseBool(os.Getenv("debug"))

	proto := getEnv("proto", "tcp")
	addr := getEnv("addr", "0.0.0.0:8001")
	tlscert := getEnv("tlscert", "")
	tlskey := getEnv("tlskey", "")

	dbuser := getEnv("dbuser", "pokerogue")
	dbpass := getEnv("dbpass", "pokerogue")
	dbproto := getEnv("dbproto", "tcp")
	dbaddr := getEnv("dbaddr", "localhost")
	dbname := getEnv("dbname", "pokeroguedb")

	redisaddr := getEnv("redisaddr", "localhost:6379")
	redispass := getEnv("redispass", "")
	redisdb := getEnv("redisdb", "0")

	discordclientid := getEnv("discordclientid", "")
	discordsecretid := getEnv("discordsecretid", "")

	googleclientid := getEnv("googleclientid", "")
	googlesecretid := getEnv("googlesecretid", "")

	callbackurl := getEnv("callbackurl", "http://localhost:8001/")

	gameurl := getEnv("gameurl", "https://pokerogue.net")

	discordbottoken := getEnv("discordbottoken", "")
	discordguildid := getEnv("discordguildid", "")

	account.GameURL = gameurl

	account.DiscordClientID = discordclientid
	account.DiscordClientSecret = discordsecretid
	account.DiscordCallbackURL = callbackurl + "/auth/discord/callback"

	account.GoogleClientID = googleclientid
	account.GoogleClientSecret = googlesecretid
	account.GoogleCallbackURL = callbackurl + "/auth/google/callback"
	account.DiscordSession, _ = discordgo.New("Bot " + discordbottoken)
	account.DiscordGuildID = discordguildid
	// register gob types
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

	// get database connection
	err := db.Init(dbuser, dbpass, dbproto, dbaddr, dbname)
	if err != nil {
		log.Fatalf("failed to initialize database: %s", err)
	}

	// get redis connection
	err = cache.InitRedis(redisaddr, redispass, redisdb)
	if err != nil {
		log.Fatalf("failed to initialize redis: %s", err)
	}

	// create listener
	listener, err := createListener(proto, addr)
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
	if debug {
		handler = debugHandler(mux)
	}

	if tlscert == "" {
		err = http.Serve(listener, handler)
	} else {
		err = http.ServeTLS(listener, handler, tlscert, tlskey)
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

func prodHandler(router *http.ServeMux, clienturl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST")
		w.Header().Set("Access-Control-Allow-Origin", clienturl)

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

func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
