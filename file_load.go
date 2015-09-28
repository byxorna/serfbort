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

func loadConfigFromFile(configFile string) (Config, error) {
	var cfg Config
	if _, err := os.Stat(configFile); err != nil {
		return cfg, err
	}
	configRaw, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(configRaw, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
