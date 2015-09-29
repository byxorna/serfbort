package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	//"github.com/hashicorp/serf/client"
	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/command"
	"github.com/hashicorp/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

var (
	defaultName, _ = os.Hostname()
	config         Config

	// these variables are set via -ldflags="-X main.myvar=fuck" in makefile
	version string
	branch  string
	commit  string
)

func main() {

	app := cli.NewApp()
	app.Name = "serfbort"
	app.Usage = "deploy tool"
	app.Version = fmt.Sprintf("%s (branch: %s commit: %s)", version, branch, commit)
	app.Action = cli.ShowAppHelp
	//top level flags, common to all commands
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			Usage:  "JSON config describing deploy targets, commands, etc",
			EnvVar: "SERFBORT_CONFIG",
		},
		cli.StringFlag{
			Name:  "rpc",
			Value: "localhost:7373",
			Usage: "Listen on the address for RPC requests (i.e. from deploy command)",
		},
		cli.StringFlag{
			Name:  "rpc-auth",
			Usage: "Auth token to use for RPC",
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
			Before: LoadConfig,
			Action: StartAgent,
		},

		{
			Name:   "deploy",
			Flags:  []cli.Flag{},
			Usage:  "Perform a deploy to a target",
			Action: DoDeploy,
			Before: LoadConfig,
		},
		{
			Name:   "verify",
			Flags:  []cli.Flag{},
			Usage:  "Verify a deploy target",
			Action: DoVerify,
		},

		{
			Name: "cluster-status",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "tag",
					Usage: `Filter by requiring tag on agent (tag=value) (can be a regexp like "val.*", and passed multiple times)`,
				},
				cli.StringFlag{
					Name:  "name",
					Usage: `Filter by requiring name of agent to match (can be a regexp like "web-1.*")`,
				},
			},
			Usage:  "Check the status of all cluster members",
			Action: DoClusterStatus,
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
		log.Fatalf("Error starting RPC listener: %s", err)
	}
	agent.NewAgentIPC(a, rpcAuthKey, rpcListener, logOutput, logWriter)
	select {}

}

func DoVerify(c *cli.Context) {
	panic("fuck implement me")
}

func DoClusterStatus(c *cli.Context) {
	rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
	statusFilter := ""
	nameFilter := c.String("name")
	tagsUnparsed := c.StringSlice("tag")
	tagsRequired := map[string]string{}
	for _, t := range tagsUnparsed {
		vals := strings.Split(t, "=")
		if len(vals) != 2 {
			log.Fatalf("tag must take parameters of the format tag=value! %s is fucked up", t)
		}
		tagsRequired[vals[0]] = vals[1]
	}

	log.Printf("Checking cluster status (tags: %v, name: %q)", tagsRequired, nameFilter)

	rpcclient, err := command.RPCClient(rpcAddress, rpcAuthKey)
	if err != nil {
		log.Fatalf("Unable to connect to RPC at %s: %s", rpcAddress, err)
	}
	defer rpcclient.Close()
	members, err := rpcclient.MembersFiltered(tagsRequired, statusFilter, nameFilter)
	if err != nil {
		log.Fatalf("Error retrieving members: %s", err)
	}

	fmt.Printf("%d nodes reporting\n", len(members))
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "name\taddr\tport\ttags\tstatus\n")
	for _, member := range members {
		fmt.Fprintf(w, "%s\t%s\t%d\t%v\t%s\n", member.Name, member.Addr, member.Port, member.Tags, member.Status)
	}
	w.Flush()

}

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

// loads the config into a global variable
func LoadConfig(c *cli.Context) error {
	if !c.GlobalIsSet("config") {
		log.Fatalln("config is required")
	}
	configFile := c.GlobalString("config")
	cfg, err := loadConfigFromFile(configFile)
	if err != nil {
		//TODO this is dumb that returning an error just causes the command to not be run
		// it doesnt actually print any messages out
		log.Fatalf("Unable to load config: %s", err)
	}
	config = cfg
	log.Printf("Loaded %d targets from %s: %v", len(config.Targets), configFile, config.Targets)

	return err
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
