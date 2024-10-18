package globals

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/birabittoh/myks"
	"github.com/kkdai/youtube/v2"
	"github.com/utking/extemplate"
)

var (
	Debug bool
	Proxy bool
	Port  string

	C  = http.DefaultClient
	YT = youtube.Client{}
	XT = extemplate.New().Funcs(funcMap)

	KS  = myks.New[youtube.Video](3 * time.Hour)
	PKS *myks.KeyStore[bytes.Buffer]

	AdminUser string
	AdminPass string

	funcMap = template.FuncMap{"parseFormat": parseFormat}
)

func parseFormat(f youtube.Format) (res string) {
	isAudio := f.QualityLabel == ""

	if isAudio {
		bitrate := f.AverageBitrate
		if bitrate == 0 {
			bitrate = f.Bitrate
		}
		res = strconv.Itoa(bitrate/1000) + "kbps"
	} else {
		res = f.QualityLabel
	}

	mime := strings.Split(f.MimeType, ";")
	res += " - " + mime[0]

	codecs := " (" + strings.Split(mime[1], "\"")[1] + ")"

	if isAudio {
		return res + " - audio only" + codecs
	}

	res += fmt.Sprintf(" (%d FPS)", f.FPS)

	if f.AudioChannels == 0 {
		res += " - video only"
	}

	res += codecs
	return
}
