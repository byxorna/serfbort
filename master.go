package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

func StartMaster(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	masterAddress := c.String("master")
	listenAddress := c.String("listen")

	fields := strings.Split(listenAddress, ":")
	if len(fields) != 2 {
		log.Fatalf("listen requires host:port! %s is not valid", listenAddress)
	}
	bindAddr := fields[0]
	bindPort, err := strconv.Atoi(fields[1])
	if err != nil {
		log.Fatalf("Unable to parse %s into port", fields[1])
	}
	serfConfig := serf.DefaultConfig()
	serfConfig.MemberlistConfig.BindAddr = bindAddr
	serfConfig.MemberlistConfig.BindPort = bindPort
	if c.IsSet("name") {
		serfConfig.NodeName = c.String("name")
	}
	serfConfig.Tags = map[string]string{"master": "true"}

	log.Printf("Starting master on %s", listenAddress)
	log.Printf("Starting master RPC listener on %s", rpcAddress)
	logOutput := os.Stderr
	logWriter := agent.NewLogWriter(123)
	a, err := agent.Create(agent.DefaultConfig(), serfConfig, logWriter)
	if err != nil {
		log.Fatalf("Unable to create agent: %s", err)
	}

	//register event handlers
	meh := MasterEventHandler{}
	a.RegisterEventHandler(&meh)

	if err := a.Start(); err != nil {
		log.Fatalf("Unable to start agent: %s", err)
	}

	log.Printf("Joining cluster by way of %s", masterAddress)
	n, err := a.Join([]string{masterAddress}, true)
	if n > 0 {
		log.Printf("Cluster joined; %d nodes participating", n)
	}
	if err != nil {
		log.Fatalf("unable to join cluster: %s", err)
	}

	members := a.Serf().Members()
	log.Printf("%d nodes currently in cluster:", len(members))
	for _, m := range members {
		log.Printf("  %s %s:%d %v %s", m.Name, m.Addr, m.Port, m.Tags, m.Status)
	}

	rpcListener, err := net.Listen("tcp", rpcAddress)
	if err != nil {
		log.Fatalf("Error starting RPC listener: %s", err)
	}
	//TODO should we listen for shutdown signals and close the agent properly?
	agent.NewAgentIPC(a, rpcAuthKey, rpcListener, logOutput, logWriter)
	select {}

}

type MasterEventHandler struct{}

func (m *MasterEventHandler) HandleEvent(e serf.Event) {
	switch evt := e.(type) {
	case *serf.Query:
		log.Printf("%s: payload %q", evt.EventType(), evt.Payload)
		/* we dont need to keep track of queries on the master...
		m.Lock()
		defer m.Unlock()
		m.Queries = append(m.Queries, evt)
		*/
	case serf.UserEvent:
		log.Printf("%s: %s with payload %q (coalescable: %t)", evt.EventType(), evt.Name, evt.Payload, evt.Coalesce)
	case serf.MemberEvent:
		log.Printf("%s: %v", evt.EventType(), evt.Members)
	default:
		log.Printf("[EVENT] %s", evt.EventType())
	}
}
