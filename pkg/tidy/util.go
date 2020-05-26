package tidy

import (
	"strings"
)

func Escape(str string) string {
	return strings.Replace(str, "/", "∕", -1)
}
