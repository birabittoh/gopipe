package app

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	g "github.com/birabittoh/gopipe/src/globals"
	"github.com/birabittoh/gopipe/src/subs"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if !g.Proxy {
		http.Error(w, err404, http.StatusNotFound)
		return
	}

	videoID := r.PathValue("videoID")
	formatID := getFormatID(r.PathValue("formatID"))

	video, err := g.KS.Get(videoID)
	if err != nil || video == nil {
		http.Error(w, err404, http.StatusNotFound)
		return
	}

	format := getFormat(*video, formatID)
	if format == nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	key := fmt.Sprintf("%s:%d", videoID, formatID)
	content, err := g.PKS.Get(key)
	if err == nil && content != nil {
		log.Println("Using cached content for ", key)
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", strconv.Itoa(content.Len()))
		w.Write(content.Bytes())
		return
	}

	res, err := g.C.Get(format.URL)
	if err != nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", res.Header.Get("Content-Length"))

	pr, pw := io.Pipe()

	// Save the content to the cache asynchronously
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, pr)
		if err != nil {
			log.Println("Error while copying to buffer for cache:", err)
			return
		}

		g.PKS.Set(key, buf, 5*time.Minute)
		pw.Close()
	}()

	// Stream the content to the client while it's being downloaded and piped
	_, err = io.Copy(io.MultiWriter(w, pw), res.Body)
	if err != nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
}

func subHandler(w http.ResponseWriter, r *http.Request) {
	videoID := r.PathValue("videoID")

	video, err := g.KS.Get(videoID)
	if err != nil || video == nil {
		http.Error(w, err404, http.StatusNotFound)
		return
	}

	captions := getCaptions(*video)
	caption, ok := captions[strings.TrimSuffix(r.PathValue("language"), ".vtt")]
	if !ok {
		http.Error(w, err404, http.StatusNotFound)
		return
	}

	res, err := g.C.Get(caption.URL)
	if err != nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	w.Header().Set("Content-Type", "text/vtt")

	content, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}

	err = subs.Parse(content, w)
	if err != nil {
		http.Error(w, err500, http.StatusInternalServerError)
		return
	}
}
