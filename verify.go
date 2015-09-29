package main

import (
	"log"
	"time"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/client"
)

var (
	CHANNEL_BUFFER = 20 // how deep the channel buffer for Ack and Response channels should be (TODO FIXME)
)

func DoVerify(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	args := c.Args()
	if len(args) < 1 {
		log.Fatalf("%s requires a verify target", c.Command.Name)
	}
	target := args[0]
	args = args[1:]

	_, ok := config.Targets[target]
	if !ok {
		log.Fatalf("Unknown verify target %q (check your config)", target)
	}

	// arg is like the version of the target to verify. Allow it to be empty
	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	filterNodes := parseHostArgs(c.String("hosts")) // filter the query for only nodes matching these

	//filter query for only tags matching these
	filterTags, err := parseTagArgs(c.StringSlice("tag"))
	if err != nil {
		log.Fatal(err)
	}

	messagePayload, err := encodeMessagePayload(MessagePayload{
		Target:   target,
		Argument: arg,
	})
	if err != nil {
		log.Fatalf("Unable to encode payload: %s", err)
	}

	rpcConfig := client.Config{Addr: rpcAddress, AuthKey: rpcAuthKey}
	rpcClient, err := client.ClientFromConfig(&rpcConfig)
	if err != nil {
		log.Fatalf("Unable to connect to RPC at %s: %s", rpcAddress, err)
	}
	defer rpcClient.Close()

	log.Printf("Verifying %s with payload %q", target, messagePayload)

	ackCh := make(chan<- string, CHANNEL_BUFFER)
	respCh := make(chan<- client.NodeResponse, CHANNEL_BUFFER)
	q := client.QueryParam{
		FilterNodes: filterNodes,
		FilterTags:  filterTags,
		RequestAck:  true,
		Timeout:     60 * time.Second,
		Name:        "verify:" + target,
		Payload:     messagePayload,
		AckCh:       ackCh,
		RespCh:      respCh,
	}
	err = rpcClient.Query(&q)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("TOOD need to stream responses! get # of nodes reporting in")
}
