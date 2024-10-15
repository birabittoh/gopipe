package globals

import (
	"github.com/birabittoh/myks"
	"github.com/kkdai/youtube/v2"
)

var (
	Debug bool
	Port  string

	YT = youtube.Client{}
	KS = myks.New[string](0)
)
