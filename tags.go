package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func loadTagsFromFile(fileName string) (map[string]string, error) {
	tags := map[string]string{}
	if _, err := os.Stat(fileName); err == nil {
		tagData, err := ioutil.ReadFile(fileName)
		if err != nil {
			return tags, fmt.Errorf("Failed to read tags file: %s", err)
		}
		if err := json.Unmarshal(tagData, &tags); err != nil {
			return tags, fmt.Errorf("Failed to decode tags file: %s", err)
		}
	}
	return tags, nil
}
