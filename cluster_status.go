package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	//"github.com/hashicorp/serf/client"
	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/command"
)

func DoClusterStatus(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	statusFilter := ""
	nameFilter := c.String("name")
	tagsUnparsed := c.StringSlice("tag")
	tagsRequired := map[string]string{}
	for _, t := range tagsUnparsed {
		vals := strings.Split(t, "=")
		if len(vals) != 2 {
			log.Fatalf("tag must take parameters of the format tag=value! %s is fucked up", t)
		}
		tagsRequired[vals[0]] = vals[1]
	}

	log.Printf("Checking cluster status (tags: %v, name: %q)", tagsRequired, nameFilter)

	rpcclient, err := command.RPCClient(rpcAddress, rpcAuthKey)
	if err != nil {
		log.Fatalf("Unable to connect to RPC at %s: %s", rpcAddress, err)
	}
	defer rpcclient.Close()
	members, err := rpcclient.MembersFiltered(tagsRequired, statusFilter, nameFilter)
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
