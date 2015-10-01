package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/byxorna/serfbort/cmd"
	"github.com/codegangsta/cli"
	"github.com/hashicorp/serf/serf"
)

func StartAgent(c *cli.Context) {
	//rpcAddress := c.GlobalString("rpc")
	rpcAuthKey := c.GlobalString("rpc-auth")
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
	if c.GlobalIsSet("rpc-auth") {
		keyBytes, err := keyToBytes(rpcAuthKey)
		if err != nil {
			log.Fatalf("Invalid encryption key: %s", err)
		}
		// this needs to be 16, 24, or 32 bytes long. TODO validate this?
		serfConfig.MemberlistConfig.SecretKey = keyBytes
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
		query := e.(*serf.Query)
		log.Printf("[QUERY] received a query %v", query)
		go handleQuery(query)

	case serf.EventUser:
		//TODO(gabe) GUT THIS! i dont think its strictly necessary anymore... everything should be a query, right?
		ue := e.(serf.UserEvent)
		switch ue.Name {
		case "deploy":
			log.Printf("[DEPLOY] received payload %q (coalescable: %t)", ue.Payload, ue.Coalesce)
			message, err := DecodeMessagePayload(ue.Payload)
			if err != nil {
				log.Printf("[ERROR] unable to decode payload: %s", err)
				return
			}
			log.Printf("[DEPLOY] parsed deploy message: %s", message)

			target, ok := config.Targets[message.Target]
			if !ok {
				log.Printf("[ERROR] No target configured named %q", message.Target)
				return
			}

			log.Printf("[DEPLOY] target %s with message %q target %s", message.Target, message, target)
			//TODO FIXME do something here...

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

// TODO this shouldnt log, nor should it swallow errors, right?
// this needs a refactor. - gabe
func respondToQuery(query *serf.Query, resp QueryResponse) {
	respEnc, err := resp.Encode()
	if err != nil {
		log.Printf("[ERROR] unable to encode query response %v: %s", resp, err)
		return
	}
	err = query.Respond(respEnc)
	if err != nil {
		log.Printf("[ERROR] unable to respond to query: %s", err)
		return
	}

}

// HandleQuery will, given a query, decode its payload and respond to it as appropriate.
// Ideally should be run in a goroutine in an event loop, so it does not return anything
// of value.
func handleQuery(query *serf.Query) {
	// we always wanna send a response to the query; failed or not
	resp := QueryResponse{
		Output: "",
		Err:    nil,
		Status: 0,
	}
	// this makes more sense than encoding duplicate information in the payload of the query.
	action := query.Name
	message, err := DecodeMessagePayload(query.Payload)
	if err != nil {
		log.Printf("[ERROR] unable to decode payload: %s", err)
		errMessage := err.Error()
		resp.Err = &errMessage
		respondToQuery(query, resp)
		return
	}

	log.Printf("[QUERY] parsed message: %s", message)
	//TODO ensure the output is truncated so we dont exceed UDP datagram size? will serf do this
	// for us? or should we be more clever?
	target, ok := config.Targets[message.Target]
	if !ok {
		log.Printf("[ERROR] No target configured named %q", message.Target)
		errMessage := fmt.Sprintf("No target named %q found in config", message.Target)
		resp.Err = &errMessage
		respondToQuery(query, resp)
		return
	}

	//TODO this should check what the action is and use the script appropriate
	scriptTemplate := ""
	switch action {
	case "deploy":
		scriptTemplate = target.DeployScript
	case "verify":
		scriptTemplate = target.VerifyScript
	default:
		log.Printf("[ERROR] Unknown action %q! Unable to run script.", action)
		errMessage := fmt.Sprintf("Unknown action %q for %q", action, message.Target)
		resp.Err = &errMessage
		respondToQuery(query, resp)
		return
	}
	if scriptTemplate == "" {
		log.Printf("[ERROR] No script configured to %s %s! Not running script.", action, message.Target)
		errMessage := fmt.Sprintf("No script configured to %s %s", action, message.Target)
		resp.Err = &errMessage
		respondToQuery(query, resp)
		return
	}
	runner := cmd.New(message.Target, scriptTemplate, message.Argument)
	outputBuf, err := runner.Run()
	if err != nil {
		log.Printf("[ERROR] Error running %s %s script: %s", message.Target, action, err)
		errMessage := err.Error()
		resp.Err = &errMessage
		resp.Output = outputBuf.String()
		respondToQuery(query, resp)
		return
	}

	//TODO read only enough bytes to fill the response? should this return io.Reader instead of the buffer?
	resp.Output = outputBuf.String()
	respondToQuery(query, resp)
}
