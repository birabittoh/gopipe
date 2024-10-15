package globals

import (
	"time"

	"github.com/birabittoh/myks"
	"github.com/kkdai/youtube/v2"
)

var (
	Debug bool
	Port  string

	YT = youtube.Client{}
	KS = myks.New[youtube.Video](3 * time.Hour)
)
