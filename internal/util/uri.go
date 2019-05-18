package util

import (
	"io"
	"net/http"
	"os"
	"strings"
)

//OpenURI a uri (http, https, file or nothing)
func OpenURI(name string) (io.ReadCloser, error) {
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
		resp, err := http.Get(name)
		if err != nil {
			return nil, err
		}
		return resp.Body, err
	} else if strings.HasPrefix(name, "file://") {
		runes := []rune(name)
		return os.Open(string(runes[7:]))
	}
	return os.Open(name)
}
