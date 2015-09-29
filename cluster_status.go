package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/client"
)

func DoClusterStatus(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	statusFilter := ""
	nameFilter := c.String("name")
	tagFilter, err := parseTagArgs(c.StringSlice("tag"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Checking cluster status (tags: %v, name: %q)", tagFilter, nameFilter)

	rpcConfig := client.Config{Addr: rpcAddress, AuthKey: rpcAuthKey}
	rpcClient, err := client.ClientFromConfig(&rpcConfig)
	if err != nil {
		log.Fatalf("Unable to connect to RPC at %s: %s", rpcAddress, err)
	}
	defer rpcClient.Close()
	members, err := rpcClient.MembersFiltered(tagFilter, statusFilter, nameFilter)
	if err != nil {
		log.Fatalf("Error retrieving members: %s", err)
	}

	fmt.Printf("%d nodes reporting\n", len(members))
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "name\taddr\tport\ttags\tstatus\n")
	for _, member := range members {
		fmt.Fprintf(w, "%s\t%s\t%d\t%v\t%s\n", member.Name, member.Addr, member.Port, member.Tags, member.Status)
	}
	w.Flush()

}
