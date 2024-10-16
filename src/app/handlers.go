package app

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"text/template"

	g "github.com/birabittoh/gopipe/src/globals"
	"golang.org/x/time/rate"
)

const (
	fmtYouTubeURL = "https://www.youtube.com/watch?v=%s"
	err404        = "Not Found"
	err500        = "Internal Server Error"
)

var (
	templates      = template.Must(template.ParseGlob("templates/*.html"))
	userAgentRegex = regexp.MustCompile(`(?i)bot|facebook|embed|got|firefox\/92|firefox\/38|curl|wget|go-http|yahoo|generator|whatsapp|preview|link|proxy|vkshare|images|analyzer|index|crawl|spider|python|cfnetwork|node`)
	videoRegex     = regexp.MustCompile(`(?i)^[a-z0-9_-]{11}$`)
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Println("indexHandler ERROR: ", err)
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
}

func videoHandler(w http.ResponseWriter, r *http.Request) {
	videoID := r.URL.Query().Get("v")
	if videoID == "" {
		videoID = r.PathValue("videoID")
		if videoID == "" {
			http.Error(w, "Missing video ID", http.StatusBadRequest)
			return
		}
	}

	if !userAgentRegex.MatchString(r.UserAgent()) {
		log.Println("Regex did not match. UA: ", r.UserAgent())
		if !g.Debug {
			log.Println("Redirecting.")
			http.Redirect(w, r, getURL(videoID), http.StatusFound)
			return
		}
	}

	if !videoRegex.MatchString(videoID) {
		log.Println("Invalid video ID: ", videoID)
		http.Error(w, "Invalid video ID.", http.StatusBadRequest)
		return
	}

	formatID := getFormatID(r.PathValue("formatID"))

	video, format, err := getVideo(videoID, formatID)
	if err != nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	if video == nil || format == nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	var thumbnail string
	if len(video.Thumbnails) > 0 {
		thumbnail = video.Thumbnails[len(video.Thumbnails)-1].URL
	}

	videoURL := format.URL
	if g.Proxy {
		videoURL = fmt.Sprintf("/proxy/%s/%d", videoID, formatID)
	}

	data := map[string]interface{}{
		"VideoID":     videoID,
		"VideoURL":    videoURL,
		"Uploader":    video.Author,
		"Title":       video.Title,
		"Description": video.Description,
		"Thumbnail":   thumbnail,
		"Duration":    video.Duration,
		"Debug":       g.Debug,
	}

	err = templates.ExecuteTemplate(w, "video.html", data)
	if err != nil {
		log.Println("indexHandler ERROR: ", err)
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
}
