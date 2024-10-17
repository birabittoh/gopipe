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

	// read env vars
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

	g.AdminUser = getEnvDefault("APP_ADMIN_USER", "admin")
	g.AdminPass = getEnvDefault("APP_ADMIN_PASS", "admin")

	if g.AdminUser == "admin" && g.AdminPass == "admin" {
		log.Println("Admin credentials not set. Please set APP_ADMIN_USER and APP_ADMIN_PASS.")
	}

	// set up extemplate
	err := g.XT.ParseDir("templates", []string{".tmpl"})
	if err != nil {
		log.Fatal(err)
	}

	// set up http server
	r := http.NewServeMux()

	var serveMux http.Handler
	if g.Debug {
		serveMux = r
	} else {
		limiter := rate.NewLimiter(rate.Limit(1), 3) // TODO: use env vars
		serveMux = limit(limiter, r)
	}

	r.HandleFunc("GET /", indexHandler)

	r.HandleFunc("GET /watch", videoHandler)
	r.HandleFunc("GET /shorts/{videoID}", videoHandler)
	r.HandleFunc("GET /{videoID}", videoHandler)
	r.HandleFunc("GET /{videoID}/{formatID}", videoHandler)

	r.HandleFunc("GET /proxy/{videoID}", proxyHandler)
	r.HandleFunc("GET /proxy/{videoID}/{formatID}", proxyHandler)
	r.HandleFunc("GET /sub/{videoID}/{language}", subHandler)

	r.HandleFunc("GET /cache", cacheHandler)

	log.Println("Serving on port " + g.Port)
	log.Fatal(http.ListenAndServe(":"+g.Port, serveMux))
}
