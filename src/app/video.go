package app

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	g "github.com/birabittoh/gopipe/src/globals"
	"github.com/kkdai/youtube/v2"
)

const (
	maxMB                = 20
	maxContentLength     = maxMB * 1048576
	defaultCacheDuration = 6 * time.Hour
)

var (
	expireRegex = regexp.MustCompile(`(?i)expire=(\d+)`)
)

func parseExpiration(url string) time.Duration {
	expireString := expireRegex.FindStringSubmatch(url)
	if len(expireString) < 2 {
		return defaultCacheDuration
	}

	expireTimestamp, err := strconv.ParseInt(expireString[1], 10, 64)
	if err != nil {
		log.Println("parseExpiration ERROR:", err)
		return defaultCacheDuration
	}

	return time.Until(time.Unix(expireTimestamp, 0))
}

func getFormat(video youtube.Video, formatID int) *youtube.Format {
	if formatID != 0 {
		f := video.Formats.Select(formatsSelectFn)
		l := len(f)
		if l > 0 {
			return &f[(formatID-1)%l]
		}
	}

	f := video.Formats.Select(formatsSelectFnBest)
	if len(f) > 0 {
		return &f[0]
	}

	return nil
}

func getCaptions(video youtube.Video) map[string]Captions {
	c := make(map[string]Captions)
	for _, caption := range video.CaptionTracks {
		c[caption.LanguageCode] = Captions{
			VideoID:  video.ID,
			Language: caption.LanguageCode,
			URL:      caption.BaseURL,
		}
	}

	return c
}

func formatsSelectFn(f youtube.Format) bool {
	return f.AudioChannels > 0 && f.ContentLength < maxContentLength && strings.HasPrefix(f.MimeType, "video/mp4")
}

func formatsSelectFnBest(f youtube.Format) bool {
	return f.AudioChannels > 0 && strings.HasPrefix(f.MimeType, "video/mp4")
}

func formatsSelectFnAudioVideo(f youtube.Format) bool {
	return f.AudioChannels > 0 && f.QualityLabel != ""
}

func formatsSelectFnVideo(f youtube.Format) bool {
	return f.QualityLabel != "" && f.AudioChannels == 0
}

func formatsSelectFnAudio(f youtube.Format) bool {
	return f.QualityLabel == ""
}

func getURL(videoID string) string {
	return fmt.Sprintf(fmtYouTubeURL, videoID)
}

func getFromCache(videoID string, formatID int) (video *youtube.Video, format *youtube.Format, err error) {
	video, err = g.KS.Get(videoID)
	if err != nil {
		return
	}

	if video == nil {
		err = errors.New("video should not be nil")
		return
	}

	format = getFormat(*video, formatID)
	return
}

func getFromYT(videoID string, formatID int) (video *youtube.Video, format *youtube.Format, err error) {
	url := getURL(videoID)

	const maxRetries = 3
	const maxBytesToCheck = 1024
	duration := defaultCacheDuration

	for i := 0; i < maxRetries; i++ {
		log.Println("Requesting video", url, "attempt", i+1)
		video, err = g.YT.GetVideo(url)
		if err != nil || video == nil {
			log.Println("Error fetching video info:", err)
			continue
		}

		format = getFormat(*video, formatID)
		if format != nil {
			duration = parseExpiration(format.URL)
		}

		resp, err := g.C.Get(format.URL)
		if err != nil {
			log.Println("Error fetching video URL:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.ContentLength <= 0 {
			log.Println("Invalid video link, no content length...")
			continue
		}

		buffer := make([]byte, maxBytesToCheck)
		n, err := resp.Body.Read(buffer)
		if err != nil {
			log.Println("Error reading video content:", err)
			continue
		}

		if n > 0 {
			log.Println("Valid video link found.")
			g.KS.Set(videoID, *video, duration)
			return video, format, nil
		}

		log.Println("Invalid video link, content is empty...")
		time.Sleep(1 * time.Second)
	}

	err = fmt.Errorf("failed to fetch valid video after %d attempts", maxRetries)
	return nil, nil, err
}

func getVideo(videoID string, formatID int) (video *youtube.Video, format *youtube.Format, err error) {
	video, format, err = getFromCache(videoID, formatID)
	if err != nil {
		video, format, err = getFromYT(videoID, formatID)
	}
	return
}
