package app

import (
	"log"
	"net/http"
	"regexp"
	"text/template"
	"time"

	g "github.com/birabittoh/gopipe/src/globals"
)

const err500 = "Internal Server Error"
const urlDuration = 6 * time.Hour

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

	url := "https://www.youtube.com/watch?v=" + videoID

	if !userAgentRegex.MatchString(r.UserAgent()) {
		log.Println("Regex did not match.")
		if !g.Debug {
			log.Println("Redirecting. UA:", r.UserAgent())
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
	}

	if !videoRegex.MatchString(videoID) {
		log.Println("Invalid video ID: ", videoID)
		http.Error(w, "Invalid video ID.", http.StatusBadRequest)
		return
	}

	video, err := g.YT.GetVideo(url)
	if err != nil {
		log.Println("videoHandler ERROR: ", err)
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		log.Println("videoHandler ERROR: ", err)
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	// TODO: check formats[i].ContentLength
	// g.KS.Set(videoID, formats[0].URL, urlDuration)

	data := map[string]interface{}{
		"VideoID":     videoID,
		"VideoURL":    formats[0].URL,
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
