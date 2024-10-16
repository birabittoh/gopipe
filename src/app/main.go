package app

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	g "github.com/birabittoh/gopipe/src/globals"
	"github.com/birabittoh/myks"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func Main() {
	godotenv.Load()

	g.Debug = parseBool(os.Getenv("APP_DEBUG"))
	if g.Debug {
		log.Println("Debug mode enabled.")
	}

	g.Proxy = parseBool(os.Getenv("APP_PROXY"))
	if g.Proxy {
		g.PKS = myks.New[bytes.Buffer](3 * time.Minute)
		log.Println("Proxy mode enabled.")
	}

	g.Port = getEnvDefault("APP_PORT", "3000")

	r := http.NewServeMux()

	var serveMux http.Handler
	if g.Debug {
		serveMux = r
	} else {
		limiter := rate.NewLimiter(rate.Limit(1), 3) // TODO: use env vars
		serveMux = limit(limiter, r)
	}

	r.HandleFunc("GET /", indexHandler)

	// video handlers
	r.HandleFunc("GET /watch", videoHandler)
	r.HandleFunc("GET /shorts/{videoID}", videoHandler)
	r.HandleFunc("GET /{videoID}", videoHandler)
	r.HandleFunc("GET /{videoID}/{formatID}", videoHandler)

	r.HandleFunc("GET /proxy/{videoID}", proxyHandler)
	r.HandleFunc("GET /proxy/{videoID}/{formatID}", proxyHandler)

	// r.HandleFunc("GET /robots.txt", robotsHandler)

	log.Println("Serving on port " + g.Port)
	log.Fatal(http.ListenAndServe(":"+g.Port, serveMux))
}
