package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/command"
)

func DoDeploy(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	cmd := "deploy"
	args := c.Args()
	if len(args) < 1 {
		log.Fatalf("%s requires a deploy target", c.Command.Name)
	}
	target := args[0]
	args = args[1:]

	_, ok := config.Targets[target]
	if !ok {
		log.Fatalf("Unable to find target %q in the config", target)
	}

	var arg string
	if len(args) > 0 {
		arg = args[0]
	}
	messagePayload, err := encodeDeployMessage(DeployMessage{
		Target:       target,
		RequiredTags: map[string]string{},
		Argument:     arg,
	})
	if err != nil {
		log.Fatalf("Unable to encode payload: %s", err)
	}

	log.Printf("Deploying %s with payload %q", target, messagePayload)

	rpcclient, err := command.RPCClient(rpcAddress, rpcAuthKey)
	if err != nil {
		log.Fatalf("Unable to connect to RPC at %s: %s", rpcAddress, err)
	}
	defer rpcclient.Close()

	err = rpcclient.UserEvent(cmd, messagePayload, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("OK")

}
