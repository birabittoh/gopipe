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

func parseExpiration(url string) (time.Duration, error) {
	expireString := expireRegex.FindStringSubmatch(url)
	expireTimestamp, err := strconv.ParseInt(expireString[1], 10, 64)
	if err != nil {
		log.Println("parseExpiration ERROR: ", err)
		return time.Duration(0), err
	}

	return time.Until(time.Unix(expireTimestamp, 0)), nil
}

func formatsSelectFn(f youtube.Format) bool {
	return f.AudioChannels > 1 && f.ContentLength < maxContentLength
}

func getURL(videoID string) string {
	return fmt.Sprintf(fmtYouTubeURL, videoID)
}

func getFromCache(videoID string) (video *youtube.Video, format youtube.Format, err error) {
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

	formats := video.Formats.Select(formatsSelectFn)
	if len(formats) == 0 {
		err = errors.New("no formats for this video")
		return
	}

	format = formats[0]
	return
}

func getFromYT(videoID string) (video *youtube.Video, format youtube.Format, err error) {
	video, err = g.YT.GetVideo(getURL(videoID))
	if err != nil {
		return
	}

	formats := video.Formats.Select(formatsSelectFn)
	if len(formats) == 0 {
		g.KS.Set(videoID, emptyVideo, defaultCacheDuration)
		return
	}

	format = formats[0]
	expiration, err := parseExpiration(format.URL)
	if err != nil {
		expiration = defaultCacheDuration
	}

	g.KS.Set(videoID, *video, expiration)
	return
}

func getVideo(videoID string) (video *youtube.Video, format youtube.Format, err error) {
	video, format, err = getFromCache(videoID)
	if err != nil {
		video, format, err = getFromYT(videoID)
	}
	return
}
