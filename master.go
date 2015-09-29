package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

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
	//TODO pull this crap into the Master type
	logOutput := os.Stderr
	logWriter := agent.NewLogWriter(123)
	a, err := agent.Create(agent.DefaultConfig(), serfConfig, logWriter)
	if err != nil {
		log.Fatalf("Unable to create agent: %s", err)
	}

	//register event handlers
	meh := MasterEventHandler{
	//Agent: a,
	}
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

type MasterEventHandler struct {
	//Agent   *agent.Agent //TODO is this necessary? how do we rebroadcast a query that comes in over RPC?
	Queries []*serf.Query
	sync.Mutex
}

func (m *MasterEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		query := e.(*serf.Query)
		log.Printf("[EVENT] %s", query)
		m.Lock()
		defer m.Unlock()
		m.Queries = append(m.Queries, query)
		//TODO broadcast this query? it came in over RPC...
		log.Print("TODO broadcast this query!")
		//(name string, payload []byte, params *QueryParam) (*QueryResponse, error)
		//TODO what query params? (filternodes, filtertags, requestack, timeout)
		/* dont do this; it causes infinite queries to be handled by the master
		_, err := m.Agent.Serf().Query(query.Name, query.Payload, nil)
		if err != nil {
			log.Printf("[ERROR] unable to rebroadcast query: %s", err)
			return
		}
		*/
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[EVENT] user event %s with payload %q (coalescable: %t)", ue.Name, ue.Payload, ue.Coalesce)
	default:
		log.Printf("[EVENT] %s", e)
	}
}
