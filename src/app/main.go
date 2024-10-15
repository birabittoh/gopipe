package app

import (
	"log"
	"net/http"
	"os"
	"strings"

	g "github.com/birabittoh/gopipe/src/globals"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func limit(limiter *rate.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			status := http.StatusTooManyRequests
			http.Error(w, http.StatusText(status), status)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnvDefault(key string, def string) string {
	res := os.Getenv(key)
	if res == "" {
		return def
	}
	return res
}

func parseBool(s string) bool {
	return strings.ToLower(s) == "true" || s == "1"
}

func Main() {
	godotenv.Load()

	g.Debug = parseBool(os.Getenv("APP_DEBUG"))
	if g.Debug {
		log.Println("Debug mode enabled.")
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
	// r.HandleFunc("/proxy/{videoId}", proxyHandler)
	// r.HandleFunc("/clear", clearHandler)

	// video handlers
	r.HandleFunc("GET /watch", videoHandler)
	r.HandleFunc("GET /{videoID}", videoHandler)
	r.HandleFunc("GET /shorts/{videoID}", videoHandler)

	// r.HandleFunc("GET /robots.txt", robotsHandler)

	log.Println("Serving on port " + g.Port)
	log.Fatal(http.ListenAndServe(":"+g.Port, serveMux))
}
