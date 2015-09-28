package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	//"github.com/hashicorp/serf/client"
	"github.com/hashicorp/serf/command"
	"github.com/hashicorp/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

var (
	c                          = serf.DefaultConfig()
	mode           instaceMode = SlaveMode
	masterAddress  string
	listenAddress  string // this is the interface that serf runs on
	rpcAddress     string // this is the interface that serf's RPC runs on (localhost:7373)
	defaultName, _ = os.Hostname()
	rpcAuthKey     = ""
)

type instaceMode int

const (
	MasterMode instaceMode = iota
	SlaveMode
	DeployMode
)

func init() {
	var name string
	var tagsFile string
	flag.StringVar(&masterAddress, "master", "localhost:7946", "Join the cluster by coordinating with this master")
	flag.StringVar(&listenAddress, "listen", "localhost:7946", "Listen on the address for serf communication")
	flag.StringVar(&rpcAddress, "rpc", "localhost:7373", "RPC address of the serfbort master")
	flag.StringVar(&name, "name", defaultName, "Name to use in serf protocol")
	flag.StringVar(&tagsFile, "tags-file", "", "Load tags for agent from file (json format)")
	flag.Parse()

	if tagsFile != "" {
		tags, err := loadTagsFromFile(tagsFile)
		if err != nil {
			log.Fatalf("Unable to load tags from file: %s", err)
		}
		c.Tags = tags
	}
	c.NodeName = name

}

func main() {
	fields := strings.Split(listenAddress, ":")
	if len(fields) != 2 {
		log.Fatalf("-listen requires host:port! %s is not valid", listenAddress)
	}
	bindAddr := fields[0]
	bindPort, err := strconv.Atoi(fields[1])
	if err != nil {
		log.Fatalf("Unable to parse %s into port", fields[1])
	}
	c.MemberlistConfig.BindAddr = bindAddr
	c.MemberlistConfig.BindPort = bindPort
	//TODO advertiseaddr and advertiseport
	//c.MemberlistConfig.AdvertiseAddr = x
	//c.MemberlistConfig.AdvertisePort = y

	evtCh := make(chan serf.Event, 64)
	c.EventCh = evtCh

	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "master":
			mode = MasterMode
		case "deploy":
			mode = DeployMode
		default:
			mode = SlaveMode
		}
	}

	switch mode {
	case SlaveMode:
		log.Printf("Starting agent with tags %v", c.Tags)
		//TODO pull this into the Agent type
		s, err := serf.Create(c)
		if err != nil {
			log.Fatalf("Error creating Serf: %s", err)
		}

		log.Printf("Joining cluster by way of %s", masterAddress)
		n, err := s.Join([]string{masterAddress}, true)
		if n > 0 {
			log.Printf("Cluster joined; %d nodes participating", n)
		}
		if err != nil {
			log.Fatalf("unable to join cluster: %s", err)
		}

		a := AgentEventHandler{s, evtCh}
		a.EventLoop()

	case MasterMode:
		log.Printf("Starting master on %s", listenAddress)
		log.Printf("Starting master RPC listener on %s", rpcAddress)
		//TODO pull this crap into the Master type
		logOutput := os.Stderr
		logWriter := agent.NewLogWriter(123)
		a, err := agent.Create(agent.DefaultConfig(), c, logWriter)
		if err != nil {
			log.Fatalf("Unable to create agent: %s", err)
		}

		//register event handlers
		a.RegisterEventHandler(MasterEventHandler{})

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

		//TODO we should create an agent with agent.Create instead of this!
		//a := agent.Agent{}
		rpcListener, err := net.Listen("tcp", rpcAddress)
		if err != nil {
			log.Fatal("Error starting RPC listener: %s", err)
		}
		agent.NewAgentIPC(a, rpcAuthKey, rpcListener, logOutput, logWriter)
		select {}

	case DeployMode:
		rpcclient, err := command.RPCClient(rpcAddress, rpcAuthKey)
		if err != nil {
			log.Fatalf("Unable to connect to master at %s: %s", rpcAddress, err)
		}

		cmd := "deploy"
		payload := []byte("fuck")
		log.Printf("Sending %s command with payload %q", cmd, payload)
		err = rpcclient.UserEvent(cmd, payload, false)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("OK")
	}

}

func leaveCluster() {
	//TODO this defer isnt running when ctrl-c'd
	/*
			go func() {
		    sig := <-sigChan
		        log.Printf("Cleaning up registered callback after signal %s\n", sig)
		                              os.Exit(0)
				log.Print("fuck")
				if s != nil {
					log.Print("Leaving cluster")
					err := s.Leave()
					if err != nil {
						log.Print(err)
						return
					}
					log.Print("Shutting down")
					err = s.Shutdown()
					if err != nil {
						log.Print(err)
						return
					}

				}
			}()
	*/
}
