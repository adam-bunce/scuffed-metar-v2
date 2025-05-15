package util

import (
	"encoding/json"
	"net/http"
)

// RequestAndParse executed the request r and parses the body as json, placing the result into dest
func RequestAndParse[T any](url string, dest *T) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(dest)
}
