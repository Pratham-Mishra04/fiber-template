package utils

import (
	"regexp"
	"strings"
)

func SoftSlugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	// // Remove non-word characters except -
	// reg := regexp.MustCompile("[^a-zA-Z0-9-]")
	// s = reg.ReplaceAllString(s, "")

	// Replace multiple - with single -
	reg := regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")

	s = strings.Trim(s, " ")
	s = strings.Trim(s, "-")

	return s
}
