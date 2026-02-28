package encode

import (
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const pathJoiner = string(os.PathSeparator)

func Encode(s string) string {
	return url.QueryEscape(s)
}

// TODO: there should be a better way of doing this
// unquoteMeta is a function that unescapes characters in a string that were escaped by regexp.QuoteMeta.
func unquoteMeta(s string) string {
	// A map to store the escaped characters and their unescaped counterparts
	// Note: This only covers the characters escaped by regexp.QuoteMeta
	replacements := map[string]string{
		"\\.":  ".",
		"\\+":  "+",
		"\\*":  "*",
		"\\?":  "?",
		"\\(":  "(",
		"\\)":  ")",
		"\\[":  "[",
		"\\]":  "]",
		"\\{":  "{",
		"\\}":  "}",
		"\\^":  "^",
		"\\$":  "$",
		"\\|":  "|",
		"\\\\": "\\", // Handle the escaped backslash itself
	}

	// Iterate through the replacements and perform string replacement
	for escaped, unescaped := range replacements {
		s = strings.ReplaceAll(s, escaped, unescaped)
	}
	return s
}

func EncodeEscapeSpecialChars(s string) string {
	s = Encode(s)
	return regexp.QuoteMeta(s)
}

func DecodeEscapeSpecialChars(s string) string {
	s = unquoteMeta(s)
	return Decode(s)
}

func Decode(s string) string {
	res, err := url.QueryUnescape(s)
	if err != nil {
		log.Fatal("unable to decode string", "input", s)
	}
	return res
}

func EncodePath(encoder func(s string) string, s ...string) string {
	if encoder == nil {
		encoder = Encode
	}
	if len(s) == 1 {
		s = strings.Split(s[0], string(os.PathSeparator))
	}

	for i, p := range s {
		s[i] = encoder(p)
	}

	return strings.Join(s, pathJoiner)
}
