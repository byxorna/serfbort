package main

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/serf/client"
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
	hosts := []string{}
	if hostsUnparsed != "" {
		hosts = strings.Split(hostsUnparsed, ",")
	}
	return hosts
}

// converts a given encryption key (EncryptKey in serf parlance) to []byte
// must be 16 characters (limitation imposed by serf)
// head -c16 /dev/urandom | base64
func keyToBytes(key string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(key)
}

// FilterMembers takes a list of serf members and selects those that have a name in the
// given list of names. Serf doesn't provide this functionality for client.MembersFiltered
// so we implement it ourselves after the query is returned. If no names are provided,
// return the list of members as is; no filtering takes place.
func FilterMembers(members []client.Member, filterNames []string) []client.Member {
	if len(filterNames) == 0 {
		return members
	}
	filteredMembers := []client.Member{}
	sort.Sort(SortableStringSlice(filterNames))
	for _, m := range members {
		fmt.Println(m)
		if Contains(filterNames, m.Name) {
			fmt.Println(m.Name + " is in list of names")
			filteredMembers = append(filteredMembers, m)
		}
	}
	return filteredMembers
}

// Contains checks to see if a string provided is in the list of strings
func Contains(l []string, val string) bool {
	i := sort.Search(len(l), func(i int) bool { return l[i] >= val })
	if i < len(l) && l[i] == val {
		return true
	} else {
		return false
	}
}

type SortableStringSlice []string

func (a SortableStringSlice) Len() int           { return len(a) }
func (a SortableStringSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortableStringSlice) Less(i, j int) bool { return a[i] < a[j] }
