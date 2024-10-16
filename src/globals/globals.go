package globals

import (
	"bytes"
	"net/http"
	"time"

	"github.com/birabittoh/myks"
	"github.com/kkdai/youtube/v2"
)

var (
	Debug bool
	Proxy bool
	Port  string

	C  = http.DefaultClient
	YT = youtube.Client{}

	KS  = myks.New[youtube.Video](3 * time.Hour)
	PKS *myks.KeyStore[bytes.Buffer]
)
