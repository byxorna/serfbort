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

type AgentEventHandler struct {
	serf    *serf.Serf
	eventCh chan serf.Event
}

func (a AgentEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		log.Print("[QUERY] received a query")
		query := e.(*serf.Query)
		//TODO track this query!
		log.Printf("[QUERY] received a query %v", query)
		payload, err := decodeMessagePayload(query.Payload)
		if err != nil {
			log.Printf("[ERROR] unable to decode payload: %s", err)
			return
		}
		log.Printf("[QUERY] parsed message: %s", payload)
		err = query.Respond([]byte("Hey this is a response"))
		if err != nil {
			log.Printf("[ERROR] unable to respond to query: %s", err)
			return
		}

	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[TESTING] %v", ue)
		switch ue.Name {
		case "deploy":
			log.Printf("[DEPLOY] received payload %q (coalescable: %t)", ue.Payload, ue.Coalesce)
			messagePayload, err := decodeMessagePayload(ue.Payload)
			if err != nil {
				log.Printf("[ERROR] unable to decode payload: %s", err)
				return
			}
			log.Printf("[DEPLOY] parsed deploy message: %s", messagePayload)

			target, ok := config.Targets[messagePayload.Target]
			if !ok {
				log.Printf("[ERROR] No target configured named %q", messagePayload.Target)
				return
			}

			log.Printf("[DEPLOY] target %s with message %q target %s", messagePayload.Target, messagePayload, target)
			//TODO FIXME do something here...
		case "verify":
			//TODO implement me
		default:
			log.Printf("[WARN] unknown message received: %s with payload %q", ue.Name, ue.Payload)
		}
	}
}

func (a *AgentEventHandler) EventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	for {
		select {
		case e := <-a.eventCh:
			log.Printf("[INFO] agent: Received event: %s", e.String())
			a.HandleEvent(e)

		case <-serfShutdownCh:
			log.Printf("[WARN] agent: Serf shutdown detected, quitting")
			a.serf.Shutdown()
			return
		default:
			// no work to do here!
			//log.Printf("no work to do...")
		}
	}
}
