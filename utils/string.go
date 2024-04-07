package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CommonPrefix returns the common root of all the given strings.
func CommonPrefix(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}

	if len(strs) == 1 {
		return strs[0]
	}

	lengths := Map(strs, func(str string) int { return len(str) })
	maxLength := Min(lengths[0], lengths[1:]...)
	prefix := ""

	for c := range maxLength {
		if len(strs[0]) <= c {
			return prefix
		}

		prefix += string(strs[0][c])

		for _, str := range strs[1:] {
			if !strings.HasPrefix(str, prefix) {
				return prefix[:c]
			}
		}
	}

	return prefix
}

func GetRegexMatchGroup(re *regexp.Regexp, str string, group int) (string, bool) {
	matches := re.FindStringSubmatch(str)
	if len(matches) <= group {
		return "", false
	}

	return matches[group], true
}

func GetRegexNameGroup(re *regexp.Regexp, str, groupName string) (string, bool) {
	matches := re.FindStringSubmatch(str)
	if matches == nil {
		return "", false
	}

	for i, name := range re.SubexpNames() {
		if name == groupName {
			return matches[i], true
		}
	}

	return "", false
}

func FormatDuration(d time.Duration) string {
	res := ""

	if d >= 24*time.Hour {
		days := int(d / (24 * time.Hour))
		res += strconv.Itoa(days) + "d"
		d -= time.Duration(days) * 24 * time.Hour
	}

	return res + d.Round(time.Second).String()
}
