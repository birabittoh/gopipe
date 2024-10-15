package app

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
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

func getFormat(video youtube.Video) *youtube.Format {
	formats := video.Formats.Select(formatsSelectFn)
	if len(formats) == 0 {
		return nil
	}

	return &formats[0]
}

func formatsSelectFn(f youtube.Format) bool {
	return f.AudioChannels > 1 && f.ContentLength < maxContentLength
}

func getURL(videoID string) string {
	return fmt.Sprintf(fmtYouTubeURL, videoID)
}

func getFromCache(videoID string) (video *youtube.Video, format *youtube.Format, err error) {
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

	format = getFormat(*video)
	return
}

func getFromYT(videoID string) (video *youtube.Video, format *youtube.Format, err error) {
	url := getURL(videoID)

	log.Println("Requesting video ", url)
	video, err = g.YT.GetVideo(url)
	if err != nil {
		return
	}

	format = getFormat(*video)
	duration := defaultCacheDuration
	v := emptyVideo
	if format != nil {
		v = *video
		duration = parseExpiration(format.URL)
	}

	g.KS.Set(videoID, v, duration)
	return
}

func getVideo(videoID string) (video *youtube.Video, format *youtube.Format, err error) {
	video, format, err = getFromCache(videoID)
	if err != nil {
		video, format, err = getFromYT(videoID)
	}
	return
}
