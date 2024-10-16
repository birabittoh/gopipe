package app

import (
	"os"
	"strconv"
	"strings"
)

func getEnvDefault(key string, def string) string {
	res := os.Getenv(key)
	if res == "" {
		return def
	}
	return res
}

func parseBool(s string) bool {
	return strings.ToLower(s) == "true" || s == "1"
}

func getFormatID(s string) int {
	formatID, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		formatID = 1
	}
	return int(formatID)
}
