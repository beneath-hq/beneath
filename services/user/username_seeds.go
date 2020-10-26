package user

import (
	"regexp"
	"strings"
	"unicode"

	uuid "github.com/satori/go.uuid"
)

const (
	usernameMinLength = 3
	usernameMaxLength = 50
)

var nonAlphanumericRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func usernameSeeds(email string, name string, nickname string) []string {
	// gather candidates
	var seeds []string
	var shortest string

	// base on nickname
	username := nonAlphanumericRegex.ReplaceAllString(nickname, "")
	if len(username) >= usernameMinLength {
		seeds = append(seeds, finalizeUsernameSeed(username))
		shortest = username
	}
	if len(username) < len(shortest) {
		shortest = username
	}

	// base on email
	username = strings.Split(email, "@")[0]
	username = nonAlphanumericRegex.ReplaceAllString(username, "")
	if len(username) >= usernameMinLength {
		seeds = append(seeds, finalizeUsernameSeed(username))
	}
	if len(username) < len(shortest) {
		shortest = username
	}

	// base on name
	username = nonAlphanumericRegex.ReplaceAllString(name, "")
	if len(username) >= usernameMinLength {
		seeds = append(seeds, finalizeUsernameSeed(username))
	}
	if len(username) < len(shortest) {
		shortest = username
	}

	// final fallback -- a uuid
	username = shortest + uuid.NewV4().String()[0:8]
	seeds = append(seeds, finalizeUsernameSeed(username))

	return seeds
}

func finalizeUsernameSeed(seed string) string {
	seed = strings.ToLower(seed)
	if len(seed) == (usernameMinLength-1) && seed[0] != '_' {
		seed = "_" + seed
	}
	if unicode.IsDigit(rune(seed[0])) {
		seed = "b" + seed
	}
	if len(seed) > usernameMaxLength {
		seed = seed[0:usernameMaxLength]
	}
	return seed
}
