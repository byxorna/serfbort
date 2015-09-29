package main

import (
	"fmt"
	"strings"
)

// parses a slice of strings that look like key=val key2=val2 into a map
func parseTagArgs(tagsUnparsed []string) (map[string]string, error) {
	tagsRequired := map[string]string{}
	for _, t := range tagsUnparsed {
		vals := strings.Split(t, "=")
		if len(vals) != 2 {
			return tagsRequired, fmt.Errorf("tag must take parameters of the format tag=value! %s is fucked up", t)
		}
		tagsRequired[vals[0]] = vals[1]
	}
	return tagsRequired, nil
}

// parses a string that looks like host1,host2,host3 into a list of hosts
func parseHostArgs(hostsUnparsed string) []string {
	hosts := strings.Split(hostsUnparsed, ",")
	return hosts
}
