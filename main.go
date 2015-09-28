package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	//"github.com/hashicorp/serf/client"
	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/command"
	"github.com/hashicorp/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

var (
	defaultName, _ = os.Hostname()
)

func main() {

	app := cli.NewApp()
	app.Name = "serfbort"
	app.Usage = "deploy tool"
	app.Action = cli.ShowAppHelp
	//top level flags, common to all commands
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rpc",
			Value: "localhost:7373",
			Usage: "Listen on the address for RPC requests (i.e. from deploy command)",
		},
		cli.StringFlag{
			Name:  "rpc-auth",
			Usage: "Auth token to use for RPC",
		},
		cli.StringFlag{
			Name:   "config",
			Usage:  "JSON config describing deploy targets, commands, etc",
			EnvVar: "SERFBORT_CONFIG",
		},
	}
	// subcommands
	app.Commands = []cli.Command{
		{
			Name: "master",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Value: defaultName,
					Usage: "Name to use in serf protocol",
				},
				cli.StringFlag{
					Name:  "master",
					Value: "localhost:7946",
					Usage: "Join the cluster by coordinating with this master",
				},
				cli.StringFlag{
					Name:  "listen",
					Value: "localhost:7946",
					Usage: "Listen on the address for serf communication",
				},
			},
			Usage:  "Run the serfbort master process",
			Action: StartMaster,
		},
		{
			Name: "deploy",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verify, v",
					Usage: "Run verification of the deploy",
				},
			},
			//TODO add subcommands dynamically based on a config?
			Subcommands: []cli.Command{},
			Usage:       "Perform a deploy",
			Action:      DoDeploy,
		},
		{
			Name: "agent",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "master",
					Value: "localhost:7946",
					Usage: "Join the cluster by coordinating with this master",
				},
				cli.StringFlag{
					Name:  "listen",
					Value: "localhost:7947",
					Usage: "Listen on the address for serf communication",
				},
				cli.StringFlag{
					Name:  "name",
					Value: defaultName,
					Usage: "Name to use in serf protocol",
				},
				cli.StringFlag{
					Name:  "tags-file",
					Usage: "Load tags for agent from file (json format)",
				},
			},
			Usage:  "Run the serfbort agent",
			Action: StartAgent,
		},
	}

	app.Run(os.Args)

}

func StartAgent(c *cli.Context) {
	//TODO make this work with authkeys
	//rpcAddress := c.GlobalString("rpc")
	//rpcAuthKey := c.GlobalString("rpc-auth")
	masterAddress := c.String("master")
	listenAddress := c.String("listen")

	serfConfig := serf.DefaultConfig()
	fields := strings.Split(listenAddress, ":")
	if len(fields) != 2 {
		log.Fatalf("listen requires host:port! %s is not valid", listenAddress)
	}
	bindAddr := fields[0]
	bindPort, err := strconv.Atoi(fields[1])
	if err != nil {
		log.Fatalf("Unable to parse %s into port", fields[1])
	}
	serfConfig.MemberlistConfig.BindAddr = bindAddr
	serfConfig.MemberlistConfig.BindPort = bindPort

	evtCh := make(chan serf.Event, 64)
	serfConfig.EventCh = evtCh

	if c.IsSet("tags-file") {
		tagsLoaded, err := loadTagsFromFile(c.String("tags-file"))
		if err != nil {
			log.Fatalf("Unable to load tags from file: %s", err)
		}
		serfConfig.Tags = tagsLoaded
	}
	if c.IsSet("name") {
		serfConfig.NodeName = c.String("name")
	}
	log.Printf("Starting agent with tags %v", serfConfig.Tags)
	//TODO pull this into the Agent type
	s, err := serf.Create(serfConfig)
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
}

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

}
func DoDeploy(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	cmd := "deploy"
	payload := []byte("fuck")

	rpcclient, err := command.RPCClient(rpcAddress, rpcAuthKey)
	if err != nil {
		log.Fatalf("Unable to connect to master at %s: %s", rpcAddress, err)
	}

	log.Printf("Sending %s command with payload %q", cmd, payload)
	err = rpcclient.UserEvent(cmd, payload, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("OK")

}

/*
func leaveCluster() {
	//TODO this defer isnt running when ctrl-c'd
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
}
*/
