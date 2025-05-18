package util

import (
	"encoding/json"
	"net/http"
	"os"
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

// ReadFileToStruct reads the json file from the given filePath and unmarshal's its data to dest
// filePath is the relative path to the calling file
func ReadFileToStruct[T any](filePath string, dest *T) error {
	val, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(val, dest)
}
