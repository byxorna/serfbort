package main

import (
	"fmt"
	"log"
	"os"

	//"github.com/hashicorp/serf/client"
	"github.com/codegangsta/cli"
)

var (
	defaultName, _ = os.Hostname()
	config         Config

	// these variables are set via -ldflags="-X main.myvar=fuck" in makefile
	version string = "???"
	branch  string = "???"
	commit  string = "???"
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
			Name: "deploy",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "hosts",
					Usage: "Only check status of these hosts (comma separated list)",
				},
				cli.StringSliceFlag{
					Name:  "tag",
					Usage: `Restrict deploy by requiring tag on host (tag=value) (can be a regexp like "val.*", and passed multiple times)`,
				},
			},
			Usage:  "Perform a deploy to a target",
			Action: DoQuery,
			Before: LoadConfig,
		},
		{
			Name: "verify",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "hosts",
					Usage: "Only verify these hosts (comma separated list)",
				},
				cli.StringSliceFlag{
					Name:  "tag",
					Usage: `Restrict verify by requiring tag on host (tag=value) (can be a regexp like "val.*", and passed multiple times)`,
				},
			},
			Usage:  "Verify a deploy target",
			Action: DoQuery,
			Before: LoadConfig,
		},

		{
			Name: "cluster-status",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "hosts",
					Usage: "Only check status of these hosts (comma separated list)",
				},
				cli.StringSliceFlag{
					Name:  "tag",
					Usage: `Only show status of hosts with tag (tag=value) (can be a regexp like "val.*", and passed multiple times)`,
				},
			},
			Usage:  "Check the status of all cluster members",
			Action: DoClusterStatus,
		},
	}

	app.Run(os.Args)

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
	//log.Printf("Loaded %d targets from %s: %v", len(config.Targets), configFile, config.Targets)

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
