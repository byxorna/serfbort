package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/client"
)

func DoClusterStatus(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	filterNodes := parseHostArgs(c.String("hosts"))
	tagFilter, err := parseTagArgs(c.StringSlice("tag"))

	exitcode := 0
	defer func(i int) { os.Exit(i) }(exitcode)

	if err != nil {
		fmt.Println(err)
		exitcode = 1
		return
	}

	rpcConfig := client.Config{Addr: rpcAddress, AuthKey: rpcAuthKey}
	rpcClient, err := client.ClientFromConfig(&rpcConfig)
	if err != nil {
		fmt.Printf("Unable to connect to RPC at %s: %s\n", rpcAddress, err)
		exitcode = 1
		return
	}
	defer rpcClient.Close()
	members, err := rpcClient.MembersFiltered(tagFilter, "", "")
	if err != nil {
		fmt.Printf("Error retrieving members: %s\n", err)
		exitcode = 1
		return
	}

	// filter the list of members returned for only those matching our -hosts filter if present
	filteredMembers := FilterMembers(members, filterNodes)

	fmt.Printf("%d/%d hosts matching %v %v\n", len(filteredMembers), len(members), tagFilter, filterNodes)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "Name\tAddr\tTags\tStatus\n")
	for _, member := range filteredMembers {
		fmt.Fprintf(w, "%s\t%s:%d\t%v\t%s\n", member.Name, member.Addr, member.Port, member.Tags, member.Status)
	}
	w.Flush()

}
