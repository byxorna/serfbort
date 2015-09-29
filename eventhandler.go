package main

import (
	"log"

	"github.com/hashicorp/serf/serf"
)

type MasterEventHandler struct{}

func (m MasterEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		log.Printf("[EVENT] query: %s", e)
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[EVENT] user event %s with payload %q (coalescable: %t)", ue.Name, ue.Payload, ue.Coalesce)
	default:
		log.Printf("[EVENT] %s", e)
	}
}

type AgentEventHandler struct {
	serf    *serf.Serf
	eventCh chan serf.Event
}

func (a AgentEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		//TODO
		log.Printf("[QUERY] implement me")
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[TESTING] %v", ue)
		switch ue.Name {
		case "deploy":
			log.Printf("[DEPLOY] received payload %q (coalescable: %t)", ue.Payload, ue.Coalesce)
			deployMessage, err := decodeDeployMessage(ue.Payload)
			if err != nil {
				//TODO this should probably be a user query instead of a event...
				log.Printf("[ERROR] unable to decode payload: %s", err)
				return
			}
			log.Printf("[DEPLOY] parsed deploy message: %s", deployMessage)

			target, ok := config.Targets[deployMessage.Target]
			if !ok {
				log.Printf("[ERROR] No target configured named %q", deployMessage.Target)
				return
			}

			log.Printf("[DEPLOY] target %s with message %q target %s", deployMessage.Target, deployMessage, target)
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

		}
	}
}
