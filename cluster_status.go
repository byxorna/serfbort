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
	statusFilter, nameFilter := "", ""
	filterNodes := parseHostArgs(c.String("hosts"))
	tagFilter, err := parseTagArgs(c.StringSlice("tag"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Checking cluster status (tags: %v, hosts: %v)\n", tagFilter, filterNodes)

	rpcConfig := client.Config{Addr: rpcAddress, AuthKey: rpcAuthKey}
	rpcClient, err := client.ClientFromConfig(&rpcConfig)
	if err != nil {
		fmt.Printf("Unable to connect to RPC at %s: %s\n", rpcAddress, err)
		os.Exit(1)
	}
	defer rpcClient.Close()
	members, err := rpcClient.MembersFiltered(tagFilter, statusFilter, nameFilter)
	if err != nil {
		fmt.Printf("Error retrieving members: %s\n", err)
		os.Exit(1)
	}

	// filter the list of members returned for only those matching our -hosts filter if present
	filteredMembers := FilterMembers(members, filterNodes)

	fmt.Printf("%d/%d hosts matching %v %v\n", len(filteredMembers), len(members), tagFilter, filterNodes)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "Name\tAddr\tTags\tStatus\n")
	for _, member := range members {
		fmt.Fprintf(w, "%s\t%s:%d\t%v\t%s\n", member.Name, member.Addr, member.Port, member.Tags, member.Status)
	}
	w.Flush()

}
