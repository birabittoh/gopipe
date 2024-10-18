package app

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"

	g "github.com/birabittoh/gopipe/src/globals"
	"github.com/kkdai/youtube/v2"
	"golang.org/x/time/rate"
)

const (
	fmtYouTubeURL = "https://www.youtube.com/watch?v=%s"
	err404        = "Not Found"
	err500        = "Internal Server Error"
	err401        = "Unauthorized"
	heading       = `<!--
 .d8888b.       88888888888       888              
d88P  Y88b          888           888              
888    888          888           888              
888         .d88b.  888  888  888 88888b.   .d88b. 
888  88888 d88""88b 888  888  888 888 "88b d8P  Y8b
888    888 888  888 888  888  888 888  888 88888888
Y88b  d88P Y88..88P 888  Y88b 888 888 d88P Y8b.    
 "Y8888P88  "Y88P"  888   "Y88888 88888P"   "Y8888 

A better way to embed YouTube videos on Telegram.
-->`
)

var (
	videoRegex = regexp.MustCompile(`(?i)^[a-z0-9_-]{11}$`)
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
	err := g.XT.ExecuteTemplate(w, "index.tmpl", nil)
	if err != nil {
		log.Println("indexHandler ERROR:", err)
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

	if !videoRegex.MatchString(videoID) {
		log.Println("Invalid video ID:", videoID)
		http.Error(w, err404, http.StatusNotFound)
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
		"VideoID":           videoID,
		"VideoURL":          videoURL,
		"Author":            video.Author,
		"Title":             video.Title,
		"Description":       video.Description,
		"Thumbnail":         thumbnail,
		"Duration":          video.Duration,
		"Captions":          getCaptions(*video),
		"Heading":           template.HTML(heading),
		"AudioVideoFormats": video.Formats.Select(formatsSelectFnAudioVideo),
		"VideoFormats":      video.Formats.Select(formatsSelectFnVideo),
		"AudioFormats":      video.Formats.Select(formatsSelectFnAudio),
	}

	err = g.XT.ExecuteTemplate(w, "video.tmpl", data)
	if err != nil {
		log.Println("indexHandler ERROR:", err)
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	videoID := r.FormValue("video")
	if videoID == "" {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}

	if !videoRegex.MatchString(videoID) {
		log.Println("Invalid video ID:", videoID)
		http.Error(w, err404, http.StatusNotFound)
		return
	}

	itagno := r.FormValue("itagno")
	if itagno == "" {
		http.Error(w, "Missing ItagNo", http.StatusBadRequest)
		return
	}

	video, err := g.KS.Get(videoID)
	if err != nil || video == nil {
		http.Error(w, err404, http.StatusNotFound)
		return
	}

	formats := video.Formats.Quality(itagno)
	if len(formats) == 0 {
		http.Error(w, err404, http.StatusNotFound)
		return
	}

	http.Redirect(w, r, formats[0].URL, http.StatusFound)
}

func cacheHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != g.AdminUser || password != g.AdminPass {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, err401, http.StatusUnauthorized)
		return
	}

	var videos []youtube.Video
	for s := range g.KS.Keys() {
		video, err := g.KS.Get(s)
		if err != nil || video == nil {
			continue
		}
		videos = append(videos, *video)
	}

	data := map[string]interface{}{"Videos": videos}
	err := g.XT.ExecuteTemplate(w, "cache.tmpl", data)
	if err != nil {
		log.Println("cacheHandler ERROR:", err)
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
}
