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

	/*TODO using empty slices and maps causes every agent to filter these queries. WTF? omit for now...
	filterNodes := parseHostArgs(c.String("hosts")) // filter the query for only nodes matching these

	//filter query for only tags matching these
	filterTags, err := parseTagArgs(c.StringSlice("tag"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Got filternodes %v and filtertags %v", filterNodes, filterTags)
	*/

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
		//FilterNodes: filterNodes,
		//FilterTags:  filterTags,
		RequestAck: true,
		Timeout:    10 * time.Second, // let serf set this default: serf.DefaultQueryTimeout()
		Name:       "verify:" + target,
		Payload:    messageEnc,
		AckCh:      ackCh,
		RespCh:     respCh,
	}
	err = rpcClient.Query(&q)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//TODO determine how many nodes we _should_ be hearing from, so we can display percentages...
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
				if queryResponse.Status != 0 {
					status = "ERROR"
				}
				fmt.Printf("%s %s says %q\n", status, resp.From, strings.TrimSpace(queryResponse.Output))
				resps = append(resps, resp)
			}
		default: //chill out, squire! no messages
		}
	}
	fmt.Printf("Got %d acks and %d responses in %s\n", len(acks), len(resps), q.Timeout.String())

}
