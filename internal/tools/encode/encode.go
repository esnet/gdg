package encode

import (
	"log"
	"net/url"
	"os"
	"strings"
)

const pathJoiner = string(os.PathSeparator)

func Encode(s string) string {
	return url.QueryEscape(s)
}

func Decode(s string) string {
	res, err := url.QueryUnescape(s)
	if err != nil {
		log.Fatal("unable to decode string", "input", s)
	}
	return res
}

func EncodePath(s ...string) string {
	for i, p := range s {
		s[i] = Encode(p)
	}
	return strings.Join(s, pathJoiner)
}
