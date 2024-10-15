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
	emptyVideo  = youtube.Video{}
	expireRegex = regexp.MustCompile(`(?i)expire=(\d+)`)
)

func parseExpiration(url string) time.Duration {
	expireString := expireRegex.FindStringSubmatch(url)
	expireTimestamp, err := strconv.ParseInt(expireString[1], 10, 64)
	if err != nil {
		log.Println("parseExpiration ERROR: ", err)
		return defaultCacheDuration
	}

	return time.Until(time.Unix(expireTimestamp, 0))
}

func getFormat(video youtube.Video, formatID int) *youtube.Format {
	formats := video.Formats.Select(formatsSelectFn)
	l := len(formats)
	if l == 0 {
		return nil
	}

	return &formats[formatID%l]
}

func formatsSelectFn(f youtube.Format) bool {
	return f.AudioChannels > 1 && f.ContentLength < maxContentLength && strings.HasPrefix(f.MimeType, "video/mp4")
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

	if video.ID == emptyVideo.ID {
		err = errors.New("no formats for this video")
		return
	}

	format = getFormat(*video, formatID)
	return
}

func getFromYT(videoID string, formatID int) (video *youtube.Video, format *youtube.Format, err error) {
	url := getURL(videoID)

	log.Println("Requesting video ", url)
	video, err = g.YT.GetVideo(url)
	if err != nil {
		return
	}

	format = getFormat(*video, formatID)
	duration := defaultCacheDuration
	v := emptyVideo
	if format != nil {
		v = *video
		duration = parseExpiration(format.URL)
	}

	g.KS.Set(videoID, v, duration)
	return
}

func getVideo(videoID string, formatID int) (video *youtube.Video, format *youtube.Format, err error) {
	video, format, err = getFromCache(videoID, formatID)
	if err != nil {
		video, format, err = getFromYT(videoID, formatID)
	}
	return
}
