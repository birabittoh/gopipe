package app

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"text/template"
	"time"

	g "github.com/birabittoh/gopipe/src/globals"
)

const (
	fmtYouTubeURL = "https://www.youtube.com/watch?v=%s"
	err500        = "Internal Server Error"
	urlDuration   = 6 * time.Hour
)

var (
	templates      = template.Must(template.ParseGlob("templates/*.html"))
	userAgentRegex = regexp.MustCompile(`(?i)bot|facebook|embed|got|firefox\/92|firefox\/38|curl|wget|go-http|yahoo|generator|whatsapp|preview|link|proxy|vkshare|images|analyzer|index|crawl|spider|python|cfnetwork|node`)
	videoRegex     = regexp.MustCompile(`(?i)^[a-z0-9_-]{11}$`)
)

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

	formatID, err := strconv.ParseUint(r.PathValue("formatID"), 10, 64)
	if err != nil {
		formatID = 0
	}

	video, format, err := getVideo(videoID, int(formatID))
	if err != nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	if video == nil || format == nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"VideoID":     videoID,
		"VideoURL":    format.URL,
		"Uploader":    video.Author,
		"Title":       video.Title,
		"Description": video.Description,
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
