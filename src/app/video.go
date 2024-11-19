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

	log.Println("Requesting video", url)
	video, err = g.YT.GetVideo(url)
	if err != nil || video == nil {
		return
	}

	format = getFormat(*video, formatID)
	duration := defaultCacheDuration
	if format != nil {
		duration = parseExpiration(format.URL)
	}

	g.KS.Set(videoID, *video, duration)
	return
}

func getVideo(videoID string, formatID int) (video *youtube.Video, format *youtube.Format, err error) {
	video, format, err = getFromCache(videoID, formatID)
	if err != nil {
		video, format, err = getFromYT(videoID, formatID)
	}
	return
}
