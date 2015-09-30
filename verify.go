package main

import (
	"fmt"
	"os"
	"strings"
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
		fmt.Printf("%s requires a verify target\n", c.Command.Name)
		os.Exit(1)
	}
	target := args[0]
	args = args[1:]

	_, ok := config.Targets[target]
	if !ok {
		fmt.Printf("Unknown verify target %q (check your config)\n", target)
		os.Exit(1)
	}

	// arg is like the version of the target to verify. Allow it to be empty
	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	filterNodes := parseHostArgs(c.String("hosts")) // filter the query for only nodes matching these
	filterTags, err := parseTagArgs(c.StringSlice("tag"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//filter query for only tags matching these
	fmt.Printf("Got filternodes %v and filtertags %v\n", filterNodes, filterTags)

	message := MessagePayload{
		Target:   target,
		Argument: arg,
	}
	messageEnc, err := message.Encode()
	if err != nil {
		fmt.Printf("Unable to encode payload: %s\n", err)
		os.Exit(1)
	}

	rpcConfig := client.Config{Addr: rpcAddress, AuthKey: rpcAuthKey}
	rpcClient, err := client.ClientFromConfig(&rpcConfig)
	if err != nil {
		fmt.Printf("Unable to connect to RPC at %s: %s\n", rpcAddress, err)
		os.Exit(1)
	}
	defer rpcClient.Close()

	fmt.Printf("Verifying %s with payload %q\n", target, message)

	ackCh := make(chan string, CHANNEL_BUFFER)
	respCh := make(chan client.NodeResponse, CHANNEL_BUFFER)
	q := client.QueryParam{
		RequestAck: true,
		Timeout:    10 * time.Second, // let serf set this default: serf.DefaultQueryTimeout()
		Name:       "verify",
		Payload:    messageEnc,
		AckCh:      ackCh,
		RespCh:     respCh,
	}
	// if we blindly set an empty map/slice here, we wont get any responses from agents :(
	// so, we only set them if there is anything of value in them
	if len(filterTags) > 0 {
		q.FilterTags = filterTags
	}
	if len(filterNodes) > 0 {
		q.FilterNodes = filterNodes
	}
	err = rpcClient.Query(&q)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//TODO determine how many nodes we _should_ be hearing from, so we can display percentages...
	//TODO should this list come from collins/a source of truth? Maybe not...
	clusterMembers, err := rpcClient.MembersFiltered(filterTags, "", "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// filter by node name manually, cause MembersFiltered doesnt take a list of hosts (only a name regexp)
	expectedClusterMembers := FilterMembers(clusterMembers, filterNodes)
	fmt.Printf("%d members reporting in (filtered from %d)\n", len(expectedClusterMembers), len(clusterMembers))

	// track our incoming acks and responses
	acks := []string{}
	resps := []client.NodeResponse{}
	pendingAcks, pendingResponses := true, true
	for pendingAcks || pendingResponses {
		select {
		case ack, open := <-ackCh:
			if !open {
				// channel was closed, so lets get outta here
				fmt.Println("Ack channel is closed!")
				pendingAcks = false
			} else {
				acks = append(acks, ack)
			}
		case resp, open := <-respCh:
			if !open {
				fmt.Println("Response channel is closed")
				pendingResponses = false
			} else {
				queryResponse, err := DecodeQueryResponse(resp.Payload)
				if err != nil {
					fmt.Printf("Unable to decode response from %s: %s\n", resp.From, err)
					continue
				}
				status := "OK   "
				output := strings.TrimSpace(queryResponse.Output)
				if queryResponse.Err != nil {
					status = "ERROR"
					output = queryResponse.Err.Error()
				}
				fmt.Printf("%s %s says %q\n", status, resp.From, output)
				resps = append(resps, resp)
			}
		default: //chill out, squire! no messages
		}
	}
	fmt.Printf("Got %d acks and %d responses in %s\n", len(acks), len(resps), q.Timeout.String())

}
