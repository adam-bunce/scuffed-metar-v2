package util

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

// GetAndParseJson executed the request r and parses the body as json, placing the result into dest
func GetAndParseJson[T any](url string, dest *T) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(dest)
}

// GetAndParseJson executed the request r and parses the body as json, placing the result into dest
func GetAndParseString(url string, dest *string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	*dest = string(body)

	return nil
}

// RequestAndParse executed the request r and parses the body as json, placing the result into dest
func RequestAndParse[T any](r *http.Request, dest *T) error {
	res, err := http.DefaultClient.Do(r)
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
