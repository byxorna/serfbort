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

// DoQuery will perform a query named c.Command.Name with provided parameters, and await responses.
func DoQuery(action string) func(c *cli.Context) {
	// TODO(gabe) FIXME(gabe) this is annoying. For some reason, the cli.Context.Command provided by cli is
	// empty, so we cannot extract the c.Command.Name given a context. Due to this, we need to parameterize
	// this function with the action when we setup the cli actions, so each query will know what its action
	// is. Ideally, c.Command.Name is populated to be the command name (i.e. "verify") so we can get rid of
	// this crazy closure.
	return func(c *cli.Context) {
		rpcAddress := c.GlobalString("rpc")
		rpcAuthKey := c.GlobalString("rpc-auth")
		timeout := 10 * time.Second //TODO(gabe) make this configurable
		args := c.Args()
		if len(args) < 1 {
			fmt.Printf("%s requires a target\n", action)
			os.Exit(1)
		}
		target := args[0]
		args = args[1:]

		if action == "" {
			panic(fmt.Sprintf("FUCK command name is empty: %+v %s", c.App, target))
		}

		_, ok := config.Targets[target]
		if !ok {
			fmt.Printf("Unknown %s target %q (check your config)\n", action, target)
			os.Exit(1)
		}

		// arg something we allow the action script to be parameterized by. Allow it to be empty
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
		//fmt.Printf("Got filternodes %v and filtertags %v\n", filterNodes, filterTags)

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

		ackCh := make(chan string, CHANNEL_BUFFER)
		respCh := make(chan client.NodeResponse, CHANNEL_BUFFER)
		q := client.QueryParam{
			RequestAck: true,
			Timeout:    timeout,
			Name:       action,
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
		//fmt.Printf("%d members reporting in (filtered from %d)\n", len(expectedClusterMembers), len(clusterMembers))
		fmt.Printf("%s %s %s on %d/%d hosts matching %v %v\n", action, message.Target,
			message.Argument, len(expectedClusterMembers), len(clusterMembers), filterTags, filterNodes)

		// track our incoming acks and responses
		acks := []string{}
		resps := []client.NodeResponse{}
		errorResponses := []QueryResponse{}
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
						errorResponses = append(errorResponses, queryResponse)
					}
					fmt.Printf("%s %s says %q\n", status, resp.From, output)
					resps = append(resps, resp)
				}
			default: //chill out, squire! no messages
			}
		}
		var ackPct float64 = float64(len(acks)) / float64(len(expectedClusterMembers)) * 100.0
		var respPct float64 = float64(len(resps)) / float64(len(expectedClusterMembers)) * 100.0
		var errPct float64 = float64(len(errorResponses)) / float64(len(expectedClusterMembers)) * 100.0
		//TODO this should show the actual duration of the query, not the timeout...
		fmt.Printf("%d/%d (%.1f%%) ACKed and %d/%d (%.1f%%) responded in %s\n",
			len(acks), len(expectedClusterMembers), ackPct,
			len(resps), len(expectedClusterMembers), respPct,
			q.Timeout.String())
		fmt.Printf("%d/%d (%.1f%%) reported errors\n",
			len(errorResponses), len(expectedClusterMembers), errPct)
		if len(errorResponses) > 0 {
			os.Exit(2)
		}

	}

}
