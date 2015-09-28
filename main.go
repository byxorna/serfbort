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
	name           string
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
	flag.StringVar(&masterAddress, "master", "mymaster.company.net:7946", "Join the cluster by coordinating with this master")
	flag.StringVar(&listenAddress, "listen", "localhost:7946", "Listen on the address for serf communication")
	flag.StringVar(&rpcAddress, "rpc", "localhost:7373", "RPC address of the serfbort master")
	flag.StringVar(&name, "name", defaultName, "Name to use in serf protocol")
	flag.Parse()
}

func main() {
	c.Tags = map[string]string{
		"role": "web",
		"env":  "dev",
	}
	c.NodeName = name

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

	log.Print("Running...")
	switch mode {
	case SlaveMode:
		log.Print("Starting node")
		s := joinCluster()
		defer leaveCluster()
		slave := Slave{s, evtCh}
		slave.Run()
	case MasterMode:
		log.Printf("Starting master on %s", listenAddress)
		s := joinCluster()
		defer leaveCluster()
		master := Master{s, evtCh}

		log.Printf("Starting master RPC listener on %s", rpcAddress)
		//TODO we should create an agent with agent.Create instead of this!
		logOutput := os.Stdout
		logWriter := agent.NewLogWriter(123)
		a := agent.Agent{}
		rpcListener, err := net.Listen("tcp", rpcAddress)
		if err != nil {
			log.Fatal("Error starting RPC listener: %s", err)
		}
		agent.NewAgentIPC(&a, rpcAuthKey, rpcListener, logOutput, logWriter)

		master.Run()
	case DeployMode:
		/*
			clientCfg := client.Config{
				Addr:    listenAddress,
				AuthKey: "",
			} */
		//rpcclient, err := client.NewRPCClient(listenAddress)
		rpcclient, err := command.RPCClient(rpcAddress, rpcAuthKey)
		if err != nil {
			log.Fatalf("Unable to connect to master at %s: %s", rpcAddress, err)
		}

		log.Printf("Sending event to cluster...")
		err = rpcclient.UserEvent("deploy", []byte{}, false)
		if err != nil {
			log.Fatal(err)
		}

	}

}

func joinCluster() *serf.Serf {
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

	members := s.Members()
	log.Printf("%d nodes currently in cluster:", len(members))
	for _, m := range members {
		/*
		   Name   string
		   Addr   net.IP
		   Port   uint16
		   Tags   map[string]string
		   Status MemberStatus
		*/
		log.Printf("  %s %s:%d %v %s", m.Name, m.Addr, m.Port, m.Tags, m.Status)
	}
	return s

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
