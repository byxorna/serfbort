package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/serf"
)

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